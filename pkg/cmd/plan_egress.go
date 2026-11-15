package cmd

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/cid/pkg/egress"
	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/cidverse/cid/pkg/egress/report"
	"github.com/cidverse/cid/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	egressFilterEnv       = "CID_EGRESS_FILTER"        // default of --egress-filter (off, block, audit)
	egressFilterActiveEnv = "CID_EGRESS_FILTER_ACTIVE" // set for the child process that runs under the filter
	egressFilterOff       = "off"
)

// addPlanEgressFlags registers the egress filter flags of `plan execute`
func addPlanEgressFlags(cmd *cobra.Command) {
	defaultMode := os.Getenv(egressFilterEnv)
	if defaultMode == "" {
		defaultMode = egressFilterOff
	}
	cmd.Flags().String("egress-filter", defaultMode, "run the steps under the builtin egress filter (experimental) - allowed: off, block, audit; defaults to $"+egressFilterEnv+". Allows the network access the steps declare, the workflow's global allow list and the ci server / registry hosts. Requires root / CAP_NET_ADMIN; in audit mode a filter that can't be started is skipped with a warning")
	cmd.Flags().StringArray("egress-allow", nil, "additional allowed endpoint for the egress filter (host:port, *.domain:port, ip:port, cidr, or the short form with type / http rules, see pkg/egress/README.md), repeatable")
	cmd.Flags().StringArray("egress-policy-file", nil, "additional egress policy file (yaml, see pkg/egress/README.md), repeatable")
	cmd.Flags().String("egress-report-file", "", "json-lines file that receives every request outside the egress policy")
}

func egressFilterEnabled(cmd *cobra.Command) bool {
	if os.Getenv(egressFilterActiveEnv) != "" {
		return false // already running under the filter
	}
	mode, _ := cmd.Flags().GetString("egress-filter")
	return mode != "" && mode != egressFilterOff
}

// runPlanExecuteWithEgress starts the egress filter with the network access of the selected steps and re-executes the
// current command as a child process under it. `plan execute` exits the process on step failures, a child process
// guarantees that the filter is always cleaned up. Never returns.
func runPlanExecuteWithEgress(cmd *cobra.Command, projectDir string, plan plangenerate.Plan, networkGlobal plangenerate.StepAccessNetwork, stages []string, steps []string) {
	modeStr, _ := cmd.Flags().GetString("egress-filter")
	mode, err := policy.ParseMode(modeStr)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid --egress-filter, allowed: off, block, audit")
	}
	extraAllow, _ := cmd.Flags().GetStringArray("egress-allow")
	policyFiles, _ := cmd.Flags().GetStringArray("egress-policy-file")
	reportFile, _ := cmd.Flags().GetString("egress-report-file")

	ciHosts, err := ciServerEndpoints(os.Getenv)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid ci server / registry host")
	}
	selected := selectSteps(plan, stages, steps)
	cfg := egress.Config{
		Mode:        mode,
		Allow:       append(egressHostsForSteps(selected, networkGlobal), extraAllow...),
		Endpoints:   ciHosts,
		PolicyFiles: policyFiles,
		ReportFile:  reportFile,
		CABundle:    filepath.Join(os.TempDir(), fmt.Sprintf("cid-egress-ca-%d.crt", os.Getpid())),
		OnSummary: func(file report.File) {
			writeEgressSummary(file, filepath.Join(projectDir, ".dist"), selected)
		},
	}
	log.Info().Str("mode", string(mode)).Strs("allow", cfg.Allow).Int("ci_hosts", len(ciHosts)).Msg("running plan under the egress filter")

	self, err := os.Executable()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to determine the cid executable")
	}
	command := append([]string{self}, os.Args[1:]...)
	env := []string{egressFilterActiveEnv + "=1", util.EgressCABundleEnv + "=" + cfg.CABundle}
	os.Exit(egress.RunCommand(cfg, command, env, mode == policy.ModeAudit))
}

// egressSummaryFile is the egress summary next to each step's state, uploaded with the step's artifacts
const egressSummaryFile = "egress.json"

// writeEgressSummary writes the summary into the artifact directory of every step that ran under the filter. Generated
// ci jobs run one step each; when several steps share a filter run, each file covers all of them (see File.Steps).
func writeEgressSummary(file report.File, artifactDir string, steps []plangenerate.Step) {
	for _, step := range steps {
		file.Steps = append(file.Steps, step.Slug)
	}
	for _, step := range steps {
		path := filepath.Join(artifactDir, step.Slug, egressSummaryFile)
		if err := file.Write(path); err != nil {
			log.Warn().Err(err).Str("step", step.Slug).Msg("failed to write egress summary")
		}
	}
}

// selectSteps returns the steps selected by the stage / step filters (all steps without filters)
func selectSteps(plan plangenerate.Plan, stages []string, steps []string) []plangenerate.Step {
	var result []plangenerate.Step
	for _, step := range plan.Steps {
		if len(stages) > 0 && !slices.Contains(stages, step.Stage) {
			continue
		}
		if len(steps) > 0 && !slices.Contains(steps, step.ID) && !slices.Contains(steps, step.Slug) {
			continue
		}
		result = append(result, step)
	}
	return result
}

// egressHostsForSteps returns the declared network access of the steps, merged with the workflow's global allow list
func egressHostsForSteps(steps []plangenerate.Step, networkGlobal plangenerate.StepAccessNetwork) []string {
	var hosts plangenerate.StepAccessNetwork
	for _, step := range steps {
		hosts = plangenerate.MergeNetworkHosts(hosts, step.Access.Network)
	}
	return plangenerate.MergeNetworkHosts(hosts, networkGlobal)
}

// anyAddress allows the ci server / registry to resolve to private addresses (self-hosted instances), loopback and
// link-local addresses stay blocked by the address guard
var anyAddress = []netip.Prefix{netip.MustParsePrefix("0.0.0.0/0"), netip.MustParsePrefix("::/0")}

// ciServerEndpoints returns the ci server and container registry of the current pipeline, cid talks to them during step
// execution
func ciServerEndpoints(getenv func(string) string) ([]policy.Endpoint, error) {
	endpoints, err := policy.ParseEndpoints(ciServerHosts(getenv))
	if err != nil {
		return nil, err
	}
	for i := range endpoints {
		endpoints[i].AllowedIPs = anyAddress
	}
	return endpoints, nil
}

func ciServerHosts(getenv func(string) string) []string {
	var result []string
	if host := getenv("CI_SERVER_HOST"); host != "" {
		port := getenv("CI_SERVER_PORT")
		if port == "" {
			port = "443"
		}
		result = append(result, net.JoinHostPort(host, port))
	}
	if registry := getenv("CI_REGISTRY"); registry != "" {
		if _, _, err := net.SplitHostPort(registry); err != nil && !strings.Contains(registry, "/") {
			registry = net.JoinHostPort(registry, "443")
		}
		result = append(result, registry)
	}
	return result
}
