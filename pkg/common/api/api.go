package api

import (
	"slices"
	"strings"

	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/deployment"
	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/cidverse/normalizeci/pkg/envstruct"
	"github.com/cidverse/normalizeci/pkg/normalizer"
	"github.com/cidverse/normalizeci/pkg/normalizer/api"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

var systemEnvVariables = []string{
	"PATH",
	"HOME",
	"PWD",
	"OLDPWD",
	"PAGER",
	"SHELL",
	"LOCALE_ARCHIVE",
	"LOGNAME",
	"INFOPATH",
	"LS_COLORS",
	"SSH_AUTH_SOCK",
	"LIBEXEC_PATH",
	"TERMINFO_DIRS",
	"PULSE_SERVER",
	"LESSKEYIN_SYSTEM",
	"GTK_PATH",
	"GIO_EXTRA_MODULES",
	"DBUS_SESSION_BUS_ADDRESS",
	"QTWEBKIT_PLUGIN_PATH",
	"SPEECHD_CMD",
	"TZDIR",
	"DISPLAY",
	"WAYLAND_DISPLAY",
	"GPG_TTY",
	"GTK_A11Y",
	"LESSOPEN",
	"NIXPKGS_CONFIG",
	"NIX_LD",
	"NIX_PATH",
	"NIX_PROFILES",
	"NIX_LD_LIBRARY_PATH",
	"NIX_USER_PROFILE_DIR",
	"NIX_XDG_DESKTOP_PORTAL_DIR",
	"WT_SESSION",
	"WT_PROFILE_ID",
	"XCURSOR_PATH",
	"XDG_CONFIG_DIRS",
	"XDG_DATA_DIRS",
	"XDG_RUNTIME_DIR",
	"WSLPATH",
	"WSLENV",
	"WSL_INTEROP",
	"WSL_DISTRO_NAME",
	"DONT_PROMPT_WSL_INSTALL",
	"WSL2_GUI_APPS_ENABLED",
	"_",
}

// GetCIDEnvironment returns the normalized ci variables
func GetCIDEnvironment(configEnv map[string]string, projectDirectory string) (map[string]string, error) {
	// TODO: allow overriding of NCI_ variables?
	normalized, err := normalizer.NormalizeEnv(normalizer.Options{ProjectDir: projectDirectory, Env: api.GetMachineEnvironment()})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to prepare ci environment variables")
	}
	env := envstruct.StructToEnvMap(normalized)
	for key := range env {
		if !strings.HasPrefix(key, "NCI") {
			delete(env, key)
		}
	}

	// append cid vars
	env["CID_CONVENTION_BRANCHING"] = string(config.Current.Conventions.Branching)
	env["CID_CONVENTION_COMMIT"] = string(config.Current.Conventions.Commit)

	// prio 3: cid config file
	for key, value := range configEnv {
		env[key] = value
	}

	// prio 2: dotenv override
	if filesystem.FileExists(".env") {
		dotEnv, err := godotenv.Read(".env")
		if err != nil {
			return nil, err
		}

		for k, v := range dotEnv {
			env[k] = v
		}
	}

	// prio 1: system env override
	for k, v := range api.GetMachineEnvironment() {
		if slices.Contains(systemEnvVariables, k) || strings.HasPrefix(k, "_") {
			continue
		}
		env[k] = v
	}

	// decode encoded or encrypted values
	redact.ProtectPhrase(env["CID_SECRET_PGP_PRIVATE_KEY"])
	redact.ProtectPhrase(env["CID_SECRET_PGP_PRIVATE_KEY_PASSWORD"])
	decodedEnv, err := deployment.DecodeSecrets(env, deployment.DecodeSecretsConfig{
		PGPPrivateKey:         env["CID_SECRET_PGP_PRIVATE_KEY"],
		PGPPrivateKeyPassword: env["CID_SECRET_PGP_PRIVATE_KEY_PASSWORD"],
	})

	// customization
	// - suggested release version
	/*
		enrichErr := EnrichEnvironment(projectDirectory, string(config.Current.Conventions.Branching), env)
		if enrichErr != nil {
			log.Err(enrichErr).Msg("failed to enrich project context")
		}
	*/

	return decodedEnv, nil
}

func AutoProtectValues(key string, original string, decoded string) {
	upperKey := strings.ToUpper(key)
	if strings.Contains(upperKey, "KEY") || strings.Contains(upperKey, "USER") || strings.Contains(upperKey, "PASS") || strings.Contains(upperKey, "PRIVATE") || strings.Contains(upperKey, "TOKEN") || strings.Contains(upperKey, "SECRET") || strings.Contains(upperKey, "AUTH") {
		if original != "" {
			redact.ProtectPhrase(original)
		}
		if decoded != "" {
			redact.ProtectPhrase(decoded)
		}
	}
}
