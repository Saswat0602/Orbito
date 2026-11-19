package crypto

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// DefaultBcryptCost matches NestJS auth cost (12 rounds).
const DefaultBcryptCost = 12

// HashPassword generates a bcrypt hash of the password with cost 12.
func HashPassword(password string) (string, error) {
	return HashPasswordWithCost(password, DefaultBcryptCost)
}

// HashPasswordWithCost generates a bcrypt hash with the specified cost factor.
func HashPasswordWithCost(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a plaintext password with a bcrypt hash.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
