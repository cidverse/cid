package secret

// see: https://github.com/ProtonMail/gopenpgp

import (
	"github.com/ProtonMail/gopenpgp/v2/helper"
)

// EncryptOpenPGP encrypts a secretString with the publicKey using OpenPGP
func EncryptOpenPGP(publicKey string, secretString string) (string, error) {
	// encrypt plain text message using public key
	return helper.EncryptMessageArmored(publicKey, secretString)
}

// DecryptOpenPGP decrypts an encrypted string using OpenPGP
func DecryptOpenPGP(privateKey string, privateKeyPassword string, encString string) (string, error) {
	// decrypt armored encrypted message using the private key and obtain plain text
	return helper.DecryptMessageArmored(privateKey, []byte(privateKeyPassword), encString)
}
