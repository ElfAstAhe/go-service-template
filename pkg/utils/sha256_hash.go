package utils

import (
	"crypto/sha256"
	"encoding/hex" // Для красивого вывода строк

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

type SHA256Hash struct{}

func NewSHA256Hash() *SHA256Hash {
	return &SHA256Hash{}
}

// Encrypt - потокобезопасен, так как не хранит состояние внутри структуры
func (sh *SHA256Hash) Encrypt(data []byte) ([]byte, error) {
	// Вычисляем хеш за один вызов — это быстро и безопасно
	hash := sha256.Sum256(data)
	return hash[:], nil
}

func (sh *SHA256Hash) EncryptString(s string) (string, error) {
	res, err := sh.Encrypt([]byte(s))
	if err != nil {
		return "", err
	}

	// Возвращаем HEX-строку (например, "e3b0c442...")
	return hex.EncodeToString(res), nil
}

// Decrypt — здесь лучше возвращать специфичную ошибку "Unsupported",
// так как хеш в принципе нельзя расшифровать.
func (sh *SHA256Hash) Decrypt(data []byte) ([]byte, error) {
	return nil, errs.NewCommonError("SHA256 is a one-way hash and cannot be decrypted", nil)
}

func (sh *SHA256Hash) DecryptString(s string) (string, error) {
	return "", errs.NewCommonError("SHA256 is a one-way hash and cannot be decrypted", nil)
}
