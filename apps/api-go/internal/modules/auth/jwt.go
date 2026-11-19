package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// SystemUserID is the special system account barred from authenticating.
var SystemUserID = uuid.MustParse("00000000-0000-0000-0000-000000000000")

// Claims represents the standard JWT payload emitted and validated across Orbito.
type Claims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// TokenService manages minting and parsing JWT access and refresh tokens.
type TokenService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// NewTokenService constructs a new TokenService.
func NewTokenService(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *TokenService {
	if accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &TokenService{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

// MintAccessToken creates a signed HS256 access token.
func (s *TokenService) MintAccessToken(userID uuid.UUID, email, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		Sub:   userID.String(),
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.accessSecret)
}

// MintRefreshToken creates a signed HS256 refresh token.
func (s *TokenService) MintRefreshToken(userID uuid.UUID, email, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		Sub:   userID.String(),
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.refreshSecret)
}

// ValidateAccessToken parses and validates an access token.
func (s *TokenService) ValidateAccessToken(tokenStr string) (*Claims, error) {
	return parseToken(tokenStr, s.accessSecret)
}

// ValidateToken satisfies middleware.TokenValidator for access tokens.
func (s *TokenService) ValidateToken(tokenStr string) (string, string, string, error) {
	claims, err := s.ValidateAccessToken(tokenStr)
	if err != nil {
		return "", "", "", err
	}
	return claims.Sub, claims.Email, claims.Role, nil
}

// ValidateRefreshToken parses and validates a refresh token.
func (s *TokenService) ValidateRefreshToken(tokenStr string) (*Claims, error) {
	return parseToken(tokenStr, s.refreshSecret)
}

func parseToken(tokenStr string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
