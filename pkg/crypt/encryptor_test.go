package crypt

import (
	"github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"testing"
)

func TestEncryptor_DecryptFact(t *testing.T) {
	gomega.RegisterFailHandler(func(message string, callerSkip ...int) {
		t.Errorf(message)
	})

	privateKey := "some-private-key"
	encryptor := NewEncryptor(privateKey)

	expectedFact := uuid.NewV4().String()
	encryptedFact, err := encryptor.EncryptFact(expectedFact)
	if err != nil {
		t.Fatalf(err.Error())
	}

	println(encryptedFact)

	actualFact, err := encryptor.DecryptFact(encryptedFact)
	if err != nil {
		t.Fatalf(err.Error())
	}

	gomega.Expect(actualFact).To(gomega.Equal(expectedFact))
}

func TestEncryptor_twoEncryptor(t *testing.T) {
	gomega.RegisterFailHandler(func(message string, callerSkip ...int) {
		t.Errorf(message)
	})

	privateKey := "some-private-key"
	encryptor := NewEncryptor(privateKey)

	expectedFact := "some-fact"
	encryptedFact, err := encryptor.EncryptFact(expectedFact)
	if err != nil {
		t.Fatalf(err.Error())
	}

	anotherEncryptor := NewEncryptor(privateKey)

	actualFact, err := anotherEncryptor.DecryptFact(encryptedFact)
	if err != nil {
		t.Fatalf(err.Error())
	}

	gomega.Expect(actualFact).To(gomega.Equal(expectedFact))
}

func TestEncryptor_anotherEncryptorFail(t *testing.T) {
	gomega.RegisterFailHandler(func(message string, callerSkip ...int) {
		t.Errorf(message)
	})

	expectedFact := "some-fact"
	encryptor := NewEncryptor("some-private-key")
	encryptedFact, err := encryptor.EncryptFact(expectedFact)
	if err != nil {
		t.Fatalf(err.Error())
	}

	anotherEncryptor := NewEncryptor("some-other-private-key")
	_, err = anotherEncryptor.DecryptFact(encryptedFact)
	gomega.Expect(err).NotTo(gomega.BeNil())
}
