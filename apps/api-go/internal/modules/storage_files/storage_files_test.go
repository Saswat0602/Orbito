package storage_files

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/platform/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageFilesUploadAndServe(t *testing.T) {
	mgr, err := storage.NewManager(context.Background(), &config.Config{UploadDest: t.TempDir()}, nil)
	require.NoError(t, err)

	fh := NewHandler(mgr, &config.Config{UploadDest: mgr.UploadDir()})
	sh := NewS3Handler(mgr)

	r := chi.NewRouter()
	MountRoutes(r, fh, sh, func(next http.Handler) http.Handler { return next })

	// 1. Reject unsupported folder.
	badBody := &bytes.Buffer{}
	wBad := multipart.NewWriter(badBody)
	_, _ = wBad.CreateFormFile("file", "x.png")
	_ = wBad.Close()
	badReq := httptest.NewRequest(http.MethodPost, "/uploads/upload/tasks", badBody)
	badReq.Header.Set("Content-Type", wBad.FormDataContentType())
	badRec := httptest.NewRecorder()
	r.ServeHTTP(badRec, badReq)
	assert.Equal(t, http.StatusBadRequest, badRec.Code)

	// 2. Upload valid avatar image.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="pic.png"`)
	h.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(h)
	require.NoError(t, err)
	_, err = part.Write([]byte("\x89PNG\r\n\x1a\n"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/uploads/upload/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// 3. S3 endpoint returns 400 when S3 is not configured.
	s3Req := httptest.NewRequest(http.MethodGet, "/s3/presigned-get-url?key=avatar/pic.png", nil)
	s3Rec := httptest.NewRecorder()
	r.ServeHTTP(s3Rec, s3Req)
	assert.Equal(t, http.StatusBadRequest, s3Rec.Code)
}
