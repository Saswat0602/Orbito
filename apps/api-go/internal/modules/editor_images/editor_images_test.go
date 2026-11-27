package editor_images

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/platform/storage"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepo struct {
	created []sqlc.CreateMediaAssetRecordParams
}

func (s *stubRepo) CreateMediaAsset(_ context.Context, arg sqlc.CreateMediaAssetRecordParams) (sqlc.MediaAsset, error) {
	s.created = append(s.created, arg)
	return sqlc.MediaAsset{ID: arg.ID, AssetType: arg.AssetType, StorageKey: arg.StorageKey, Filename: arg.Filename}, nil
}

func TestEditorImagesUploadSuccessAndValidation(t *testing.T) {
	mgr, err := storage.NewManager(context.Background(), &config.Config{UploadDest: t.TempDir()}, nil)
	require.NoError(t, err)

	repo := &stubRepo{}
	svc := NewService(repo, mgr, &config.Config{MaxEditorImageSize: 5 * 1024 * 1024})
	h := NewHandler(svc)

	userID := uuid.New()
	r := chi.NewRouter()
	MountRoutes(r, h, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(w, req.WithContext(ctxhelper.WithUserID(req.Context(), userID.String())))
		})
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	hdr := make(textproto.MIMEHeader)
	hdr.Set("Content-Disposition", `form-data; name="file"; filename="shot.png"`)
	hdr.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(hdr)
	require.NoError(t, err)
	_, err = part.Write([]byte("\x89PNG\r\n\x1a\ncontent"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/editor-images/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp UploadResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Image uploaded successfully", resp.Message)
	require.NotNil(t, resp.ID)
	require.NotNil(t, resp.Key)
	assert.Contains(t, *resp.Key, "editor-images/")
	assert.False(t, resp.InCloud)
	require.Len(t, repo.created, 1)
	assert.Equal(t, "editor-image", repo.created[0].AssetType)

	// Invalid MIME type returns 400.
	badBody := &bytes.Buffer{}
	bw := multipart.NewWriter(badBody)
	bHdr := make(textproto.MIMEHeader)
	bHdr.Set("Content-Disposition", `form-data; name="file"; filename="bad.txt"`)
	bHdr.Set("Content-Type", "text/plain")
	bPart, _ := bw.CreatePart(bHdr)
	_, _ = bPart.Write([]byte("hello"))
	_ = bw.Close()

	badReq := httptest.NewRequest(http.MethodPost, "/editor-images/upload", badBody)
	badReq.Header.Set("Content-Type", bw.FormDataContentType())
	badRec := httptest.NewRecorder()
	r.ServeHTTP(badRec, badReq)
	assert.Equal(t, http.StatusBadRequest, badRec.Code)
}
