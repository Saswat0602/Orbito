package admin

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler processes administrator HTTP operations.
type Handler struct {
	service *Service
}

// NewHandler constructs a new admin HTTP handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) assertAdmin(w http.ResponseWriter, r *http.Request) bool {
	role := ctxhelper.GetUserRole(r.Context())
	if role != string(sqlc.RoleSUPERADMIN) {
		httpx.RespondError(w, r, apperror.Forbidden("Forbidden resource"))
		return false
	}
	return true
}

// GetDashboardStats handles GET /api/admin/dashboard.
func (h *Handler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	if !h.assertAdmin(w, r) {
		return
	}

	stats, err := h.service.GetDashboardStats(r.Context())
	if err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to load dashboard stats", err))
		return
	}

	httpx.RespondJSON(w, http.StatusOK, stats)
}

// GetUsers handles GET /api/admin/users.
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	if !h.assertAdmin(w, r) {
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")
	role := r.URL.Query().Get("role")

	res, err := h.service.GetUsers(r.Context(), page, limit, search, status, role)
	if err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to list users", err))
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetUserDetail handles GET /api/admin/users/{id}.
func (h *Handler) GetUserDetail(w http.ResponseWriter, r *http.Request) {
	if !h.assertAdmin(w, r) {
		return
	}

	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}

	detail, err := h.service.GetUserDetail(r.Context(), targetID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, detail)
}

// UpdateUserRole handles PATCH /api/admin/users/{id}/role.
func (h *Handler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	if !h.assertAdmin(w, r) {
		return
	}

	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}

	var dto UpdateUserRoleDTO
	if bindErr := httpx.BindJSON(r, &dto); bindErr != nil {
		httpx.RespondError(w, r, bindErr)
		return
	}

	currentUserID, _ := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	res, err := h.service.UpdateUserRole(r.Context(), targetID, dto.Role, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// UpdateUserStatus handles PATCH /api/admin/users/{id}/status.
func (h *Handler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	if !h.assertAdmin(w, r) {
		return
	}

	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}

	var dto UpdateUserStatusDTO
	if bindErr := httpx.BindJSON(r, &dto); bindErr != nil {
		httpx.RespondError(w, r, bindErr)
		return
	}

	currentUserID, _ := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	res, err := h.service.UpdateUserStatus(r.Context(), targetID, dto.Status, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// DeleteUser handles DELETE /api/admin/users/{id}.
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if !h.assertAdmin(w, r) {
		return
	}

	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}

	currentUserID, _ := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	res, err := h.service.DeleteUser(r.Context(), targetID, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// ResetUserPassword handles POST /api/admin/users/{id}/reset-password.
func (h *Handler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	if !h.assertAdmin(w, r) {
		return
	}

	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}

	res, err := h.service.ResetUserPassword(r.Context(), targetID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetOrganizations handles GET /api/admin/organizations.
func (h *Handler) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	if !h.assertAdmin(w, r) {
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")

	res, err := h.service.GetOrganizations(r.Context(), page, limit, search)
	if err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to list organizations", err))
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetOrganizationDetail handles GET /api/admin/organizations/{id}.
func (h *Handler) GetOrganizationDetail(w http.ResponseWriter, r *http.Request) {
	if !h.assertAdmin(w, r) {
		return
	}

	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}

	detail, err := h.service.GetOrganizationDetail(r.Context(), targetID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, detail)
}

// DeleteOrganization handles DELETE /api/admin/organizations/{id}.
func (h *Handler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	if !h.assertAdmin(w, r) {
		return
	}

	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}

	res, err := h.service.DeleteOrganization(r.Context(), targetID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// ToggleOrganizationArchive handles PATCH /api/admin/organizations/{id}/archive.
func (h *Handler) ToggleOrganizationArchive(w http.ResponseWriter, r *http.Request) {
	if !h.assertAdmin(w, r) {
		return
	}

	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}

	res, err := h.service.ToggleOrganizationArchive(r.Context(), targetID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// TransferOrganizationOwnership handles PATCH /api/admin/organizations/{id}/transfer-ownership.
func (h *Handler) TransferOrganizationOwnership(w http.ResponseWriter, r *http.Request) {
	if !h.assertAdmin(w, r) {
		return
	}

	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}

	var dto TransferOwnershipDTO
	if bindErr := httpx.BindJSON(r, &dto); bindErr != nil {
		httpx.RespondError(w, r, bindErr)
		return
	}

	newOwnerID, err := uuid.Parse(dto.NewOwnerID)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid newOwnerId format"))
		return
	}

	res, err := h.service.TransferOrganizationOwnership(r.Context(), targetID, newOwnerID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetSystemConfig handles GET /api/admin/config.
func (h *Handler) GetSystemConfig(w http.ResponseWriter, r *http.Request) {
	if !h.assertAdmin(w, r) {
		return
	}

	category := r.URL.Query().Get("category")
	res, err := h.service.GetSystemConfig(r.Context(), category)
	if err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to load system config", err))
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// SaveSystemConfig handles POST /api/admin/config.
func (h *Handler) SaveSystemConfig(w http.ResponseWriter, r *http.Request) {
	if !h.assertAdmin(w, r) {
		return
	}

	var dto SaveSystemConfigDTO
	if bindErr := httpx.BindJSON(r, &dto); bindErr != nil {
		httpx.RespondError(w, r, bindErr)
		return
	}

	if err := h.service.SaveSystemConfig(r.Context(), dto.Settings); err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to save system config", err))
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// TestSMTP handles POST /api/admin/config/test-smtp.
func (h *Handler) TestSMTP(w http.ResponseWriter, r *http.Request) {
	if !h.assertAdmin(w, r) {
		return
	}

	var dto TestSMTPDTO
	_ = httpx.BindJSON(r, &dto)

	res, err := h.service.TestSMTPConfig(r.Context(), dto.Email)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}
