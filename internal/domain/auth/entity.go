package auth

// User represents an authenticated user with bcrypt-hashed password.
type User struct {
	Username     string
	PasswordHash []byte
}
