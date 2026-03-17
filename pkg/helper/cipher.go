package helper

import (
	"bytes"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

const (
	CipherStringPrefix string = "cipher::"
)

var (
	CipherPrefix = []byte(CipherStringPrefix)
)

type Cipher interface {
	EncryptString(string) string
	DecryptString(string) string
	EncryptBinary([]byte) []byte
	DecryptBinary([]byte) []byte
	IsStringEncrypted(string) bool
	IsEncrypted([]byte) bool
}

type CipherImpl struct {
	cipher utils.Cipher
}

func NewCipherHelper(cipher utils.Cipher) *CipherImpl {
	return &CipherImpl{
		cipher: cipher,
	}
}

func (ch *CipherImpl) EncryptString(s string) string {
	if s == "" || ch.IsStringEncrypted(s) {
		return s
	}

	// шифруем
	res, err := ch.cipher.EncryptString(s)
	if err != nil {
		return s
	}

	// результат в base64 + prefix
	return CipherStringPrefix + res
}

func (ch *CipherImpl) DecryptString(s string) string {
	if s == "" || !ch.IsStringEncrypted(s) {
		return s
	}

	// убираем префикс и проверяем есть хоть что-нибудь
	encrypted := strings.TrimPrefix(s, CipherStringPrefix)
	if encrypted == "" {
		return s
	}

	// расшифровываем
	res, err := ch.cipher.DecryptString(encrypted)
	if err != nil {
		return s
	}

	// возвращаем результат
	return res
}

func (ch *CipherImpl) EncryptBinary(data []byte) []byte {
	if ch.IsEncrypted(data) {
		return data
	}

	encrypted, err := ch.cipher.Encrypt(data)
	if err != nil {
		return data
	}

	prefixLen := len(CipherPrefix)
	res := make([]byte, prefixLen+len(encrypted))

	// Копируем части
	copy(res, CipherPrefix)
	copy(res[prefixLen:], encrypted)

	return res
}

func (ch *CipherImpl) DecryptBinary(data []byte) []byte {
	if !ch.IsEncrypted(data) {
		return data
	}

	prefixLen := len(CipherPrefix)

	// расшифровываем
	res, err := ch.cipher.Decrypt(data[prefixLen:])
	if err != nil {
		return data
	}

	return res
}

func (ch *CipherImpl) IsStringEncrypted(s string) bool {
	return strings.HasPrefix(s, CipherStringPrefix)
}

func (ch *CipherImpl) IsEncrypted(data []byte) bool {
	prefixLen := len(CipherPrefix)

	// Проверяем, что данных достаточно, чтобы в них физически мог быть префикс
	if len(data) < prefixLen {
		return false
	}

	// Сравниваем только начальную часть данных с префиксом
	return bytes.Equal(data[:prefixLen], CipherPrefix)
}
