package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	ivLength  = 12
	tagLength = 16
)

// Common cookie encryption and decryption errors.
var (
	// ErrInvalidCookie indicates the encrypted cookie payload has invalid parts or length.
	ErrInvalidCookie = errors.New("invalid encrypted cookie format or length")
	// ErrDecryption indicates AEAD authentication or decryption failed.
	ErrDecryption = errors.New("cookie decryption failed")
)

func deriveCookieKey(secret string) []byte {
	hash := sha256.Sum256([]byte(secret))
	return hash[:]
}

// EncryptCookieValue encrypts a plaintext string using AES-256-GCM matching NestJS cookie-crypto.utils.ts.
// Output format: ivHex:tagHex:cipherHex
func EncryptCookieValue(plaintext, secret string) (string, error) {
	key := deriveCookieKey(secret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	iv := make([]byte, ivLength)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	// Go's gcm.Seal appends the 16-byte authentication tag to the ciphertext
	sealed := gcm.Seal(nil, iv, []byte(plaintext), nil)
	if len(sealed) < tagLength {
		return "", errors.New("sealed ciphertext shorter than tag length")
	}

	ciphertext := sealed[:len(sealed)-tagLength]
	tag := sealed[len(sealed)-tagLength:]

	return fmt.Sprintf("%x:%x:%x", iv, tag, ciphertext), nil
}

// DecryptCookieValue decrypts an AES-256-GCM string in the format ivHex:tagHex:cipherHex matching NestJS.
func DecryptCookieValue(encryptedValue, secret string) (string, error) {
	parts := strings.Split(encryptedValue, ":")
	if len(parts) != 3 {
		return "", ErrInvalidCookie
	}

	ivHex, tagHex, cipherHex := parts[0], parts[1], parts[2]

	iv, err := hex.DecodeString(ivHex)
	if err != nil || len(iv) != ivLength {
		return "", ErrInvalidCookie
	}

	tag, err := hex.DecodeString(tagHex)
	if err != nil || len(tag) != tagLength {
		return "", ErrInvalidCookie
	}

	ciphertext, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", ErrInvalidCookie
	}

	key := deriveCookieKey(secret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Go's gcm.Open expects [ciphertext... | tag...]
	combined := append(ciphertext, tag...)
	plaintext, err := gcm.Open(nil, iv, combined, nil)
	if err != nil {
		return "", ErrDecryption
	}

	return string(plaintext), nil
}
