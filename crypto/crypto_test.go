package crypto_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/im10furry/hj1239-go-sdk/crypto"
)

func TestRegistryNoop(t *testing.T) {
	r := crypto.NewRegistry()
	data := []byte("hello world")

	enc, err := r.Encrypt(crypto.ModeNone, data)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if !bytes.Equal(enc, data) {
		t.Error("noop encrypt should return identical data")
	}

	dec, err := r.Decrypt(crypto.ModeNone, enc)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(dec, data) {
		t.Error("noop decrypt should return identical data")
	}
}

func TestAES128CBCRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef")
	iv := []byte("fedcba9876543210")

	e, err := crypto.NewAES128CBCEncryptor(key, iv)
	if err != nil {
		t.Fatalf("new aes: %v", err)
	}
	if e.Mode() != crypto.ModeAES128 {
		t.Errorf("mode: expected 0x05, got 0x%02x", e.Mode())
	}

	r := crypto.NewRegistry()
	r.Register(e)

	plaintext := []byte("HJ1239 test data for AES128 CBC encryption")
	ciphertext, err := r.Encrypt(crypto.ModeAES128, plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Error("ciphertext should differ from plaintext")
	}

	decrypted, err := r.Decrypt(crypto.ModeAES128, ciphertext)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("round-trip mismatch:\n  got:  %x\n  want: %x", decrypted, plaintext)
	}

	t.Logf("AES128-CBC: %d bytes -> %d bytes -> %d bytes", len(plaintext), len(ciphertext), len(decrypted))
}

func TestAES128CBCOneByte(t *testing.T) {
	e, _ := crypto.NewAES128CBCEncryptor([]byte("0123456789abcdef"), nil)
	r := crypto.NewRegistry()
	r.Register(e)

	plaintext := []byte{0x42}
	ciphertext, err := r.Encrypt(crypto.ModeAES128, plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	decrypted, err := r.Decrypt(crypto.ModeAES128, ciphertext)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("1-byte round-trip mismatch")
	}
}

func TestRSAEncryptWithPublicKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	r := crypto.NewRegistry()
	encryptor := crypto.NewRSAEncryptor(privateKey)
	r.Register(encryptor)

	plaintext := []byte("RSA OAEP test message")
	ciphertext, err := r.Encrypt(crypto.ModeRSA, plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Error("ciphertext should differ from plaintext")
	}

	decrypted, err := r.Decrypt(crypto.ModeRSA, ciphertext)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("RSA round-trip mismatch")
	}
	t.Logf("RSA-2048: %d bytes -> %d bytes -> %d bytes", len(plaintext), len(ciphertext), len(decrypted))
}

func TestAES128WithRandomIV(t *testing.T) {
	key := []byte("0123456789abcdef")
	e, err := crypto.NewAES128CBCWithRandomIV(key)
	if err != nil {
		t.Fatalf("new aes with random iv: %v", err)
	}
	iv := e.IV()
	if len(iv) != 16 {
		t.Errorf("iv length: expected 16, got %d", len(iv))
	}

	r := crypto.NewRegistry()
	r.Register(e)

	plaintext := []byte("random iv test")
	ciphertext, _ := r.Encrypt(crypto.ModeAES128, plaintext)
	decrypted, _ := r.Decrypt(crypto.ModeAES128, ciphertext)
	if !bytes.Equal(decrypted, plaintext) {
		t.Error("round-trip mismatch with random IV")
	}
}

func TestRegistryUnknownMode(t *testing.T) {
	r := crypto.NewRegistry()
	_, err := r.Encrypt(0x99, []byte("test"))
	if err == nil {
		t.Error("expected error for unknown mode")
	}
}
