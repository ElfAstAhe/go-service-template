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

// Допустимые и криптографически стойкие размеры ключей RSA.
const (
	RSAKey2048 = 2048 // Минимально допустимый enterprise-стандарт длины ключа
	RSAKey4096 = 4096 // Ультимативный уровень стойкости для долговечных секретов
)

// RSAKeys расширяет базовый дженерик-интерфейс Keys для указателей на структуры RSA.
type RSAKeys interface {
	Keys[*rsa.PrivateKey, *rsa.PublicKey]
}

// RSAKeysHelper реализует интерфейс RSAKeys, инкапсулируя логику генерации,
// маршалинга в форматы PEM/X509 и асимметричного шифрования по схеме OAEP.
type RSAKeysHelper struct {
	bits int // Целевой размер ключа в битах (2048 или 4096)
}

// Гарантируем полное соответствие интерфейсам на этапе компиляции
var _ Keys[*rsa.PrivateKey, *rsa.PublicKey] = (*RSAKeysHelper)(nil)
var _ RSAKeys = (*RSAKeysHelper)(nil)

// NewRSAKeysHelper — фабричный конструктор хелпера ключей RSA.
func NewRSAKeysHelper(bits int) *RSAKeysHelper {
	return &RSAKeysHelper{
		bits: bits,
	}
}

// Generate создает новую пару ключей RSA заданной длины и кодирует результат
// в стандартные текстовые блоки PEM (PKCS#1 для приватного и PKIX для публичного ключа).
func (kh *RSAKeysHelper) Generate() (string, string, error) {
	// 1. Генерируем пару ключей с использованием системного генератора энтропии
	privateKey, err := rsa.GenerateKey(rand.Reader, kh.bits)
	if err != nil {
		return "", "", errs.NewUtlCipherError("generate private key", err)
	}

	// 2. Кодируем и маршалируем приватный ключ в стандарт PKCS#1 ASN.1 DER
	privateBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateBytes,
	})

	// 3. Кодируем и маршалируем публичный ключ в универсальный стандарт PKIX ASN.1 DER
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

// ParsePrivateKey парсит входящую текстовую строку PEM, извлекая структуру *rsa.PrivateKey (PKCS#1 спецификация).
func (kh *RSAKeysHelper) ParsePrivateKey(pemString string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		return nil, errs.NewUtlCipherError("failed decode PEM block", nil)
	}

	// Декодируем ASN.1 DER байты структуры PKCS#1
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errs.NewUtlCipherError("failed to parse PEM block", err)
	}

	return privateKey, nil
}

// ParsePublicKey парсит входящую текстовую строку PEM, извлекая структуру *rsa.PublicKey (PKIX спецификация).
func (kh *RSAKeysHelper) ParsePublicKey(pemString string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		return nil, errs.NewUtlCipherError("failed decode PEM block", nil)
	}

	// Парсим универсальный PKIX открытый ключ
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errs.NewUtlCipherError("failed to parse PEM block", err)
	}

	// Верифицируем, что извлеченный интерфейс скрывает именно структуру открытого ключа RSA
	switch pub := publicKey.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, errs.NewUtlCipherError("invalid RSA public key type", nil)
	}
}

// Decrypt осуществляет безопасное асимметричное дешифрование данных по схеме RSA-OAEP с хэшем SHA-256.
func (kh *RSAKeysHelper) Decrypt(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	res, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, data, nil)
	if err != nil {
		return nil, errs.NewUtlCipherError("decrypt error", err)
	}

	return res, nil
}

// DecryptString принимает зашифрованную Hex-строку, преобразует её в бинарный вид и производит дешифрование.
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

// Encrypt выполняет асимметричное шифрование данных на публичном ключе RSA с защитой от атак Блейхенбахера (OAEP).
func (kh *RSAKeysHelper) Encrypt(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	res, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, data, nil)
	if err != nil {
		return nil, errs.NewUtlCipherError("encrypt error", err)
	}

	return res, nil
}

// EncryptString шифрует строку текста и преобразует бинарный результат в печатный Hex-формат.
func (kh *RSAKeysHelper) EncryptString(data string, publicKey *rsa.PublicKey) (string, error) {
	res, err := kh.Encrypt([]byte(data), publicKey)
	if err != nil {
		return "", errs.NewUtlCipherError("encrypt string error", err)
	}

	return hex.EncodeToString(res), nil
}
