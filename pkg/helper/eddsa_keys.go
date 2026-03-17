package helper

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/teserakt-io/golang-ed25519/extra25519"
	"golang.org/x/crypto/nacl/box"
)

type EdDSAKeys interface {
	Keys[ed25519.PrivateKey, ed25519.PublicKey]
}

type EdDSAKeysHelper struct {
	// Здесь не нужны биты, так как алгоритм фиксирован
}

func NewEdDSAKeysHelper() *EdDSAKeysHelper {
	return &EdDSAKeysHelper{}
}

func (kh *EdDSAKeysHelper) Generate() (string, string, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", errs.NewUtlCipherError("generate eddsa keys", err)
	}

	return hex.EncodeToString(privateKey), hex.EncodeToString(publicKey), nil
}

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

func (kh *EdDSAKeysHelper) Encrypt(data []byte, publicKey ed25519.PublicKey) ([]byte, error) {
	var publicCurve [32]byte
	var pubKey [32]byte
	copy(pubKey[:], publicKey)
	// Конвертация EdDSA -> Curve25519 (X25519)
	if !extra25519.PublicKeyToCurve25519(&publicCurve, &pubKey) {
		return nil, errs.NewUtlCipherError("convert pub key failed", nil)
	}

	out, err := box.SealAnonymous(nil, data, &publicCurve, rand.Reader)
	if err != nil {
		return nil, errs.NewUtlCipherError("encrypt error", err)
	}

	return out, nil
}

func (kh *EdDSAKeysHelper) Decrypt(data []byte, privateKey ed25519.PrivateKey) ([]byte, error) {
	var publicCurve [32]byte
	var privateCurve [32]byte
	var pubKey [32]byte
	var privKey [64]byte
	copy(pubKey[:], privateKey.Public().(ed25519.PublicKey))
	copy(privKey[:], privateKey)

	extra25519.PublicKeyToCurve25519(&publicCurve, &pubKey)
	extra25519.PrivateKeyToCurve25519(&privateCurve, &privKey)

	res, ok := box.OpenAnonymous(nil, data, &publicCurve, &privateCurve)
	if !ok {
		return nil, errs.NewUtlCipherError("decrypt error", nil)
	}

	return res, nil
}

func (kh *EdDSAKeysHelper) EncryptString(data string, publicKey ed25519.PublicKey) (string, error) {
	res, err := kh.Encrypt([]byte(data), publicKey)
	if err != nil {
		return "", errs.NewUtlCipherError("encrypt string error", err)
	}

	return hex.EncodeToString(res), nil
}

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
