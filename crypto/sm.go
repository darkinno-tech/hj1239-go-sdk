package crypto

// SM2 and SM4 are Chinese national cryptographic standards.
// Go stdlib does not include SM2/SM4 implementations.
// To enable SM2/SM4 encryption, implement the Encryptor interface
// using github.com/tjfoc/gmsm or another SM library.
//
// Example:
//
//	type SM4CBCEncryptor struct { key [16]byte }
//	func (s *SM4CBCEncryptor) Mode() byte { return crypto.ModeSM4 }
//	func (s *SM4CBCEncryptor) Encrypt(p []byte) ([]byte, error) { ... }
//	func (s *SM4CBCEncryptor) Decrypt(c []byte) ([]byte, error) { ... }
//
// Then register with registry.Register(&SM4CBCEncryptor{...})
//
// The crypto.Registry will route EncryptMode 0x02 (SM2) and 0x03 (SM4)
// to the registered implementations.
