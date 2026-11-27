package events

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

type handshakeMeta struct {
	queryToken  string
	bearerToken string
	cookieToken string
}

func extractHandshakeMeta(r *http.Request) handshakeMeta {
	meta := handshakeMeta{
		queryToken: strings.TrimSpace(r.URL.Query().Get("token")),
	}

	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			meta.bearerToken = strings.TrimSpace(parts[1])
		}
	}

	if cookie, err := r.Cookie("access_token"); err == nil && cookie.Value != "" {
		decoded, decErr := url.QueryUnescape(cookie.Value)
		if decErr == nil {
			meta.cookieToken = strings.TrimSpace(decoded)
		} else {
			meta.cookieToken = strings.TrimSpace(cookie.Value)
		}
	}

	return meta
}

func resolveToken(authPayloadToken string, meta handshakeMeta) string {
	if tok := strings.TrimSpace(authPayloadToken); tok != "" {
		return tok
	}
	if meta.queryToken != "" {
		return meta.queryToken
	}
	if meta.bearerToken != "" {
		return meta.bearerToken
	}
	return meta.cookieToken
}

func (g *Gateway) authenticateToken(ctx context.Context, tokenStr string) (string, error) {
	if tokenStr == "" {
		return "", errors.New("Authentication token required")
	}
	if g.tokenValidator == nil {
		return "", errors.New("Authentication validator unavailable")
	}

	sub, _, _, err := g.tokenValidator.ValidateToken(tokenStr)
	if err != nil {
		return "", errors.New("Invalid or expired token")
	}
	if sub == "" {
		return "", errors.New("Invalid token payload")
	}

	userUUID, parseErr := uuid.Parse(sub)
	if parseErr != nil {
		return "", errors.New("Invalid token payload")
	}
	if userUUID == middleware.SystemUserID {
		return "", errors.New("Invalid user account")
	}

	if g.userChecker != nil {
		account, checkErr := g.userChecker(ctx, userUUID)
		if checkErr != nil || account == nil {
			return "", errors.New("User account not found")
		}
		if account.Status != sqlc.UserStatusACTIVE || account.DeletedAt.Valid {
			return "", errors.New("Account is not active")
		}
	}

	return sub, nil
}
