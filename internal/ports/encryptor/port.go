package encryptor

// Encryptor defines the contract for symmetric encryption and decryption.
type Encryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}
