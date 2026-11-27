package search

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP search endpoints under /api/search.
type Handler struct {
	service Service
}

// NewHandler builds a search HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// GlobalSearch handles POST /api/search/global.
func (h *Handler) GlobalSearch(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}
	var req GlobalRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.GlobalSearch(r.Context(), req, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// AdvancedSearch handles POST /api/search/advanced.
func (h *Handler) AdvancedSearch(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}
	var req AdvancedRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.AdvancedSearch(r.Context(), req, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Suggestions handles GET /api/search/suggestions.
func (h *Handler) Suggestions(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}
	q := httpx.QueryParam(r, "q", "")
	limit := httpx.QueryParamInt(r, "limit", 10)
	res, err := h.service.GetSuggestions(r.Context(), q, limit, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// QuickSearch handles GET /api/search/quick.
func (h *Handler) QuickSearch(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}
	q := httpx.QueryParam(r, "q", httpx.QueryParam(r, "query", ""))
	entityType := httpx.QueryParam(r, "type", httpx.QueryParam(r, "entityType", "all"))

	req := GlobalRequest{
		Query:          q,
		EntityType:     entityType,
		OrganizationID: httpx.QueryParam(r, "organizationId", ""),
		WorkspaceID:    httpx.QueryParam(r, "workspaceId", ""),
		ProjectID:      httpx.QueryParam(r, "projectId", ""),
		Page:           httpx.QueryParamInt(r, "page", 1),
		Limit:          httpx.QueryParamInt(r, "limit", 20),
		SortBy:         httpx.QueryParam(r, "sortBy", "relevance"),
		SortOrder:      httpx.QueryParam(r, "sortOrder", "desc"),
	}
	res, err := h.service.GlobalSearch(r.Context(), req, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Routes builds the authenticated, rate-limited (30 req/min) /api/search subrouter.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) chi.Router {
	sr := chi.NewRouter()
	sr.Use(authMw)
	sr.Use(middleware.UserRateLimit(30, time.Minute))
	sr.Post("/global", h.GlobalSearch)
	sr.Post("/advanced", h.AdvancedSearch)
	sr.Get("/suggestions", h.Suggestions)
	sr.Get("/quick", h.QuickSearch)
	return sr
}

// MountRoutes registers authenticated, rate-limited (30 req/min) /api/search endpoints.
func MountRoutes(r chi.Router, h *Handler, authMw func(http.Handler) http.Handler) {
	r.Mount("/search", Routes(h, authMw))
}
