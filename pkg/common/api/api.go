package api

import (
	"encoding/base64"
	"strings"

	"github.com/cidverse/cid/pkg/common/protectoutput"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/secret"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/normalizeci/pkg/common"
	"github.com/cidverse/normalizeci/pkg/ncispec"
	ncimain "github.com/cidverse/normalizeci/pkg/normalizeci"
	"github.com/rs/zerolog/log"
)

// FindProjectDir finds the project directory from the current dir
func FindProjectDir() string {
	projectDirectory, projectDirectoryErr := filesystem.GetProjectDirectory()
	if projectDirectoryErr != nil {
		log.Fatal().Err(projectDirectoryErr).Msg(projectDirectoryErr.Error())
	}

	return projectDirectory
}

// GetCIDEnvironment returns the normalized ci variables
func GetCIDEnvironment(configEnv map[string]string, projectDirectory string) map[string]string {
	spec := ncimain.Normalize()
	env := ncispec.ToMap(spec)
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
		machineEnv := common.GetMachineEnvironment()
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
			protectoutput.ProtectPhrase(original)
		}
		if decoded != "" {
			protectoutput.ProtectPhrase(decoded)
		}
	}
}
