package crypto

import "fmt"

const (
	ModeNone   byte = 0x01
	ModeSM2    byte = 0x02
	ModeSM4    byte = 0x03
	ModeRSA    byte = 0x04
	ModeAES128 byte = 0x05
)

type Encryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
	Mode() byte
}

type Registry struct {
	encryptors map[byte]Encryptor
}

func NewRegistry() *Registry {
	r := &Registry{encryptors: make(map[byte]Encryptor)}
	r.Register(&NoopEncryptor{})
	return r
}

func (r *Registry) Register(e Encryptor) {
	r.encryptors[e.Mode()] = e
}

func (r *Registry) Encrypt(mode byte, plaintext []byte) ([]byte, error) {
	if mode == ModeNone {
		return plaintext, nil
	}
	e, ok := r.encryptors[mode]
	if !ok {
		return nil, fmt.Errorf("crypto: no encryptor registered for mode 0x%02x", mode)
	}
	return e.Encrypt(plaintext)
}

func (r *Registry) Decrypt(mode byte, ciphertext []byte) ([]byte, error) {
	if mode == ModeNone {
		return ciphertext, nil
	}
	e, ok := r.encryptors[mode]
	if !ok {
		return nil, fmt.Errorf("crypto: no encryptor registered for mode 0x%02x", mode)
	}
	return e.Decrypt(ciphertext)
}

type NoopEncryptor struct{}

func (n *NoopEncryptor) Encrypt(p []byte) ([]byte, error) { return p, nil }
func (n *NoopEncryptor) Decrypt(c []byte) ([]byte, error) { return c, nil }
func (n *NoopEncryptor) Mode() byte                       { return ModeNone }
