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

type CipherHelper struct {
	cipher utils.Cipher
}

func NewCipherHelper(cipher utils.Cipher) *CipherHelper {
	return &CipherHelper{
		cipher: cipher,
	}
}

func (ch *CipherHelper) EncryptString(s string) string {
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

func (ch *CipherHelper) DecryptString(s string) string {
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

func (ch *CipherHelper) EncryptBinary(data []byte) []byte {
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

func (ch *CipherHelper) DecryptBinary(data []byte) []byte {
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

func (ch *CipherHelper) IsStringEncrypted(s string) bool {
	return strings.HasPrefix(s, CipherStringPrefix)
}

func (ch *CipherHelper) IsEncrypted(data []byte) bool {
	prefixLen := len(CipherPrefix)

	// Проверяем, что данных достаточно, чтобы в них физически мог быть префикс
	if len(data) < prefixLen {
		return false
	}

	// Сравниваем только начальную часть данных с префиксом
	return bytes.Equal(data[:prefixLen], CipherPrefix)
}
