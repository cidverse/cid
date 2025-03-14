package deployment

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cidverse/cid/pkg/core/secret"
)

type DecodeSecretsConfig struct {
	PGPPrivateKey         string
	PGPPrivateKeyPassword string
}

var (
	ErrSecretDecryptionFailed = fmt.Errorf("secret decryption failed")
)

func DecodeSecrets(env map[string]string, conf DecodeSecretsConfig) (map[string]string, error) {
	secrets := make(map[string]string)

	for key, value := range env {
		if strings.HasPrefix(value, "ENC[") && strings.HasSuffix(value, "]") {
			innerValue := value[4 : len(value)-1] // remove ENC[ ]

			switch {
			case strings.HasPrefix(innerValue, "base64:"):
				encodedValue := strings.TrimPrefix(innerValue, "base64:")
				decoded, err := secret.DecodeBase64(encodedValue)

				if err != nil {
					return nil, errors.Join(ErrSecretDecryptionFailed, fmt.Errorf("decoding property %s", key), err)
				}

				secrets[key] = decoded
			case strings.HasPrefix(innerValue, "pgp:") && conf.PGPPrivateKey != "":
				secretValue := strings.TrimPrefix(innerValue, "pgp:")

				if !strings.HasPrefix(secretValue, "-----BEGIN PGP MESSAGE-----") {
					decoded, err := secret.DecodeBase64(secretValue)
					if err != nil {
						return nil, errors.Join(ErrSecretDecryptionFailed, fmt.Errorf("decoding property %s", key), err)
					}

					secretValue = decoded
				}

				decrypted, err := secret.DecryptOpenPGP(conf.PGPPrivateKey, conf.PGPPrivateKeyPassword, secretValue)
				if err != nil {
					return nil, errors.Join(ErrSecretDecryptionFailed, fmt.Errorf("decoding property %s", key), err)
				}

				secrets[key] = decrypted
			default:
				secrets[key] = value
			}
		} else {
			secrets[key] = value
		}
	}

	return secrets, nil
}
