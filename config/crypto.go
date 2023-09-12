package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"log"
)

var (
	EncryptionEnabled bool = false
)

func Encrypt(stringKey, stringText string) []byte {
	key := []byte(stringKey)
	text := []byte(stringText)

	if !EncryptionEnabled {
		return text
	}

	// generate a new aes cipher using our 32 byte long key
	c, err := aes.NewCipher(key)
	// if there are any errors, handle them
	if err != nil {
		log.Println(err.Error())
	}

	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(c)
	// if any error generating new GCM
	// handle them
	if err != nil {
		log.Println(err.Error())
	}

	// creates a new byte array the size of the nonce
	// which must be passed to Seal
	nonce := make([]byte, gcm.NonceSize())
	// populates our nonce with a cryptographically secure
	// random sequence
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		log.Println(err.Error())
	}

	// here we encrypt our text using the Seal function
	// Seal encrypts and authenticates plaintext, authenticates the
	// additional data and appends the result to dst, returning the updated
	// slice. The nonce must be NonceSize() bytes long and unique for all
	// time, for a given key.
	return gcm.Seal(nonce, nonce, text, nil)
}

func Decrypt(key, text []byte) []byte {
	if !EncryptionEnabled {
		return text
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err.Error())
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		log.Println(err.Error())
	}

	nonceSize := gcm.NonceSize()
	if len(text) < nonceSize {
		log.Println(err.Error())
	}

	nonce, text := text[:nonceSize], text[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, text, nil)
	if err != nil {
		log.Println(err.Error())
	}

	return plaintext
}
