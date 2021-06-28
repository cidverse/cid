package api

import (
	"bytes"
	"encoding/base64"
	"golang.org/x/crypto/openpgp"
	"io/ioutil"
	"strings"
)

func EncryptOpenPGP(publicKey string, secretString string) (string, error) {
	entityList, err := openpgp.ReadArmoredKeyRing(strings.NewReader(publicKey))
	if err != nil {
		return "", err
	}

	// encrypt string
	buf := new(bytes.Buffer)
	w, err := openpgp.Encrypt(buf, entityList, nil, nil, nil)
	if err != nil {
		return "", err
	}
	_, err = w.Write([]byte(secretString))
	if err != nil {
		return "", err
	}
	err = w.Close()
	if err != nil {
		return "", err
	}

	// Encode to base64
	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return "", err
	}
	encStr := base64.StdEncoding.EncodeToString(bytes)

	return encStr, nil
}

func DecryptOpenPGP(privateKey string, privateKeyPassword string, encString string) (string, error) {
	// init some vars
	var entity *openpgp.Entity
	var entityList openpgp.EntityList

	// Open the private key file
	entityList, err := openpgp.ReadArmoredKeyRing(strings.NewReader(privateKey))
	if err != nil {
		return "", err
	}
	entity = entityList[0]

	// Get the passphrase and read the private key.
	// Have not touched the encrypted string yet
	passphraseByte := []byte(privateKeyPassword)
	err = entity.PrivateKey.Decrypt(passphraseByte)
	if err != nil {
		return "", err
	}
	for _, subKey := range entity.Subkeys {
		err = subKey.PrivateKey.Decrypt(passphraseByte)
		if err != nil {
			return "", err
		}
	}

	// Decode the base64 string
	dec, err := base64.StdEncoding.DecodeString(encString)
	if err != nil {
		return "", err
	}

	// Decrypt it with the contents of the private key
	md, err := openpgp.ReadMessage(bytes.NewBuffer(dec), entityList, nil, nil)
	if err != nil {
		return "", err
	}
	bytes, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		return "", err
	}
	decStr := string(bytes)

	return decStr, nil
}
