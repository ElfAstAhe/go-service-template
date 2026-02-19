package helper

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

const (
	RSAKey2048 = 2048
	RSAKey4096 = 4096
)

type RSAKeysHelper struct {
	bits int // 2048 или 4096
}

func NewRSAKeysHelper(bits int) *RSAKeysHelper {
	return &RSAKeysHelper{
		bits: bits,
	}
}

func (kh *RSAKeysHelper) Generate() (string, string, error) {
	// 1. Генерируем ключи
	privateKey, err := rsa.GenerateKey(rand.Reader, kh.bits)
	if err != nil {
		return "", "", errs.NewUtlCipherError("generate private key", err)
	}

	// 2. Кодируем приватный ключ
	privateBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateBytes,
	})

	// 3. Кодируем публичный ключ
	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", errs.NewUtlCipherError("generate public key", err)
	}
	publicPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubBytes,
	})

	return string(privatePem), string(publicPem), nil
}

func (kh *RSAKeysHelper) ParsePrivateKey(pemString string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		return nil, errs.NewUtlCipherError("failed decode PEM block", nil)
	}

	// Мы использовали x509.MarshalPKCS1PrivateKey при генерации
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errs.NewUtlCipherError("failed to parse PEM block", err)
	}

	return privateKey, nil
}

func (kh *RSAKeysHelper) ParsePublicKey(pemString string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		return nil, errs.NewUtlCipherError("failed decode PEM block", nil)
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errs.NewUtlCipherError("failed to parse PEM block", err)
	}

	switch pub := publicKey.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, errs.NewUtlCipherError("invalid RSA public key type", nil)
	}
}

func (kh *RSAKeysHelper) Decrypt(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	res, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, data, nil)
	if err != nil {
		return nil, errs.NewUtlCipherError("decrypt error", err)
	}

	return res, nil
}

func (kh *RSAKeysHelper) DecryptString(data string, privateKey *rsa.PrivateKey) (string, error) {
	encrypted, err := hex.DecodeString(data)
	if err != nil {
		return "", errs.NewUtlCipherError("hex decode error", err)
	}
	res, err := kh.Decrypt(encrypted, privateKey)
	if err != nil {
		return "", errs.NewUtlCipherError("decrypt string error", err)
	}

	return string(res), nil
}

func (kh *RSAKeysHelper) Encrypt(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	res, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, data, nil)
	if err != nil {
		return nil, errs.NewUtlCipherError("encrypt error", err)
	}

	return res, nil
}

func (kh *RSAKeysHelper) EncryptString(data string, publicKey *rsa.PublicKey) (string, error) {
	res, err := kh.Encrypt([]byte(data), publicKey)
	if err != nil {
		return "", errs.NewUtlCipherError("encrypt string error", err)
	}

	return hex.EncodeToString(res), nil
}
