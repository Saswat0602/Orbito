package editor_images

import (
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP uploads for rich-text editor images.
type Handler struct {
	service Service
}

// NewHandler creates an editor images HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// UploadImage handles POST /api/editor-images/upload.
func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}

	maxBytes := h.service.MaxSize()
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes+(1<<20))
	if pErr := r.ParseMultipartForm(maxBytes); pErr != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Uploaded image payload exceeds allowed size"))
		return
	}

	file, header, fErr := r.FormFile("file")
	if fErr != nil {
		httpx.RespondJSON(w, http.StatusOK, UploadResponse{
			Message: "No file uploaded",
			Size:    0,
			InCloud: false,
		})
		return
	}
	defer func() { _ = file.Close() }()

	content, readErr := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if readErr != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Failed reading uploaded file"))
		return
	}

	mimeType := detectMimeType(header.Header.Get("Content-Type"), content)
	res, upErr := h.service.UploadImage(r.Context(), UploadInput{
		OriginalName: header.Filename,
		MimeType:     mimeType,
		Content:      content,
	}, userID)
	if upErr != nil {
		httpx.RespondError(w, r, upErr)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

func detectMimeType(headerMime string, content []byte) string {
	trimmed := strings.ToLower(strings.TrimSpace(strings.Split(headerMime, ";")[0]))
	if trimmed != "" && trimmed != "application/octet-stream" {
		return trimmed
	}
	return strings.ToLower(strings.TrimSpace(strings.Split(http.DetectContentType(content), ";")[0]))
}

// Routes builds the /api/editor-images subrouter behind JWT authentication.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) chi.Router {
	er := chi.NewRouter()
	er.Use(authMw)
	er.Post("/upload", h.UploadImage)
	return er
}

// MountRoutes registers /api/editor-images endpoints behind JWT authentication.
func MountRoutes(r chi.Router, h *Handler, authMw func(http.Handler) http.Handler) {
	r.Mount("/editor-images", Routes(h, authMw))
}
