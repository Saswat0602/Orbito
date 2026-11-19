package auth

import (
	"net/http"

	"github.com/orbito/api-go/internal/platform/crypto"
)

// SetAuthCookies writes access_token and encrypted refresh_token httpOnly cookies.
func SetAuthCookies(w http.ResponseWriter, accessToken, refreshToken string, isProd bool, cookieSecret string) {
	encryptedRefresh := refreshToken
	if cookieSecret != "" {
		enc, err := crypto.EncryptCookieValue(refreshToken, cookieSecret)
		if err == nil {
			encryptedRefresh = enc
		}
	}

	sameSite := http.SameSiteLaxMode

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    encryptedRefresh,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		HttpOnly: true,
		Secure:   isProd,
		SameSite: sameSite,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   60 * 60, // 1 hour
		HttpOnly: true,
		Secure:   isProd,
		SameSite: sameSite,
	})
}

// ClearAuthCookies expires both access_token and refresh_token cookies.
func ClearAuthCookies(w http.ResponseWriter, isProd bool) {
	sameSite := http.SameSiteLaxMode

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProd,
		SameSite: sameSite,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProd,
		SameSite: sameSite,
	})
}
