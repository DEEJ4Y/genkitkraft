package hasher

// PasswordHasher defines the contract for password hashing and comparison.
type PasswordHasher interface {
	Hash(password string) ([]byte, error)
	Compare(hashedPassword []byte, plainPassword string) error
}
