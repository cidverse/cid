package secret

// see: https://github.com/ProtonMail/gopenpgp

import (
	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

// EncryptOpenPGP encrypts a secretString with the publicKey using OpenPGP (armored encrypted message)
func EncryptOpenPGP(publicKey string, secretString string) (string, error) {
	pubKey, err := crypto.NewKeyFromArmored(publicKey)
	if err != nil {
		return "", err
	}

	// encrypt plain text message using public key
	pgp := crypto.PGP()
	encHandle, err := pgp.Encryption().Recipient(pubKey).New()
	if err != nil {
		return "", err
	}
	pgpMessage, err := encHandle.Encrypt([]byte(secretString))
	if err != nil {
		return "", err
	}
	armoredSecret, err := pgpMessage.ArmorBytes()
	if err != nil {
		return "", err
	}

	return string(armoredSecret), nil
}

// DecryptOpenPGP decrypts an encrypted string using OpenPGP (armored encrypted message)
func DecryptOpenPGP(privateKey string, privateKeyPassword string, encString string) (string, error) {
	privKey, err := crypto.NewPrivateKeyFromArmored(privateKey, []byte(privateKeyPassword))
	if err != nil {
		return "", err
	}

	// decrypt armored encrypted message using the private key and obtain plain text
	pgp := crypto.PGP()
	decHandle, err := pgp.Decryption().DecryptionKey(privKey).New()
	if err != nil {
		return "", err
	}
	decrypted, err := decHandle.Decrypt([]byte(encString), crypto.Armor)
	if err != nil {
		return "", err
	}
	decHandle.ClearPrivateParams()

	return string(decrypted.Bytes()), nil
}
