package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/scrypt"
)

// FieldCrypto handles field-level encryption for sensitive integration tokens matching encryption.util.ts.
type FieldCrypto struct {
	secretKey []byte
}

// NewFieldCrypto derives a 32-byte AES key using scrypt matching Node.js crypto.scryptSync(encryptionKey, 'salt', 32).
func NewFieldCrypto(encryptionKey string) (*FieldCrypto, error) {
	if encryptionKey == "" {
		encryptionKey = "default-key-for-development-only"
	}

	// Node.js default scrypt parameters: N=16384, r=8, p=1, keyLen=32
	key, err := scrypt.Key([]byte(encryptionKey), []byte("salt"), 16384, 8, 1, 32)
	if err != nil {
		return nil, fmt.Errorf("scrypt key derivation failed: %w", err)
	}

	return &FieldCrypto{secretKey: key}, nil
}

// Encrypt encrypts a string using AES-256-GCM.
// Returns formatted string: ivHex:authTagHex:encryptedHex
func (fc *FieldCrypto) Encrypt(text string) (string, error) {
	if text == "" {
		return "", nil
	}

	block, err := aes.NewCipher(fc.secretKey)
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

	sealed := gcm.Seal(nil, iv, []byte(text), nil)
	if len(sealed) < tagLength {
		return "", errors.New("sealed text shorter than tag length")
	}

	ciphertext := sealed[:len(sealed)-tagLength]
	authTag := sealed[len(sealed)-tagLength:]

	return fmt.Sprintf("%x:%x:%x", iv, authTag, ciphertext), nil
}

// Decrypt decrypts an AES-256-GCM string formatted as ivHex:authTagHex:encryptedHex.
// If the input is not in the expected format or decryption fails, it returns the original string
// matching the fallback behavior in NestJS encryption.util.ts.
func (fc *FieldCrypto) Decrypt(encryptedText string) string {
	if encryptedText == "" {
		return ""
	}

	parts := strings.Split(encryptedText, ":")
	if len(parts) != 3 {
		return encryptedText
	}

	ivHex, authTagHex, encryptedHex := parts[0], parts[1], parts[2]

	iv, err := hex.DecodeString(ivHex)
	if err != nil || len(iv) != ivLength {
		return encryptedText
	}

	authTag, err := hex.DecodeString(authTagHex)
	if err != nil || len(authTag) != tagLength {
		return encryptedText
	}

	ciphertext, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return encryptedText
	}

	block, err := aes.NewCipher(fc.secretKey)
	if err != nil {
		return encryptedText
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return encryptedText
	}

	combined := append(ciphertext, authTag...)
	plaintext, err := gcm.Open(nil, iv, combined, nil)
	if err != nil {
		return encryptedText
	}

	return string(plaintext)
}
