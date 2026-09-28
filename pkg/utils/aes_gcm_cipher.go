package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

// AesGcmCipher реализует криптографический интерфейс Cipher, используя алгоритм симметричного
// блочного шифрования AES в режиме счетчика с аутентификацией Галуа (GCM - Galois/Counter Mode).
//
// Режим GCM является индустриальным стандартом шифрования (AEAD), обеспечивая одновременную
// конфиденциальность, подлинность и проверку целостности данных на уровне рантайма фреймворка.
type AesGcmCipher struct {
	gcm cipher.AEAD // Низкоуровневый интерфейс аутентифицированного шифрования Go
}

// Гарантируем полное соответствие интерфейсу Cipher на этапе компиляции
var _ Cipher = (*AesGcmCipher)(nil)

// NewAesGcmCipher — фабричный конструктор криптографического узла.
// Принимает секретный ключ длиной 16 байт (AES-128), 24 байта (AES-192) или 32 байта (AES-256).
func NewAesGcmCipher(key []byte) (*AesGcmCipher, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errs.NewUtlCipherError("error create cipher block", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errs.NewUtlCipherError("error create gcm mode", err)
	}

	return &AesGcmCipher{
		gcm: gcm,
	}, nil
}

// MustNewAesGcmCipher — конструктор-предохранитель для DI-контейнеров (например, InfraContainer).
// Возвращает живой инстанс, но инициирует панику, если передан невалидный ключ (неверная длина).
func MustNewAesGcmCipher(key []byte) *AesGcmCipher {
	instance, err := NewAesGcmCipher(key)
	if err != nil {
		panic(err)
	}

	return instance
}

// Encrypt выполняет симметричное шифрование байтового массива с генерацией уникального Nonce.
// На выходе формируется монолитный слайс структуры [Nonce (12 байт)][Ciphertext].
func (a *AesGcmCipher) Encrypt(data []byte) ([]byte, error) {
	nonce := make([]byte, a.gcm.NonceSize())
	// Используем криптографически стойкий системный генератор (rand.Reader) для исключения коллизий векторов
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errs.NewUtlCipherError("error fill random nonce", err)
	}

	// Экономим аллокации: пишем nonce в dst буфер, шифрованные данные приклеятся следом автоматически
	return a.gcm.Seal(nonce, nonce, data, nil), nil
}

// EncryptString шифрует строку и упаковывает бинарный результат в печатный Hex-формат.
func (a *AesGcmCipher) EncryptString(s string) (string, error) {
	res, err := a.Encrypt([]byte(s))

	return hex.EncodeToString(res), err
}

// Decrypt проверяет целостность входящего пакета, извлекает Nonce и атомарно дешифрует данные.
func (a *AesGcmCipher) Decrypt(data []byte) ([]byte, error) {
	nonceSize := a.gcm.NonceSize()
	// Предохранитель: защищаем рантайм от паники index out of range при битых бинарных пакетах
	if len(data) < nonceSize {
		return nil, errs.NewUtlCipherError("error data validation", errs.NewInvalidArgumentError("data", data))
	}

	// Разрезаем монолит на вектор инициализации и зашифрованный payload
	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plain, err := a.gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		// Ошибка вернется, если данные были модифицированы или подделан проверочный тег (Tag Validation Fail)
		return nil, errs.NewUtlCipherError("error decrypt data / integrity check failed", err)
	}

	return plain, nil
}

// DecryptString принимает Hex-строку, распаковывает её в бинарный вид и производит дешифрование.
func (a *AesGcmCipher) DecryptString(s string) (string, error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return "", errs.NewUtlCipherError("error decoding hex string", err)
	}
	res, err := a.Decrypt(data)

	return string(res), err
}
