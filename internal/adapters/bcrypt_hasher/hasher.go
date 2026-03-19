package bcrypthasher

import (
	"github.com/DEEJ4Y/genkitkraft/internal/ports/hasher"
	"golang.org/x/crypto/bcrypt"
)

// Compile-time check that BcryptHasher implements hasher.PasswordHasher.
var _ hasher.PasswordHasher = (*BcryptHasher)(nil)

// BcryptHasher implements password hashing using bcrypt.
type BcryptHasher struct{}

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

func (h *BcryptHasher) Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func (h *BcryptHasher) Compare(hashedPassword []byte, plainPassword string) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(plainPassword))
}
