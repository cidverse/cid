package api

import (
	"encoding/base64"
	"strings"

	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/secret"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/cidverse/normalizeci/pkg/envstruct"
	"github.com/cidverse/normalizeci/pkg/normalizer"
	"github.com/cidverse/normalizeci/pkg/normalizer/api"
	"github.com/rs/zerolog/log"
)

// GetCIDEnvironment returns the normalized ci variables
func GetCIDEnvironment(configEnv map[string]string, projectDirectory string) map[string]string {
	normalized, err := normalizer.Normalize()
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

	// append env from configuration file
	for key, value := range configEnv {
		env[key] = value
	}

	// decode all values
	for key, value := range env {
		env[key] = DecodeEnvValue(value)
	}

	// customization
	// - suggested release version
	/*
		enrichErr := EnrichEnvironment(projectDirectory, string(config.Current.Conventions.Branching), env)
		if enrichErr != nil {
			log.Err(enrichErr).Msg("failed to enrich project context")
		}
	*/

	return env
}

func DecodeEnvValue(value string) string {
	// Base64
	if strings.HasPrefix(value, "base64~") {
		dec, decErr := base64.StdEncoding.DecodeString(strings.TrimPrefix(value, "base64~"))
		if decErr == nil {
			return string(dec)
		}
	}
	// OpenPGP
	if strings.HasPrefix(value, "openpgp~") {
		// todo: cache
		machineEnv := api.GetMachineEnvironment()
		privateKey := machineEnv["CID_MASTER_GPG_PRIVATEKEY"]
		privateKeyPassphrase := machineEnv["CID_MASTER_GPG_PASSWORD"]

		dec, decErr := secret.DecryptOpenPGP(privateKey, privateKeyPassphrase, strings.TrimPrefix(value, "openpgp~"))
		if decErr == nil {
			return dec
		}
	}

	return value
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
