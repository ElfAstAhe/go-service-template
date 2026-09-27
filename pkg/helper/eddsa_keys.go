package helper

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/teserakt-io/golang-ed25519/extra25519"
	"golang.org/x/crypto/nacl/box"
)

// EdDSAKeys расширяет базовый дженерик-интерфейс Keys для фиксированной пары Ed25519.
type EdDSAKeys interface {
	Keys[ed25519.PrivateKey, ed25519.PublicKey]
}

// EdDSAKeysHelper реализует интерфейс EdDSAKeys, инкапсулируя логику генерации,
// парсинга и асимметричного шифрования на базе эллиптических кривых Ed25519.
type EdDSAKeysHelper struct {
	// Здесь не нужны биты, так как алгоритм фиксирован стандартами Ed25519
}

// Гарантируем полное соответствие интерфейсам на этапе компиляции
var _ Keys[ed25519.PrivateKey, ed25519.PublicKey] = (*EdDSAKeysHelper)(nil)
var _ EdDSAKeys = (*EdDSAKeysHelper)(nil)

// NewEdDSAKeysHelper — фабричный конструктор хелпера EdDSA ключей.
func NewEdDSAKeysHelper() *EdDSAKeysHelper {
	return &EdDSAKeysHelper{}
}

// Generate создает новую пару криптографических ключей Ed25519,
// возвращая приватный и публичный ключи в формате печатных Hex-строк.
func (kh *EdDSAKeysHelper) Generate() (string, string, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", errs.NewUtlCipherError("generate eddsa keys", err)
	}

	return hex.EncodeToString(privateKey), hex.EncodeToString(publicKey), nil
}

// ParsePrivateKey десериализует Hex-строку в нативный тип ed25519.PrivateKey с проверкой размера пакета.
func (kh *EdDSAKeysHelper) ParsePrivateKey(hexString string) (ed25519.PrivateKey, error) {
	data, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, errs.NewUtlCipherError("failed decode hex private key", err)
	}
	if len(data) != ed25519.PrivateKeySize {
		return nil, errs.NewUtlCipherError("invalid eddsa private key length", nil)
	}

	return ed25519.PrivateKey(data), nil
}

// ParsePublicKey de-сериализует Hex-строку в нативный тип ed25519.PublicKey.
func (kh *EdDSAKeysHelper) ParsePublicKey(hexString string) (ed25519.PublicKey, error) {
	data, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, errs.NewUtlCipherError("failed decode hex public key", err)
	}
	if len(data) != ed25519.PublicKeySize {
		return nil, errs.NewUtlCipherError("invalid eddsa public key length", nil)
	}

	return ed25519.PublicKey(data), nil
}

// Encrypt выполняет асимметричное анонимное шифрование (Sealed Box) массива байт.
// 💡 Криптографический нюанс: Метод прозрачно конвертирует публичный ключ Ed25519
// в формат Curve25519 (X25519) для совместимости со схемой шифрования NaCL Box.
func (kh *EdDSAKeysHelper) Encrypt(data []byte, publicKey ed25519.PublicKey) ([]byte, error) {
	var publicCurve [32]byte
	var pubKey [32]byte
	copy(pubKey[:], publicKey)

	// Конвертация EdDSA -> Curve25519 (X25519)
	if !extra25519.PublicKeyToCurve25519(&publicCurve, &pubKey) {
		return nil, errs.NewUtlCipherError("convert pub key failed", nil)
	}

	// Анонимное шифрование генерирует одноразовый ключ внутри, исключая утечку метаданных об отправителе
	out, err := box.SealAnonymous(nil, data, &publicCurve, rand.Reader)
	if err != nil {
		return nil, errs.NewUtlCipherError("encrypt error", err)
	}

	return out, nil
}

// Decrypt осуществляет асимметричное дешифрование анонимного Sealed Box контейнера.
// Выполняет одновременную конвертацию приватного и публичного плеча ключей в кривую Curve25519.
func (kh *EdDSAKeysHelper) Decrypt(data []byte, privateKey ed25519.PrivateKey) ([]byte, error) {
	var publicCurve [32]byte
	var privateCurve [32]byte
	var pubKey [32]byte
	var privKey [64]byte

	// Извлекаем публичное плечо из приватного ключа и копируем байты в массивы фиксированной длины
	copy(pubKey[:], privateKey.Public().(ed25519.PublicKey))
	copy(privKey[:], privateKey)

	extra25519.PublicKeyToCurve25519(&publicCurve, &pubKey)
	extra25519.PrivateKeyToCurve25519(&privateCurve, &privKey)

	// Декодируем пакет. Метод вернет false (ok == false), если данные или тег подлинности изменены
	res, ok := box.OpenAnonymous(nil, data, &publicCurve, &privateCurve)
	if !ok {
		return nil, errs.NewUtlCipherError("decrypt error / integrity or key verification failed", nil)
	}

	return res, nil
}

// EncryptString шифрует текстовую строку на публичном ключе и возвращает печатный Hex-результат.
func (kh *EdDSAKeysHelper) EncryptString(data string, publicKey ed25519.PublicKey) (string, error) {
	res, err := kh.Encrypt([]byte(data), publicKey)
	if err != nil {
		return "", errs.NewUtlCipherError("encrypt string error", err)
	}

	return hex.EncodeToString(res), nil
}

// DecryptString принимает Hex-строку, выполняет дешифрование на приватном ключе и возвращает исходный текст.
func (kh *EdDSAKeysHelper) DecryptString(data string, privateKey ed25519.PrivateKey) (string, error) {
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
