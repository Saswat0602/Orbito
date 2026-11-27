package time_entries

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP requests for time entries and timers.
type Handler struct {
	service Service
}

// NewHandler constructs a time_entries HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

func extractCaller(r *http.Request) (CallerContext, error) {
	uid, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		return CallerContext{}, apperror.Unauthorized("Unauthenticated")
	}
	return CallerContext{
		UserID:       uid,
		IsSuperAdmin: ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin),
	}, nil
}

func parseOptionalUUID(raw string) (*uuid.UUID, error) {
	if raw == "" {
		return nil, nil
	}
	u, err := uuid.Parse(raw)
	if err != nil {
		return nil, apperror.BadRequest("Validation failed (uuid is expected)")
	}
	return &u, nil
}

func parseOptionalTime(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return &t, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, apperror.BadRequest("Invalid date query parameter")
	}
	return &t, nil
}

// Create handles POST /api/time-entries.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	caller, err := extractCaller(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req CreateRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Create(r.Context(), req, caller)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// FindAll handles GET /api/time-entries.
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	caller, err := extractCaller(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	taskID, err := parseOptionalUUID(httpx.QueryParam(r, "taskId", ""))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	userID, err := parseOptionalUUID(httpx.QueryParam(r, "userId", ""))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	startDate, err := parseOptionalTime(httpx.QueryParam(r, "startDate", ""))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	endDate, err := parseOptionalTime(httpx.QueryParam(r, "endDate", ""))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.FindAll(r.Context(), caller, taskID, userID, startDate, endDate)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetSummary handles GET /api/time-entries/summary.
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	caller, err := extractCaller(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	userID, err := parseOptionalUUID(httpx.QueryParam(r, "userId", ""))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	taskID, err := parseOptionalUUID(httpx.QueryParam(r, "taskId", ""))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	startDate, err := parseOptionalTime(httpx.QueryParam(r, "startDate", ""))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	endDate, err := parseOptionalTime(httpx.QueryParam(r, "endDate", ""))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetSummary(r.Context(), caller, userID, taskID, startDate, endDate)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindOne handles GET /api/time-entries/{id}.
func (h *Handler) FindOne(w http.ResponseWriter, r *http.Request) {
	caller, err := extractCaller(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.FindOne(r.Context(), id, caller)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Update handles PATCH /api/time-entries/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	caller, err := extractCaller(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req UpdateRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Update(r.Context(), id, req, caller.UserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Remove handles DELETE /api/time-entries/{id}.
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	caller, err := extractCaller(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if err := h.service.Remove(r.Context(), id, caller.UserID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// StartTimer handles POST /api/time-entries/timer/start.
func (h *Handler) StartTimer(w http.ResponseWriter, r *http.Request) {
	caller, err := extractCaller(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req StartTimerRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.StartTimer(r.Context(), req, caller)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// StopTimer handles POST /api/time-entries/timer/stop.
func (h *Handler) StopTimer(w http.ResponseWriter, r *http.Request) {
	caller, err := extractCaller(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req StopTimerRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.StopTimer(r.Context(), req, caller)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetActiveTimer handles GET /api/time-entries/timer/active/{userId}.
func (h *Handler) GetActiveTimer(w http.ResponseWriter, r *http.Request) {
	caller, err := extractCaller(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	targetUserID, err := httpx.URLParamUUID(r, "userId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetActiveTimer(r.Context(), targetUserID, caller)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}
