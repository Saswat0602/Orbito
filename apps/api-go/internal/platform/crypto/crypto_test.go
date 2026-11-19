package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestCookieCrypto(t *testing.T) {
	secret := "orbito-jwt-super-secret-development-key"

	t.Run("round-trip encryption and decryption", func(t *testing.T) {
		original := "session-refresh-token-xyz-12345"
		encrypted, err := EncryptCookieValue(original, secret)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted)

		decrypted, err := DecryptCookieValue(encrypted, secret)
		require.NoError(t, err)
		assert.Equal(t, original, decrypted)
	})

	t.Run("decrypts Node.js generated test vector", func(t *testing.T) {
		nodeVector := "0102030405060708090a0b0c:c848cdc73067013507626c93a7570c8d:5657bc2c407485b87c73234b5be9608419b4e45e75289377"
		expectedPlaintext := "user-session-token-12345"

		decrypted, err := DecryptCookieValue(nodeVector, secret)
		require.NoError(t, err)
		assert.Equal(t, expectedPlaintext, decrypted)
	})

	t.Run("fails on tampered ciphertext or wrong secret", func(t *testing.T) {
		original := "my-secret-token"
		encrypted, err := EncryptCookieValue(original, secret)
		require.NoError(t, err)

		// Wrong secret
		_, err = DecryptCookieValue(encrypted, "wrong-secret-key")
		assert.Error(t, err)

		// Corrupted format
		_, err = DecryptCookieValue("invalid:cookie", secret)
		assert.Error(t, err)

		// Corrupted tag
		tampered := encrypted[:len(encrypted)-4] + "dead"
		_, err = DecryptCookieValue(tampered, secret)
		assert.Error(t, err)
	})
}

func TestFieldCrypto(t *testing.T) {
	encKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	fc, err := NewFieldCrypto(encKey)
	require.NoError(t, err)

	t.Run("round-trip encryption and decryption", func(t *testing.T) {
		original := "ghp_PersonalAccessTokenSecret123"
		encrypted, err := fc.Encrypt(original)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted)

		decrypted := fc.Decrypt(encrypted)
		assert.Equal(t, original, decrypted)
	})

	t.Run("decrypts Node.js generated test vector", func(t *testing.T) {
		nodeVector := "112233445566778899aabbcc:f50f629d22c17fd15657c3d8a4dc873e:a5bfb8e7972ae440dc51b64e605d0a05d9b500ad19d341c72e"
		expectedPlaintext := "jira-api-token-secret-987"

		decrypted := fc.Decrypt(nodeVector)
		assert.Equal(t, expectedPlaintext, decrypted)
	})

	t.Run("fallback returns original text on non-encrypted or failed input", func(t *testing.T) {
		// Non-encrypted string (legacy data)
		legacy := "plain-unencrypted-token"
		assert.Equal(t, legacy, fc.Decrypt(legacy))

		// Empty string
		assert.Equal(t, "", fc.Decrypt(""))

		// Corrupted encrypted string
		corrupted := "112233445566778899aabbcc:badtagbadtagbadtagbadtagbadtagba:123456"
		assert.Equal(t, corrupted, fc.Decrypt(corrupted))
	})
}

func TestPasswordBcrypt(t *testing.T) {
	password := "SecurePassword123!"

	t.Run("hashes with cost 12 and verifies", func(t *testing.T) {
		hash, err := HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)

		// Verify cost
		cost, err := bcrypt.Cost([]byte(hash))
		require.NoError(t, err)
		assert.Equal(t, 12, cost)

		// Verify match
		assert.True(t, CheckPasswordHash(password, hash))
		assert.False(t, CheckPasswordHash("WrongPassword999!", hash))
	})

	t.Run("verifies Node.js bcrypt hash", func(t *testing.T) {
		nodeHash := "$2b$12$k5EL.mvCnWEA.OOSOqP.SeiEYteOW7LjMtJq6cQe9RnAJ6jUvZcdS"
		assert.True(t, CheckPasswordHash(password, nodeHash))
		assert.False(t, CheckPasswordHash("incorrect-password", nodeHash))
	})
}
