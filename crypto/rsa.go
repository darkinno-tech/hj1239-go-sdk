package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
)

type RSAEncryptor struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewRSAEncryptor(privateKey *rsa.PrivateKey) *RSAEncryptor {
	return &RSAEncryptor{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}
}

func NewRSAEncryptorFromPublicKey(publicKey *rsa.PublicKey) *RSAEncryptor {
	return &RSAEncryptor{publicKey: publicKey}
}

func (e *RSAEncryptor) Mode() byte { return ModeRSA }

func (e *RSAEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	if e.publicKey == nil {
		return nil, fmt.Errorf("rsa: no public key available for encryption")
	}
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, e.publicKey, plaintext, nil)
}

func (e *RSAEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	if e.privateKey == nil {
		return nil, fmt.Errorf("rsa: no private key available for decryption")
	}
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, e.privateKey, ciphertext, nil)
}
