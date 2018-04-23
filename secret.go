package main

import (
	"encoding/hex"
	"io"
	"crypto/rand"
	"crypto/aes"
)

type Secret struct {
	key   []byte
	nonce []byte
	aesecb *ecb
}

func newSecret(key string) (*Secret, error) {
	secret := new(Secret)
	k, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	secret.key = k

	// Generate nonce.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	secret.nonce = nonce

	// Set block and cipher
	block, err := aes.NewCipher(secret.key)
	if err != nil {
		panic(err.Error())
	}

	secret.aesecb = newECB(block)

	return secret, nil
}

func (secret *Secret) encrypt(value []byte) {
	//return secret.aesecb.Seal(nil, secret.nonce, []byte(value), nil)
	encrypter := ecbEncrypter(*secret.aesecb)
	encrypter.CryptBlocks(value, value)
}

func (secret *Secret) decrypt(value []byte) {
	//return secret.aesgcm.Open(nil, secret.nonce, value, nil)
	decrypter := ecbDecrypter(*secret.aesecb)
	decrypter.CryptBlocks(value, value)
}