package api

import (
	"github.com/ProtonMail/gopenpgp/v2/helper"
)

// EncryptOpenPGP encrypts a secretString with the publicKey
func EncryptOpenPGP(publicKey string, secretString string) (string, error) {
	// encrypt plain text message using public key
	return helper.EncryptMessageArmored(publicKey, secretString)
}

// DecryptOpenPGP decrypts a encrypted string
func DecryptOpenPGP(privateKey string, privateKeyPassword string, encString string) (string, error) {
	// decrypt armored encrypted message using the private key and obtain plain text
	return helper.DecryptMessageArmored(privateKey, []byte(privateKeyPassword), encString)
}
