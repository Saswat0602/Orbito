package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/crypto"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP ingress for all authentication and session endpoints.
type Handler struct {
	service      *Service
	cookieSecret string
	isProd       bool
	frontendURL  string
}

// NewHandler constructs a new auth handler.
func NewHandler(service *Service, cookieSecret string, isProd bool, frontendURL string) *Handler {
	if frontendURL == "" {
		frontendURL = "http://localhost:3001"
	}
	return &Handler{
		service:      service,
		cookieSecret: cookieSecret,
		isProd:       isProd,
		frontendURL:  frontendURL,
	}
}

// Login handles POST /api/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var dto LoginDTO
	if err := httpx.BindJSON(r, &dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	res, err := h.service.Login(r.Context(), dto)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	SetAuthCookies(w, res.AccessToken, res.RefreshToken, h.isProd, h.cookieSecret)
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Register handles POST /api/auth/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var dto RegisterDTO
	if err := httpx.BindJSON(r, &dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	res, err := h.service.Register(r.Context(), dto)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	SetAuthCookies(w, res.AccessToken, res.RefreshToken, h.isProd, h.cookieSecret)
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// GetRegistrationStatus handles GET /api/auth/registration-status.
func (h *Handler) GetRegistrationStatus(w http.ResponseWriter, r *http.Request) {
	enabled := true
	allowOrg := true
	hasDefaultOrg := false

	if h.service.settingsSvc != nil {
		regStatus, err := h.service.settingsSvc.GetRegistrationStatus(r.Context())
		if err == nil {
			enabled = regStatus.Enabled
			allowOrg = regStatus.AllowOrganizations
			hasDefaultOrg = regStatus.DefaultOrgID != nil
		}
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"enabled":                enabled,
		"allowOrgCreation":       allowOrg,
		"hasDefaultOrganization": hasDefaultOrg,
	})
}

// RefreshToken handles POST /api/auth/refresh.
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	token := ""

	// Read refresh token from encrypted httpOnly cookie first
	if cookie, err := r.Cookie("refresh_token"); err == nil && cookie.Value != "" {
		if h.cookieSecret != "" {
			decrypted, decErr := crypto.DecryptCookieValue(cookie.Value, h.cookieSecret)
			if decErr == nil && decrypted != "" {
				token = decrypted
			}
		}
		if token == "" {
			token = cookie.Value
		}
	}

	// Fallback to POST body
	if token == "" {
		var dto RefreshTokenDTO
		_ = httpx.BindJSON(r, &dto)
		token = strings.TrimSpace(dto.RefreshToken)
	}

	if token == "" {
		httpx.RespondError(w, r, apperror.Unauthorized("No refresh token provided"))
		return
	}

	res, err := h.service.RefreshToken(r.Context(), token)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	SetAuthCookies(w, res.AccessToken, res.RefreshToken, h.isProd, h.cookieSecret)
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Logout handles POST /api/auth/logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userIDStr := ctxhelper.GetUserID(r.Context())
	if userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			_ = h.service.Logout(r.Context(), uid)
		}
	}

	ClearAuthCookies(w, h.isProd)
	httpx.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Logout successful",
	})
}

// GetMe handles GET /api/auth/me and GET /api/auth/profile.
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userIDStr := ctxhelper.GetUserID(r.Context())
	if userIDStr == "" {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthorized"))
		return
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Invalid user identity"))
		return
	}

	user, err := h.service.GetUserByID(r.Context(), uid)
	if err != nil || user == nil {
		httpx.RespondError(w, r, apperror.NotFound("User not found"))
		return
	}

	var uname, avatar, bio, mobile, lang, tz *string
	if user.Username.Valid {
		uname = &user.Username.String
	}
	if user.Avatar.Valid {
		avatar = &user.Avatar.String
	}
	if user.Bio.Valid {
		bio = &user.Bio.String
	}
	if user.MobileNumber.Valid {
		mobile = &user.MobileNumber.String
	}
	if user.Language != "" {
		lang = &user.Language
	}
	if user.Timezone != "" {
		tz = &user.Timezone
	}

	httpx.RespondJSON(w, http.StatusOK, UserDTO{
		ID:           user.ID.String(),
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Username:     uname,
		Role:         string(user.Role),
		Avatar:       avatar,
		Bio:          bio,
		MobileNumber: mobile,
		Language:     lang,
		Timezone:     tz,
	})
}

