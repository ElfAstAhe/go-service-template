package helper

type Keys[Priv any, Pub any] interface {
	Generate() (string, string, error)
	ParsePrivateKey(string) (Priv, error)
	ParsePublicKey(string) (Pub, error)
	Decrypt(data []byte, privateKey Priv) ([]byte, error)
	DecryptString(data string, privateKey Priv) (string, error)
	Encrypt(data []byte, publicKey Pub) ([]byte, error)
	EncryptString(data string, publicKey Pub) (string, error)
}
