package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type AES128CBCEncryptor struct {
	key [16]byte
	iv  [16]byte
}

func NewAES128CBCEncryptor(key, iv []byte) (*AES128CBCEncryptor, error) {
	e := &AES128CBCEncryptor{}
	if len(key) != 16 {
		return nil, fmt.Errorf("aes128: key must be 16 bytes, got %d", len(key))
	}
	copy(e.key[:], key)
	if len(iv) == 16 {
		copy(e.iv[:], iv)
	} else {
		copy(e.iv[:], key)
	}
	return e, nil
}

func NewAES128CBCWithRandomIV(key []byte) (*AES128CBCEncryptor, error) {
	e := &AES128CBCEncryptor{}
	if len(key) != 16 {
		return nil, fmt.Errorf("aes128: key must be 16 bytes, got %d", len(key))
	}
	copy(e.key[:], key)
	if _, err := io.ReadFull(rand.Reader, e.iv[:]); err != nil {
		return nil, fmt.Errorf("aes128: generate IV: %w", err)
	}
	return e, nil
}

func (e *AES128CBCEncryptor) Mode() byte { return ModeAES128 }

func (e *AES128CBCEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key[:])
	if err != nil {
		return nil, err
	}
	padded := pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, e.iv[:])
	mode.CryptBlocks(ciphertext, padded)
	return ciphertext, nil
}

func (e *AES128CBCEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key[:])
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < aes.BlockSize || len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("aes128: ciphertext length %d is not a multiple of block size", len(ciphertext))
	}
	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, e.iv[:])
	mode.CryptBlocks(plaintext, ciphertext)
	return pkcs7Unpad(plaintext)
}

func (e *AES128CBCEncryptor) IV() []byte {
	v := make([]byte, 16)
	copy(v, e.iv[:])
	return v
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)%blockSize
	pad := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(data, pad...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("pkcs7: empty data")
	}
	padLen := int(data[len(data)-1])
	if padLen > len(data) || padLen == 0 {
		return nil, fmt.Errorf("pkcs7: invalid padding")
	}
	for i := 0; i < padLen; i++ {
		if data[len(data)-1-i] != byte(padLen) {
			return nil, fmt.Errorf("pkcs7: invalid padding")
		}
	}
	return data[:len(data)-padLen], nil
}
