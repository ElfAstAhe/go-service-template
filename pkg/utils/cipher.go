package utils

// Cipher описывает абстрактный агностичный контракт для сквозного шифрования и дешифрования данных (Crypto Wrapper).
//
// Обеспечивает выполнение требований безопасности (Data-at-Rest и Data-in-Transit Encryption):
//  1. Методы для []byte ориентированы на высокопроизводительную потоковую обработку файлов, сетевых пакетов и payload.
//  2. Методы для string предназначены для удобной маскировки конфиденциальных персональных данных (PII) в базах данных.
type Cipher interface {
	// Encrypt выполняет симметричное шифрование сырого массива байт.
	Encrypt(plaintext []byte) (ciphertext []byte, err error)

	// EncryptString шифрует строку и возвращает безопасный текстовый результат (обычно в кодировке Base64 или Hex).
	EncryptString(plaintext string) (ciphertext string, err error)

	// Decrypt осуществляет дешифрование зашифрованного массива байт, проверяя целостность пакета.
	Decrypt(ciphertext []byte) (plaintext []byte, err error)

	// DecryptString принимает зашифрованную строку (Base64/Hex) и возвращает исходный текстовый plaintext.
	DecryptString(ciphertext string) (plaintext string, err error)
}
