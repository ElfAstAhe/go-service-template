package helper

import (
	"bytes"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

const (
	// CipherStringPrefix Системные маркеры (Префиксы) для однозначной идентификации зашифрованного payload.
	CipherStringPrefix string = "cipher::" // Текстовый префикс для строковых полей СУБД
)

var (
	CipherPrefix = []byte(CipherStringPrefix) // Бинарный префикс для потоков данных
)

// Cipher описывает высокоуровневый контракт хелпера шифрования (Crypto Orchestrator).
// Обогащает базовые алгоритмы логикой проверки состояния (IsEncrypted) и автоматического префиксирования.
type Cipher interface {
	EncryptString(string) string
	DecryptString(string) string
	EncryptBinary([]byte) []byte
	DecryptBinary([]byte) []byte
	IsStringEncrypted(string) bool
	IsEncrypted([]byte) bool
}

// CipherImpl реализует интерфейс Cipher, выступая умной оберткой над базовым utils.Cipher.
type CipherImpl struct {
	cipher utils.Cipher // Ссылка на низкоуровневый криптографический движок (например, AesGcmCipher)
}

// Гарантируем соответствие интерфейсу Cipher на этапе компиляции
var _ Cipher = (*CipherImpl)(nil)

// NewCipherHelper — фабричный конструктор хелпера шифрования.
func NewCipherHelper(cipher utils.Cipher) *CipherImpl {
	return &CipherImpl{
		cipher: cipher,
	}
}

// EncryptString шифрует строку и добавляет текстовый маркер, если она еще не была зашифрована.
func (ch *CipherImpl) EncryptString(s string) string {
	// Предохранитель: защищает от повторного шифрования (Double Encryption Protection)
	if s == "" || ch.IsStringEncrypted(s) {
		return s
	}

	res, err := ch.cipher.EncryptString(s)
	if err != nil {
		return s // В случае сбоя мягко возвращаем исходную строку (Graceful Degradation)
	}

	return CipherStringPrefix + res
}

// DecryptString проверяет маркер и атомарно дешифрует строку, удаляя префикс схемы.
func (ch *CipherImpl) DecryptString(s string) string {
	if s == "" || !ch.IsStringEncrypted(s) {
		return s
	}

	encrypted := strings.TrimPrefix(s, CipherStringPrefix)
	if encrypted == "" {
		return s
	}

	res, err := ch.cipher.DecryptString(encrypted)
	if err != nil {
		return s
	}

	return res
}

// EncryptBinary атомарно склеивает бинарный префикс и шифротекст в единый монолитный слайс.
func (ch *CipherImpl) EncryptBinary(data []byte) []byte {
	if ch.IsEncrypted(data) {
		return data
	}

	encrypted, err := ch.cipher.Encrypt(data)
	if err != nil {
		return data
	}

	prefixLen := len(CipherPrefix)
	// Пре-аллоцируем срез точного размера для исключения лишних аллокаций в куче
	res := make([]byte, prefixLen+len(encrypted))

	// Эффективно копируем блоки памяти
	copy(res, CipherPrefix)
	copy(res[prefixLen:], encrypted)

	return res
}

// DecryptBinary извлекает шифротекст, отсекая бинарный префикс, и производит дешифрование.
func (ch *CipherImpl) DecryptBinary(data []byte) []byte {
	if !ch.IsEncrypted(data) {
		return data
	}

	prefixLen := len(CipherPrefix)

	// Расшифровываем payload, передавая слайс со смещением на длину префикса
	res, err := ch.cipher.Decrypt(data[prefixLen:])
	if err != nil {
		return data
	}

	return res
}

// IsStringEncrypted проверяет наличие текстового маркера схемы шифрования в начале строки.
func (ch *CipherImpl) IsStringEncrypted(s string) bool {
	return strings.HasPrefix(s, CipherStringPrefix)
}

// IsEncrypted выполняет быструю атомарную проверку бинарного префикса за константное время O(1).
func (ch *CipherImpl) IsEncrypted(data []byte) bool {
	prefixLen := len(CipherPrefix)

	// Предохранитель: исключает панику index out of range при проверке коротких пакетов
	if len(data) < prefixLen {
		return false
	}

	// Сравниваем только начальный срез памяти с бинарным синглтоном префикса
	return bytes.Equal(data[:prefixLen], CipherPrefix)
}
