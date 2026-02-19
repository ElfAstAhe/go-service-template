package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// AesGcmCipher -AES/GCM cipher util
type AesGcmCipher struct {
	gcm cipher.AEAD
}

// NewAesGcmCipher - aes/gcm util constructor
func NewAesGcmCipher(key []byte) (*AesGcmCipher, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errs.NewUtlCipherError("error create cipher", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errs.NewUtlCipherError("error create gcm", err)
	}

	return &AesGcmCipher{
		gcm: gcm,
	}, nil
}

// MustNewAesGcmCipher - aes/gcm util constructor, returns instance, but generate panic on error
func MustNewAesGcmCipher(key []byte) *AesGcmCipher {
	instance, err := NewAesGcmCipher(key)
	if err != nil {
		panic(err)
	}

	return instance
}

func (a *AesGcmCipher) Encrypt(data []byte) ([]byte, error) {
	nonce := make([]byte, a.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errs.NewUtlCipherError("error fill nonce", err)
	}

	// Seal(dst, nonce, plaintext, ad)
	return a.gcm.Seal(nonce, nonce, data, nil), nil
}

func (a *AesGcmCipher) EncryptString(s string) (string, error) {
	res, err := a.Encrypt([]byte(s))

	return hex.EncodeToString(res), err
}

func (a *AesGcmCipher) Decrypt(data []byte) ([]byte, error) {
	nonceSize := a.gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errs.NewUtlCipherError("error data validation", errs.NewInvalidArgumentError("data", data))
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plain, err := a.gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return nil, errs.NewUtlCipherError("error decrypt data", err)
	}

	return plain, nil
}

func (a *AesGcmCipher) DecryptString(s string) (string, error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return "", errs.NewUtlCipherError("error decoding hex string", err)
	}
	res, err := a.Decrypt(data)

	return string(res), err
}
