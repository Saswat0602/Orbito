package storage_files

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Presigner generates presigned AWS S3 URLs.
type Presigner interface {
	PresignPutURL(ctx context.Context, key string) (string, error)
	PresignGetURL(ctx context.Context, key string) (string, error)
}

// S3Handler serves /api/s3 presigned URL endpoints.
type S3Handler struct {
	presigner Presigner
}

// NewS3Handler builds an S3Handler.
func NewS3Handler(presigner Presigner) *S3Handler {
	return &S3Handler{presigner: presigner}
}

// GetPresignedPutURL handles GET /api/s3/presigned-put-url.
func (h *S3Handler) GetPresignedPutURL(w http.ResponseWriter, r *http.Request) {
	key := httpx.QueryParam(r, "key", "")
	url, err := h.presigner.PresignPutURL(r.Context(), key)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]string{"url": url})
}

// GetPresignedGetURL handles GET /api/s3/presigned-get-url.
func (h *S3Handler) GetPresignedGetURL(w http.ResponseWriter, r *http.Request) {
	key := httpx.QueryParam(r, "key", "")
	url, err := h.presigner.PresignGetURL(r.Context(), key)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]string{"url": url})
}

// UploadRoutes builds the /api/uploads subrouter.
func UploadRoutes(filesH *Handler, authMw func(http.Handler) http.Handler) chi.Router {
	ur := chi.NewRouter()
	ur.Get("/tasks/{taskId}/{filename}", filesH.ServeTaskFile)
	ur.Get("/editor-images/{filename}", filesH.ServeEditorImage)
	ur.Get("/{folder}/{filename}", filesH.ServeFile)
	ur.With(authMw).Post("/upload/{folder}", filesH.UploadFile)
	return ur
}

// S3Routes builds the /api/s3 subrouter.
func S3Routes(s3H *S3Handler, authMw func(http.Handler) http.Handler) chi.Router {
	sr := chi.NewRouter()
	sr.Use(authMw)
	sr.Get("/presigned-put-url", s3H.GetPresignedPutURL)
	sr.Get("/presigned-get-url", s3H.GetPresignedGetURL)
	return sr
}

// MountRoutes registers /api/uploads and /api/s3 endpoints.
func MountRoutes(r chi.Router, filesH *Handler, s3H *S3Handler, authMw func(http.Handler) http.Handler) {
	r.Mount("/uploads", UploadRoutes(filesH, authMw))
	if s3H != nil {
		r.Mount("/s3", S3Routes(s3H, authMw))
	}
}
