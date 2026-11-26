package labels

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler processes HTTP requests for labels.
type Handler struct {
	service Service
}

// NewHandler constructs a labels HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

func currentUserID(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		return uuid.Nil, apperror.Unauthorized("Unauthenticated")
	}
	return id, nil
}

// Create handles POST /api/labels.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req CreateLabelRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Create(r.Context(), req, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// FindAll handles GET /api/labels.
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var projectID *uuid.UUID
	if s := r.URL.Query().Get("projectId"); s != "" {
		id, pErr := uuid.Parse(s)
		if pErr != nil {
			httpx.RespondError(w, r, apperror.BadRequest("Invalid projectId format"))
			return
		}
		projectID = &id
	}
	res, err := h.service.FindAll(r.Context(), projectID, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindOne handles GET /api/labels/{id}.
func (h *Handler) FindOne(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.FindOne(r.Context(), id, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Update handles PATCH /api/labels/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req UpdateLabelRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Update(r.Context(), id, req, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Remove handles DELETE /api/labels/{id}.
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if err := h.service.Remove(r.Context(), id, userID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AssignLabelToTask handles POST /api/labels/assign.
func (h *Handler) AssignLabelToTask(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req AssignLabelRequest
	if err := httpx.BindJSON(r, &req); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if err := h.service.AssignLabelToTask(r.Context(), req, userID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// RemoveLabelFromTask handles DELETE /api/labels/task/{taskId}/label/{labelId}.
func (h *Handler) RemoveLabelFromTask(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	taskID := chi.URLParam(r, "taskId")
	labelID, err := httpx.URLParamUUID(r, "labelId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if err := h.service.RemoveLabelFromTask(r.Context(), taskID, labelID, userID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AssignMultipleLabelsToTask handles POST /api/labels/assign-multiple.
func (h *Handler) AssignMultipleLabelsToTask(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req AssignMultipleLabelsRequest
	if err := httpx.BindJSON(r, &req); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if err := h.service.AssignMultipleLabelsToTask(r.Context(), req, userID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// GetTaskLabels handles GET /api/labels/task/{taskId}.
func (h *Handler) GetTaskLabels(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	taskID := chi.URLParam(r, "taskId")
	res, err := h.service.GetTaskLabels(r.Context(), taskID, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}
