package users

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
	"github.com/orbito/api-go/internal/shared/pagination"
)

var validIdentifierRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// Handler handles HTTP requests for user profiles, directory, and status operations.
type Handler struct {
	service *Service
}

// NewHandler constructs a new users HTTP handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CheckUsersExist handles GET /api/users/exists.
func (h *Handler) CheckUsersExist(w http.ResponseWriter, r *http.Request) {
	exists, err := h.service.CheckUsersExist(r.Context())
	if err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to check user existence", err))
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]bool{"exists": exists})
}

// Create handles POST /api/users (SUPER_ADMIN only).
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	role := ctxhelper.GetUserRole(r.Context())
	if role != string(sqlc.RoleSUPERADMIN) {
		httpx.RespondError(w, r, apperror.Forbidden("Only administrators can create users directly"))
		return
	}

	var dto CreateUserDTO
	if err := httpx.BindJSON(r, &dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	res, err := h.service.Create(r.Context(), dto)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, res)
}

// FindAll handles GET /api/users (SUPER_ADMIN only).
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	role := ctxhelper.GetUserRole(r.Context())
	if role != string(sqlc.RoleSUPERADMIN) {
		httpx.RespondError(w, r, apperror.Forbidden("Only administrators can view the user directory"))
		return
	}

	limit := 100
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}
	page := 1
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	offset := pagination.Offset(page, limit)

	//nolint:gosec // limit and offset are bounded above
	users, _, err := h.service.FindAll(r.Context(), int32(limit), int32(offset))
	if err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to list users", err))
		return
	}

	if users == nil {
		users = []ResponseDTO{}
	}
	httpx.RespondJSON(w, http.StatusOK, users)
}

// FindOne handles GET /api/users/{id}.
func (h *Handler) FindOne(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid user ID format"))
		return
	}

	requesterID, _ := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	requesterRole := ctxhelper.GetUserRole(r.Context())

	user, err := h.service.FindOne(r.Context(), targetID, requesterID, requesterRole)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, user)
}

// GetPublicProfile handles GET /api/users/{id}/profile.
func (h *Handler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if !validIdentifierRegex.MatchString(idStr) {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid identifier format"))
		return
	}

	requesterID, _ := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	requesterRole := ctxhelper.GetUserRole(r.Context())

	profile, err := h.service.GetPublicProfile(r.Context(), idStr, requesterID, requesterRole)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, profile)
}

// UpdateCurrentUserProfile handles PATCH /api/users/profile.
func (h *Handler) UpdateCurrentUserProfile(w http.ResponseWriter, r *http.Request) {
	requesterIDStr := ctxhelper.GetUserID(r.Context())
	requesterID, parseErr := uuid.Parse(requesterIDStr)
	if parseErr != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthorized"))
		return
	}

	var dto UpdateUserDTO
	if err := httpx.BindJSON(r, &dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	requesterRole := ctxhelper.GetUserRole(r.Context())
	res, err := h.service.Update(r.Context(), requesterID, dto, requesterID, requesterRole)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// Update handles PATCH /api/users/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	targetID, parseErr := uuid.Parse(idStr)
	if parseErr != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid user ID format"))
		return
	}

	var dto UpdateUserDTO
	if err := httpx.BindJSON(r, &dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	requesterID, _ := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	requesterRole := ctxhelper.GetUserRole(r.Context())

	res, err := h.service.Update(r.Context(), targetID, dto, requesterID, requesterRole)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// UpdateStatus handles PATCH /api/users/{id}/status.
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	targetID, parseErr := uuid.Parse(idStr)
	if parseErr != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid user ID format"))
		return
	}

	var dto UpdateStatusDTO
	if err := httpx.BindJSON(r, &dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	requesterRole := ctxhelper.GetUserRole(r.Context())
	res, err := h.service.UpdateStatus(r.Context(), targetID, sqlc.UserStatus(dto.Status), requesterRole)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// ChangePassword handles POST /api/users/change-password and PATCH /api/users/{id}/change-password.
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	targetIDStr := chi.URLParam(r, "id")
	if targetIDStr == "" {
		targetIDStr = ctxhelper.GetUserID(r.Context())
	}

	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid user ID format"))
		return
	}

	var dto ChangePasswordDTO
	if err := httpx.BindJSON(r, &dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	if err := h.service.ChangePassword(r.Context(), targetID, dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Password changed successfully",
	})
}

// Remove handles DELETE /api/users/{id} (SUPER_ADMIN only).
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid user ID format"))
		return
	}

	requesterID, _ := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	requesterRole := ctxhelper.GetUserRole(r.Context())

	if err := h.service.Remove(r.Context(), targetID, requesterID, requesterRole); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUserStatus handles GET /api/users/{id}/status.
func (h *Handler) GetUserStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid user ID format"))
		return
	}

	status := h.service.GetOnlineStatus(r.Context(), targetID)
	httpx.RespondJSON(w, http.StatusOK, status)
}

// GetUsersStatus handles GET /api/users/status/bulk.
func (h *Handler) GetUsersStatus(w http.ResponseWriter, r *http.Request) {
	idsParam := r.URL.Query().Get("userIds")
	statusMap := make(map[string]OnlineStatusDTO)
	for _, idStr := range strings.Split(idsParam, ",") {
		idStr = strings.TrimSpace(idStr)
		if id, err := uuid.Parse(idStr); err == nil {
			statusMap[idStr] = h.service.GetOnlineStatus(r.Context(), id)
		}
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]any{"status": statusMap})
}
