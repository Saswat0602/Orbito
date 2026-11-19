package settings

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP requests for system and user settings.
type Handler struct {
	service *Service
}

// NewHandler constructs a new settings HTTP handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetAllSettings handles GET /api/settings.
func (h *Handler) GetAllSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := getUserIDFromContext(r)

	var category *string
	if cat := r.URL.Query().Get("category"); cat != "" {
		category = &cat
	}

	settings, err := h.service.GetAll(ctx, userID, category)
	if err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to retrieve settings", err))
		return
	}

	httpx.RespondJSON(w, http.StatusOK, settings)
}

// GetSetting handles GET /api/settings/{key}.
func (h *Handler) GetSetting(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := chi.URLParam(r, "key")
	if key == "" {
		httpx.RespondError(w, r, apperror.BadRequest("Setting key is required"))
		return
	}

	userID := getUserIDFromContext(r)
	var defaultValue *string
	if def := r.URL.Query().Get("defaultValue"); def != "" {
		defaultValue = &def
	}

	val, err := h.service.Get(ctx, key, userID, defaultValue)
	if err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to retrieve setting", err))
		return
	}

	httpx.RespondJSON(w, http.StatusOK, KeyValResponse{
		Key:   key,
		Value: val,
	})
}

// SetSetting handles POST /api/settings and PATCH /api/settings.
func (h *Handler) SetSetting(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var dto SetSettingDTO
	if err := httpx.BindJSON(r, &dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	userID := getUserIDFromContext(r)

	if err := h.service.Set(ctx, dto, userID); err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to update setting", err))
		return
	}

	status := http.StatusCreated
	if r.Method == http.MethodPatch {
		status = http.StatusOK
	}
	httpx.RespondJSON(w, status, map[string]string{
		"message": "Setting updated successfully",
	})
}

// BulkSetSettings handles POST /api/settings/bulk.
func (h *Handler) BulkSetSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := getUserIDFromContext(r)
	if userID == nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required for bulk settings"))
		return
	}

	var dto BulkSetSettingsDTO
	if err := httpx.BindJSON(r, &dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	if err := h.service.BulkSet(ctx, dto.Settings, *userID); err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to bulk update settings", err))
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, map[string]string{
		"message": "Settings updated successfully",
	})
}

// DeleteSetting handles DELETE /api/settings/{key}.
func (h *Handler) DeleteSetting(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := chi.URLParam(r, "key")
	if key == "" {
		httpx.RespondError(w, r, apperror.BadRequest("Setting key is required"))
		return
	}

	userID := getUserIDFromContext(r)
	if err := h.service.Delete(ctx, key, userID); err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to delete setting", err))
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Setting deleted successfully",
	})
}

// GetRegistrationStatus handles GET /api/settings/registration-status.
func (h *Handler) GetRegistrationStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.service.GetRegistrationStatus(r.Context())
	if err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to check registration status", err))
		return
	}
	httpx.RespondJSON(w, http.StatusOK, status)
}

func getUserIDFromContext(r *http.Request) *uuid.UUID {
	raw := ctxhelper.GetUserID(r.Context())
	if raw == "" {
		return nil
	}
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &parsed
}