// ForgotPassword handles POST /api/auth/forgot-password.
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var dto ForgotPasswordDTO
	if err := httpx.BindJSON(r, &dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	if err := h.service.ForgotPassword(r.Context(), dto.Email); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Password reset instructions sent to your email",
	})
}

// VerifyResetToken handles GET /api/auth/verify-reset-token/{token}.
func (h *Handler) VerifyResetToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		httpx.RespondError(w, r, apperror.BadRequest("Token is required"))
		return
	}

	isValid, err := h.service.VerifyResetToken(r.Context(), token)
	if err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to verify token", err))
		return
	}

	msg := "Invalid or expired token"
	if isValid {
		msg = "Token is valid"
	}

	httpx.RespondJSON(w, http.StatusOK, VerifyResetTokenResponseDTO{
		Valid:   isValid,
		Message: msg,
	})
}

// ResetPassword handles POST /api/auth/reset-password.
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var dto ResetPasswordDTO
	if err := httpx.BindJSON(r, &dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	if err := h.service.ResetPassword(r.Context(), dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Password has been reset successfully",
	})
}

// SetupRequired handles GET /api/auth/setup/required.
func (h *Handler) SetupRequired(w http.ResponseWriter, r *http.Request) {
	required, canSetup, msg, err := h.service.IsSetupRequired(r.Context())
	if err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to check setup requirement", err))
		return
	}

	httpx.RespondJSON(w, http.StatusOK, SetupRequiredResponseDTO{
		Required: required,
		CanSetup: canSetup,
		Message:  msg,
	})
}

// Setup handles POST /api/auth/setup.
func (h *Handler) Setup(w http.ResponseWriter, r *http.Request) {
	var dto SetupAdminDTO
	if err := httpx.BindJSON(r, &dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	res, err := h.service.SetupSuperAdmin(r.Context(), dto)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	SetAuthCookies(w, res.AccessToken, res.RefreshToken, h.isProd, h.cookieSecret)
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// OIDCConfig handles GET /api/auth/oidc/config.
func (h *Handler) OIDCConfig(w http.ResponseWriter, r *http.Request) {
	config := h.service.GetOIDCConfig(r.Context())
	httpx.RespondJSON(w, http.StatusOK, config)
}

// OIDCLogin handles GET /api/auth/oidc/login.
func (h *Handler) OIDCLogin(w http.ResponseWriter, r *http.Request) {
	// If SSO provider is not configured, redirect back to login with error
	target := h.frontendURL + "/login?error=sso_not_configured"
	http.Redirect(w, r, target, http.StatusFound)
}

// OIDCCallback handles GET /api/auth/oidc/callback.
func (h *Handler) OIDCCallback(w http.ResponseWriter, r *http.Request) {
	target := h.frontendURL + "/login?error=sso_invalid_state"
	http.Redirect(w, r, target, http.StatusFound)
}

// OIDCExchange handles POST /api/auth/oidc/exchange.
func (h *Handler) OIDCExchange(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sso_auth")
	if err != nil || cookie.Value == "" {
		httpx.RespondError(w, r, apperror.Unauthorized("No SSO session found"))
		return
	}

	decrypted := cookie.Value
	if h.cookieSecret != "" {
		dec, decErr := crypto.DecryptCookieValue(cookie.Value, h.cookieSecret)
		if decErr == nil && dec != "" {
			decrypted = dec
		}
	}

	var result ResponseDTO
	if err := json.Unmarshal([]byte(decrypted), &result); err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Invalid SSO session"))
		return
	}

	SetAuthCookies(w, result.AccessToken, result.RefreshToken, h.isProd, h.cookieSecret)
	httpx.RespondJSON(w, http.StatusOK, result)
}
