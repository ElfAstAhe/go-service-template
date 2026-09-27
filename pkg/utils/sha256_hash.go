package utils

import (
	"crypto/sha256"
	"encoding/hex" // Для красивого вывода строк

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// SHA256Hash реализует криптографический интерфейс Cipher, используя алгоритм одностороннего
// криптографического хэширования SHA-256 (Secure Hash Algorithm 2).
//
// 💡 Архитектурный нюанс (Паттерн Односторонний Адаптер):
// Структура является полностью Stateless (не хранит состояние), что гарантирует абсолютную
// потокобезопасность. Применяется для генерации контрольных сумм файлов, маскирования
// персональных данных (PII) и безопасного хранения секретов (паролей, токенов) в СУБД.
type SHA256Hash struct{}

// Гарантируем полное соответствие интерфейсу Cipher на этапе компиляции
var _ Cipher = (*SHA256Hash)(nil)

// NewSHA256Hash — фабричный конструктор хэшера SHA-256.
func NewSHA256Hash() *SHA256Hash {
	return &SHA256Hash{}
}

// Encrypt вычисляет бинарный хэш SHA-256 от переданного массива байт.
// Всегда возвращает фиксированный массив длиной 32 байта (256 бит).
func (sh *SHA256Hash) Encrypt(data []byte) ([]byte, error) {
	// Вычисляем хэш за один вызов — это экстремально быстро на уровне ассемблерных инструкций CPU
	hash := sha256.Sum256(data)
	return hash[:], nil
}

// EncryptString вычисляет хэш от строки и упаковывает результат в стандартный печатный Hex-формат (64 символа).
func (sh *SHA256Hash) EncryptString(s string) (string, error) {
	res, err := sh.Encrypt([]byte(s))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(res), nil
}

// Decrypt блокирует операцию дешифрования, возвращая фатальную системную ошибку.
// Хэширование является необратимой математической функцией (One-Way Function).
func (sh *SHA256Hash) Decrypt(data []byte) ([]byte, error) {
	return nil, errs.NewCommonError("SHA256 is a one-way hash algorithm and mathematically cannot be decrypted", nil)
}

// DecryptString блокирует операцию дешифрования Hex-строк, возвращая фатальную системную ошибку.
func (sh *SHA256Hash) DecryptString(s string) (string, error) {
	return "", errs.NewCommonError("SHA256 is a one-way hash algorithm and mathematically cannot be decrypted", nil)
}
