package inbox

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles project inbox, templates, and comment-email HTTP routes.
type Handler struct {
	service Service
}

// NewHandler creates an inbox HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// CreateInbox handles POST /api/projects/{projectId}/inbox.
func (h *Handler) CreateInbox(w http.ResponseWriter, r *http.Request) {
	projectID, err := httpx.URLParamUUID(r, "projectId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req CreateInboxRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.CreateInbox(r.Context(), projectID, req)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// GetInbox handles GET /api/projects/{projectId}/inbox.
func (h *Handler) GetInbox(w http.ResponseWriter, r *http.Request) {
	projectID, err := httpx.URLParamUUID(r, "projectId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetInbox(r.Context(), projectID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// UpdateInbox handles PUT /api/projects/{projectId}/inbox.
func (h *Handler) UpdateInbox(w http.ResponseWriter, r *http.Request) {
	projectID, err := httpx.URLParamUUID(r, "projectId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req UpdateInboxRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.UpdateInbox(r.Context(), projectID, req)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// SetupEmailAccount handles PUT /api/projects/{projectId}/inbox/email-account.
func (h *Handler) SetupEmailAccount(w http.ResponseWriter, r *http.Request) {
	projectID, err := httpx.URLParamUUID(r, "projectId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req SetupEmailRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.SetupEmailAccount(r.Context(), projectID, req)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// TriggerSync handles POST /api/projects/{projectId}/inbox/sync.
func (h *Handler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	projectID, err := httpx.URLParamUUID(r, "projectId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var userPtr *uuid.UUID
	if uid, pErr := uuid.Parse(ctxhelper.GetUserID(r.Context())); pErr == nil {
		userPtr = &uid
	}
	jobID, err := h.service.TriggerSyncJob(r.Context(), projectID, userPtr)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"message": "Email sync job queued successfully",
		"jobId":   jobID,
		"status":  "queued",
	})
}

// GetMessages handles GET /api/projects/{projectId}/inbox/messages.
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	projectID, err := httpx.URLParamUUID(r, "projectId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	f := MessageFilters{
		Status:        httpx.QueryParam(r, "status", ""),
		IncludeSpam:   httpx.QueryParam(r, "includeSpam", "") == "true",
		FromEmail:     httpx.QueryParam(r, "fromEmail", ""),
		FromDate:      httpx.QueryParam(r, "fromDate", ""),
		ToDate:        httpx.QueryParam(r, "toDate", ""),
		Search:        httpx.QueryParam(r, "search", ""),
		ConvertedOnly: httpx.QueryParam(r, "convertedOnly", "") == "true",
	}
	res, err := h.service.GetMessages(r.Context(), projectID, f)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetMessage handles GET /api/projects/{projectId}/inbox/messages/{messageId}.
func (h *Handler) GetMessage(w http.ResponseWriter, r *http.Request) {
	msgID, err := httpx.URLParamUUID(r, "messageId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetMessage(r.Context(), msgID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// ConvertToTask handles POST /api/projects/{projectId}/inbox/messages/{messageId}/convert.
func (h *Handler) ConvertToTask(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}
	msgID, err := httpx.URLParamUUID(r, "messageId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.ConvertToTask(r.Context(), msgID, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// IgnoreMessage handles POST /api/projects/{projectId}/inbox/messages/{messageId}/ignore.
func (h *Handler) IgnoreMessage(w http.ResponseWriter, r *http.Request) {
	msgID, err := httpx.URLParamUUID(r, "messageId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.UpdateMessageStatus(r.Context(), msgID, "IGNORED")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetRules handles GET /api/projects/{projectId}/inbox/rules.
func (h *Handler) GetRules(w http.ResponseWriter, r *http.Request) {
	projectID, err := httpx.URLParamUUID(r, "projectId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetRules(r.Context(), projectID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// CreateRule handles POST /api/projects/{projectId}/inbox/rules.
func (h *Handler) CreateRule(w http.ResponseWriter, r *http.Request) {
	projectID, err := httpx.URLParamUUID(r, "projectId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req CreateRuleRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.CreateRule(r.Context(), projectID, req)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// UpdateRule handles PUT /api/projects/{projectId}/inbox/rules/{ruleId}.
func (h *Handler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	ruleID, err := httpx.URLParamUUID(r, "ruleId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req CreateRuleRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.UpdateRule(r.Context(), ruleID, req)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// DeleteRule handles DELETE /api/projects/{projectId}/inbox/rules/{ruleId}.
func (h *Handler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	ruleID, err := httpx.URLParamUUID(r, "ruleId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if err := h.service.DeleteRule(r.Context(), ruleID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]string{"message": "Rule deleted successfully"})
}

// TestEmailConfig handles POST /api/projects/{projectId}/inbox/test-email/{accountId}.
func (h *Handler) TestEmailConfig(w http.ResponseWriter, r *http.Request) {
	acctID, err := httpx.URLParamUUID(r, "accountId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.TestEmailConfiguration(r.Context(), acctID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// SendCommentAsEmail handles POST /api/tasks/{taskId}/comments/{commentId}/send-email.
func (h *Handler) SendCommentAsEmail(w http.ResponseWriter, r *http.Request) {
	commentID, err := httpx.URLParamUUID(r, "commentId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.SendCommentAsEmail(r.Context(), commentID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// ProjectInboxRoutes mounts /api/projects/{projectId}/inbox routes behind JWT auth.
func ProjectInboxRoutes(h *Handler, authMw func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Use(authMw)
	r.Post("/", h.CreateInbox)
	r.Get("/", h.GetInbox)
	r.Put("/", h.UpdateInbox)
	r.Put("/email-account", h.SetupEmailAccount)
	r.Post("/sync", h.TriggerSync)
	r.Get("/messages", h.GetMessages)
	r.Get("/messages/{messageId}", h.GetMessage)
	r.Post("/messages/{messageId}/convert", h.ConvertToTask)
	r.Post("/messages/{messageId}/ignore", h.IgnoreMessage)
	r.Get("/rules", h.GetRules)
	r.Post("/rules", h.CreateRule)
	r.Put("/rules/{ruleId}", h.UpdateRule)
	r.Delete("/rules/{ruleId}", h.DeleteRule)
	r.Post("/test-email/{accountId}", h.TestEmailConfig)
	return r
}

// EmailTemplatesRoutes builds /api/email-templates routes behind JWT auth.
func EmailTemplatesRoutes(authMw func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Use(authMw)
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		cat := httpx.QueryParam(req, "category", "")
		search := httpx.QueryParam(req, "search", "")
		httpx.RespondJSON(w, http.StatusOK, GetTemplates(cat, search))
	})
	r.Get("/categories", func(w http.ResponseWriter, _ *http.Request) {
		httpx.RespondJSON(w, http.StatusOK, GetTemplateCategories())
	})
	r.Get("/variables", func(w http.ResponseWriter, _ *http.Request) {
		httpx.RespondJSON(w, http.StatusOK, GetCommonVariables())
	})
	r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
		tpl := GetTemplateByID(chi.URLParam(req, "id"))
		if tpl == nil {
			httpx.RespondJSON(w, http.StatusOK, map[string]string{"error": "Template not found"})
			return
		}
		httpx.RespondJSON(w, http.StatusOK, tpl)
	})
	return r
}
