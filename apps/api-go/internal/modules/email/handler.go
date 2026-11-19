package email

import (
	"net/http"

	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP requests for sending emails and monitoring email queue stats.
type Handler struct {
	service *Service
}

// NewHandler constructs a new email handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// SendEmail handles POST /api/email/send.
func (h *Handler) SendEmail(w http.ResponseWriter, r *http.Request) {
	var dto SendEmailDTO
	if err := httpx.BindJSON(r, &dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	if err := h.service.SendEmail(r.Context(), dto); err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to queue email", err))
		return
	}

	httpx.RespondJSON(w, http.StatusAccepted, map[string]string{
		"message": "Email queued successfully",
	})
}

// SendBulkEmail handles POST /api/email/send-bulk.
func (h *Handler) SendBulkEmail(w http.ResponseWriter, r *http.Request) {
	var dto BulkEmailDTO
	if err := httpx.BindJSON(r, &dto); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	if err := h.service.SendBulkEmail(r.Context(), dto); err != nil {
		httpx.RespondError(w, r, apperror.Internal("Failed to queue bulk emails", err))
		return
	}

	httpx.RespondJSON(w, http.StatusAccepted, map[string]any{
		"message": "Bulk emails queued successfully",
		"count":   len(dto.Recipients),
	})
}

// GetQueueStats handles GET /api/email/queue/stats.
func (h *Handler) GetQueueStats(w http.ResponseWriter, _ *http.Request) {
	stats := h.service.GetQueueStats()
	httpx.RespondJSON(w, http.StatusOK, stats)
}
