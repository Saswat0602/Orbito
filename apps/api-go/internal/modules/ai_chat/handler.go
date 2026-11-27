package ai_chat

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler serves HTTP endpoints under /api/ai-chat.
type Handler struct {
	service Service
}

// NewHandler creates an AI chat HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// GetConfig handles GET /api/ai-chat/config.
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}
	res, err := h.service.GetPublicConfig(r.Context(), userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetConversations handles GET /api/ai-chat/conversations.
func (h *Handler) GetConversations(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}
	res, err := h.service.GetConversations(r.Context(), userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// CreateConversation handles POST /api/ai-chat/conversations.
func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}
	var req CreateConversationRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.CreateConversation(r.Context(), userID, req)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// RenameConversation handles PATCH /api/ai-chat/conversations/{id}.
func (h *Handler) RenameConversation(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req RenameConversationRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.RenameConversation(r.Context(), userID, id, req)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// DeleteConversation handles DELETE /api/ai-chat/conversations/{id}.
func (h *Handler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if dErr := h.service.DeleteConversation(r.Context(), userID, id); dErr != nil {
		httpx.RespondError(w, r, dErr)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// UpdateMessages handles PUT /api/ai-chat/conversations/{id}/messages.
func (h *Handler) UpdateMessages(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req UpdateMessagesRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	if uErr := h.service.UpdateMessages(r.Context(), userID, id, req); uErr != nil {
		httpx.RespondError(w, r, uErr)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// Chat handles POST /api/ai-chat/chat.
func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}
	var req ChatRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Chat(r.Context(), userID, req)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// TestConnection handles POST /api/ai-chat/test-connection.
func (h *Handler) TestConnection(w http.ResponseWriter, r *http.Request) {
	var req TestConnectionRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.TestConnection(r.Context(), req)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GenerateDescription handles POST /api/ai-chat/generate-description.
func (h *Handler) GenerateDescription(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}
	var req GenerateDescriptionRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.GenerateDescription(r.Context(), userID, req)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// ClearContext handles DELETE /api/ai-chat/context/{sessionId}.
func (h *Handler) ClearContext(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}
	sessionID := chi.URLParam(r, "sessionId")
	if cErr := h.service.ClearContext(r.Context(), userID, sessionID); cErr != nil {
		httpx.RespondError(w, r, cErr)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// Routes builds the /api/ai-chat subrouter behind JWT auth.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Use(authMw)
	r.Get("/config", h.GetConfig)
	r.Get("/conversations", h.GetConversations)
	r.Post("/conversations", h.CreateConversation)
	r.Patch("/conversations/{id}", h.RenameConversation)
	r.Delete("/conversations/{id}", h.DeleteConversation)
	r.Put("/conversations/{id}/messages", h.UpdateMessages)
	r.Post("/chat", h.Chat)
	r.Post("/test-connection", h.TestConnection)
	r.Post("/generate-description", h.GenerateDescription)
	r.Delete("/context/{sessionId}", h.ClearContext)
	return r
}
