package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes sets up Chi routing for all auth module endpoints.
func Routes(h *Handler, authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	// Public routes
	r.Post("/login", h.Login)
	r.Post("/register", h.Register)
	r.Get("/registration-status", h.GetRegistrationStatus)
	r.Post("/refresh", h.RefreshToken)
	r.Post("/forgot-password", h.ForgotPassword)
	r.Get("/verify-reset-token/{token}", h.VerifyResetToken)
	r.Post("/reset-password", h.ResetPassword)
	r.Get("/setup/required", h.SetupRequired)
	r.Post("/setup", h.Setup)

	// SSO / OIDC routes
	r.Get("/oidc/config", h.OIDCConfig)
	r.Get("/oidc/login", h.OIDCLogin)
	r.Get("/oidc/callback", h.OIDCCallback)
	r.Post("/oidc/exchange", h.OIDCExchange)

	// Authenticated routes
	r.Group(func(sub chi.Router) {
		if authMiddleware != nil {
			sub.Use(authMiddleware)
		}
		sub.Post("/logout", h.Logout)
		sub.Get("/me", h.GetMe)
		sub.Get("/profile", h.GetMe)
	})

	return r
}
