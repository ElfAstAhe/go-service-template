package helper

// Keys описывает универсальный, строго типизированный (Generic) контракт для работы
// с парами асимметричных криптографических ключей (Asymmetric Cryptography Contract).
//
// Абстрагирует фреймворк от конкретных математических алгоритмов (Ed25519, RSA, ECDSA),
// инкапсулируя рутину генерации, парсинга из Hex-строк, а также асимметричного шифрования и дешифрования.
type Keys[Priv any, Pub any] interface {
	// Generate создает новую случайную пару ключей, возвращая приватный и публичный ключи в виде Hex-строк.
	Generate() (privateKeyStr string, publicKeyStr string, err error)

	// ParsePrivateKey десериализует Hex-строку в нативный объект приватного ключа Priv.
	ParsePrivateKey(hexString string) (Priv, error)

	// ParsePublicKey десериализует Hex-строку в нативный объект публичного ключа Pub.
	ParsePublicKey(hexString string) (Pub, error)

	// Decrypt производит дешифрование сырого бинарного массива байт с использованием приватного ключа Priv.
	Decrypt(data []byte, privateKey Priv) (plaintext []byte, err error)

	// DecryptString принимает зашифрованную Hex-строку и дешифрует её в исходный текст с помощью приватного ключа Priv.
	DecryptString(data string, privateKey Priv) (plaintext string, err error)

	// Encrypt выполняет асимметричное шифрование массива байт на публичном ключе Pub.
	Encrypt(data []byte, publicKey Pub) (ciphertext []byte, err error)

	// EncryptString шифрует текстовую строку на публичном ключе Pub, возвращая печатный Hex-результат.
	EncryptString(data string, publicKey Pub) (ciphertext string, err error)
}
