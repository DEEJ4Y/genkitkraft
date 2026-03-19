package aesgcmencryptor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/DEEJ4Y/genkitkraft/internal/ports/encryptor"
)

// Compile-time check that AESGCMEncryptor implements encryptor.Encryptor.
var _ encryptor.Encryptor = (*AESGCMEncryptor)(nil)

// AESGCMEncryptor implements symmetric encryption using AES-256-GCM.
type AESGCMEncryptor struct {
	aead cipher.AEAD
}

// NewAESGCMEncryptor creates an encryptor from an arbitrary key string.
// The key is hashed with SHA-256 to derive a 32-byte AES-256 key.
func NewAESGCMEncryptor(key string) (*AESGCMEncryptor, error) {
	hash := sha256.Sum256([]byte(key))

	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, fmt.Errorf("creating AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	return &AESGCMEncryptor{aead: aead}, nil
}

func (e *AESGCMEncryptor) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}

	ciphertext := e.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *AESGCMEncryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("decoding base64: %w", err)
	}

	nonceSize := e.aead.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, encrypted := data[:nonceSize], data[nonceSize:]
	plaintext, err := e.aead.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", fmt.Errorf("decrypting: %w", err)
	}

	return string(plaintext), nil
}
