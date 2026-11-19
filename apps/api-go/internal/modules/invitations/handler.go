package invitations

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler processes HTTP requests for the invitations module.
type Handler struct {
	service Service
}

// NewHandler constructs an invitations HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// Create handles POST /api/invitations.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	var req CreateInvitationRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.CreateInvitation(r.Context(), req, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, res)
}

// Accept handles PATCH & POST /api/invitations/{token}/accept.
func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		httpx.RespondError(w, r, apperror.BadRequest("Invitation token is required"))
		return
	}

	res, err := h.service.AcceptInvitation(r.Context(), token, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// Decline handles PATCH /api/invitations/{token}/decline.
func (h *Handler) Decline(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		httpx.RespondError(w, r, apperror.BadRequest("Invitation token is required"))
		return
	}

	res, err := h.service.DeclineInvitation(r.Context(), token)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// Verify handles GET /api/invitations/verify/{token} (and /api/invitations/token/{token}).
func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		httpx.RespondError(w, r, apperror.BadRequest("Invitation token is required"))
		return
	}

	res, err := h.service.VerifyInvitation(r.Context(), token)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetUserInvitations handles GET /api/invitations/user.
func (h *Handler) GetUserInvitations(w http.ResponseWriter, r *http.Request) {
	email := ctxhelper.GetUserEmail(r.Context())
	if email == "" {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	res, err := h.service.GetUserInvitations(r.Context(), email)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetEntityInvitations handles GET /api/invitations/entity/{entityType}/{entityId}.
func (h *Handler) GetEntityInvitations(w http.ResponseWriter, r *http.Request) {
	entityType := chi.URLParam(r, "entityType")
	entityID, err := uuid.Parse(chi.URLParam(r, "entityId"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid entity ID"))
		return
	}

	res, err := h.service.GetEntityInvitations(r.Context(), entityType, entityID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// Resend handles POST /api/invitations/{id}/resend.
func (h *Handler) Resend(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	invitationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid invitation ID"))
		return
	}

	res, err := h.service.ResendInvitation(r.Context(), invitationID, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// Delete handles DELETE /api/invitations/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	invitationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid invitation ID"))
		return
	}

	res, err := h.service.DeleteInvitation(r.Context(), invitationID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}
