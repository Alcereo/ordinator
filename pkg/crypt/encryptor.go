package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"gopkg.in/go-playground/validator.v9"
	"io"
)

type Encryptor struct {
	Gsm   cipher.AEAD
	nonce []byte
}

var validate = validator.New()

func NewEncryptor(privateKey string) *Encryptor {
	if privateKey == "" {
		panic("PrivateKey is required to create Encryptor")
	}
	gcm, nonce := generateEncryptEntities(privateKey)

	encryptor := &Encryptor{
		Gsm:   gcm,
		nonce: nonce,
	}

	if err := validate.Struct(encryptor); err != nil {
		panic(err.Error())
	}
	return encryptor
}

func generateEncryptEntities(privateKey string) (cipher.AEAD, []byte) {
	block, _ := aes.NewCipher(createHash(privateKey))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	return gcm, nonce
}

func (encryptor *Encryptor) EncryptFact(fact string) (string, error) {
	factBytes := []byte(fact)
	encryptedText := encryptor.Gsm.Seal(encryptor.nonce, encryptor.nonce, factBytes, nil)
	hexString := hex.EncodeToString(encryptedText)
	return hexString, nil
}

func createHash(value string) []byte {
	hasher := md5.New()
	hasher.Write([]byte(value))
	return hasher.Sum(nil)
}

func (encryptor *Encryptor) DecryptFact(encryptedFact string) (string, error) {
	encryptedBytes, err := hex.DecodeString(encryptedFact)
	if err != nil {
		return "", err
	}
	nonceSize := encryptor.Gsm.NonceSize()
	nonce, ciphertext := encryptedBytes[:nonceSize], encryptedBytes[nonceSize:]
	plaintext, err := encryptor.Gsm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
