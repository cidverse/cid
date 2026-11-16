package egress

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/cidverse/cid/pkg/egress/firewall"
	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/rs/zerolog/log"
)

// RunCommand starts the egress filter, runs the command once the filter is active and stops the filter when the
// command exits. Returns the command's exit code (130 if interrupted before the command ran).
//
// SIGINT / SIGTERM are forwarded to the command and the filter is only removed after the command exited, so the
// command never runs unfiltered. If the filter can't be started the command is not run, unless failOpen is set (audit
// mode: report, never break).
//
// If the firewall rules are modified while the command runs (block mode), the command and all its descendants are
// killed and RunCommand returns 1. In audit mode the modification is only reported.
func RunCommand(cfg Config, command []string, extraEnv []string, failOpen bool) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	child := &childProcess{command: command, env: extraEnv}
	stopForwarding := child.forwardSignals(cancel)
	defer stopForwarding()

	var tampered atomic.Bool
	onTamper := cfg.OnTamper
	cfg.OnTamper = func(t firewall.Tamper) {
		if onTamper != nil {
			onTamper(t)
		}
		if cfg.Mode != policy.ModeBlock {
			log.Warn().Msg("egress filter: firewall tampering detected, the command keeps running (audit mode)")
			return
		}
		tampered.Store(true)
		log.Error().Strs("command", command).Msg("egress filter: firewall tampering detected, terminating the command and all its processes")
		child.kill()
	}

	result := make(chan int, 1)
	started := false // only accessed from the goroutine calling Run (OnReady runs synchronously inside Run)
	cfg.OnReady = func() {
		started = true
		child.env = append(child.env, caBundleEnv(cfg.CABundle, os.Getenv)...)
		if err := child.start(); err != nil {
			result <- 1
			cancel()
			return
		}
		go func() {
			result <- child.wait()
			cancel() // stop the filter only after the command exited
		}()
	}

	if err := Run(ctx, cfg); err != nil {
		if !started {
			if failOpen && !child.isInterrupted() {
				log.Warn().Err(err).Msg("egress filter could not be started, running without it (audit mode)")
				if err := child.start(); err != nil {
					return 1
				}
				return child.wait()
			}
			log.Error().Err(err).Msg("egress filter could not be started, command not executed")
			return 1
		}
		log.Error().Err(err).Msg("egress filter cleanup failed")
	}

	select {
	case code := <-result:
		if tampered.Load() {
			log.Error().Int("exit_code", code).Msg("egress filter: the command was terminated because the firewall rules were modified while it was running, failing the step")
			return 1
		}
		return code
	default:
		return 130 // interrupted while the filter was starting, the command never ran
	}
}

// caBundleEnv points an `SSL_CERT_FILE` override of the environment to the filter's ca bundle (the override's roots +
// the inspection ca), clients using the override would reject inspected connections otherwise. The system bundles are
// patched in place and need no override.
func caBundleEnv(bundle string, getenv func(string) string) []string {
	if bundle == "" || getenv("SSL_CERT_FILE") == "" {
		return nil
	}
	if _, err := os.Stat(bundle); err != nil {
		return nil // the filter doesn't inspect tls, no bundle written
	}
	return []string{"SSL_CERT_FILE=" + bundle}
}

// childProcess runs a command with inherited stdio and forwards SIGINT / SIGTERM to it
type childProcess struct {
	command []string
	env     []string

	mu          sync.Mutex
	cmd         *exec.Cmd
	interrupted bool
}

// forwardSignals forwards SIGINT / SIGTERM to the running command; before the command started, a signal calls abort
// (stops the filter setup) and the command is never started. Returns a function to stop forwarding.
func (p *childProcess) forwardSignals(abort func()) func() {
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			case sig := <-signals:
				p.mu.Lock()
				p.interrupted = true
				if p.cmd != nil && p.cmd.Process != nil {
					log.Info().Str("signal", sig.String()).Msg("forwarding signal to command, egress filter stops once it exited")
					_ = p.cmd.Process.Signal(sig)
				} else {
					abort()
				}
				p.mu.Unlock()
			}
		}
	}()
	return func() {
		signal.Stop(signals)
		close(done)
	}
}

func (p *childProcess) isInterrupted() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.interrupted
}

func (p *childProcess) start() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.interrupted {
		return errors.New("interrupted before the command was started")
	}
	c := exec.Command(p.command[0], p.command[1:]...) //nolint:gosec // the command is the caller's intent
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = append(os.Environ(), p.env...)
	if err := c.Start(); err != nil {
		log.Error().Err(err).Strs("command", p.command).Msg("failed to start command")
		return err
	}
	p.cmd = c
	return nil
}

// kill sends SIGKILL to the command and all its descendants, the egress filter stops once it exited
func (p *childProcess) kill() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.interrupted = true // never start it later
	if p.cmd != nil && p.cmd.Process != nil {
		killProcessTree(p.cmd.Process)
	}
}

// wait waits for the started command and returns its exit code
func (p *childProcess) wait() int {
	err := p.cmd.Wait()
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			return 128 + int(status.Signal()) // killed by a signal, shell convention
		}
		return exitErr.ExitCode()
	}
	log.Error().Err(err).Strs("command", p.command).Msg("command failed")
	return 1
}
