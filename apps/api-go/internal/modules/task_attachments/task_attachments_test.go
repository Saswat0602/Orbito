package task_attachments

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/platform/storage"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	access map[string]sqlc.CheckAttachmentAccessTaskAndUserRow
	items  map[uuid.UUID]sqlc.GetTaskAttachmentDetailByIDRow
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		access: make(map[string]sqlc.CheckAttachmentAccessTaskAndUserRow),
		items:  make(map[uuid.UUID]sqlc.GetTaskAttachmentDetailByIDRow),
	}
}

func (m *mockRepo) CheckAccess(_ context.Context, _ uuid.UUID, taskIdentifier string) (*sqlc.CheckAttachmentAccessTaskAndUserRow, error) {
	if r, ok := m.access[taskIdentifier]; ok {
		return &r, nil
	}
	return nil, nil
}

func (m *mockRepo) Create(_ context.Context, p CreateParams) (uuid.UUID, error) {
	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	var u, k pgtype.Text
	if p.URL != nil {
		u = pgtype.Text{String: *p.URL, Valid: true}
	}
	if p.StorageKey != nil {
		k = pgtype.Text{String: *p.StorageKey, Valid: true}
	}
	m.items[p.ID] = sqlc.GetTaskAttachmentDetailByIDRow{
		ID: p.ID, FileName: p.FileName, FileSize: p.FileSize, MimeType: p.MimeType, Url: u, StorageKey: k,
		TaskID: p.TaskID, CreatedByID: pgtype.UUID{Bytes: p.CreatedBy, Valid: true}, CreatedAt: now, UpdatedAt: now,
		TaskTitle: "Task 1", TaskSlug: "OR-1", ProjectID: uuid.New(), ProjectName: "Orbito", ProjectSlug: "orbito",
	}
	return p.ID, nil
}

func (m *mockRepo) GetDetailByID(_ context.Context, id uuid.UUID) (*sqlc.GetTaskAttachmentDetailByIDRow, error) {
	it, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	return &it, nil
}

func (m *mockRepo) ListByTaskID(_ context.Context, taskID uuid.UUID) ([]sqlc.ListTaskAttachmentDetailsByTaskIDRow, error) {
	var out []sqlc.ListTaskAttachmentDetailsByTaskIDRow
	for _, it := range m.items {
		if it.TaskID == taskID {
			out = append(out, sqlc.ListTaskAttachmentDetailsByTaskIDRow{
				ID: it.ID, FileName: it.FileName, FileSize: it.FileSize, MimeType: it.MimeType,
				Url: it.Url, StorageKey: it.StorageKey, TaskID: it.TaskID,
				CreatedAt: it.CreatedAt, UpdatedAt: it.UpdatedAt,
				TaskTitle: it.TaskTitle, TaskSlug: it.TaskSlug, ProjectID: it.ProjectID,
			})
		}
	}
	return out, nil
}

func (m *mockRepo) GroupStatsByMime(_ context.Context, taskID uuid.UUID) ([]sqlc.GroupAttachmentStatsByMimeRow, error) {
	counts := make(map[string]sqlc.GroupAttachmentStatsByMimeRow)
	for _, it := range m.items {
		if it.TaskID == taskID {
			cur := counts[it.MimeType]
			cur.MimeType = it.MimeType
			cur.CountFiles++
			cur.TotalSize += int64(it.FileSize)
			counts[it.MimeType] = cur
		}
	}
	var out []sqlc.GroupAttachmentStatsByMimeRow
	for _, v := range counts {
		out = append(out, v)
	}
	return out, nil
}

func (m *mockRepo) CountRecent(_ context.Context, taskID uuid.UUID, _ time.Time) (int64, error) {
	var cnt int64
	for _, it := range m.items {
		if it.TaskID == taskID {
			cnt++
		}
	}
	return cnt, nil
}

func (m *mockRepo) DeleteByID(_ context.Context, id uuid.UUID) (int64, error) {
	if _, ok := m.items[id]; !ok {
		return 0, nil
	}
	delete(m.items, id)
	return 1, nil
}

func TestTaskAttachments_UploadPreviewDownloadDelete(t *testing.T) {
	store, err := storage.NewLocalProvider(t.TempDir())
	require.NoError(t, err)

	repo := newMockRepo()
	taskID, userID, otherID := uuid.New(), uuid.New(), uuid.New()
	repo.access[taskID.String()] = sqlc.CheckAttachmentAccessTaskAndUserRow{
		TaskID: taskID, ProjectID: uuid.New(), UserExists: true,
		ProjectRole: sqlc.NullRole{Role: sqlc.RoleMEMBER, Valid: true},
	}

	h := NewHandler(NewService(repo, store))
	currentUser := userID
	router := Routes(h, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(ctxhelper.WithUserID(r.Context(), currentUser.String())))
		})
	})

	// Multipart upload
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	hdr := textproto.MIMEHeader{}
	hdr.Set("Content-Disposition", `form-data; name="file"; filename="hello.txt"`)
	hdr.Set("Content-Type", "text/plain")
	part, err := mw.CreatePart(hdr)
	require.NoError(t, err)
	_, err = part.Write([]byte("hello orbito attachment"))
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	reqUp := httptest.NewRequest(http.MethodPost, "/upload/"+taskID.String(), &body)
	reqUp.Header.Set("Content-Type", mw.FormDataContentType())
	rrUp := httptest.NewRecorder()
	router.ServeHTTP(rrUp, reqUp)
	require.Equal(t, http.StatusCreated, rrUp.Code)

	var created AttachmentResponse
	require.NoError(t, json.Unmarshal(rrUp.Body.Bytes(), &created))
	assert.Equal(t, "hello.txt", created.FileName)

	// Preview streams file
	reqPrev := httptest.NewRequest(http.MethodGet, "/"+created.ID+"/preview", nil)
	rrPrev := httptest.NewRecorder()
	router.ServeHTTP(rrPrev, reqPrev)
	assert.Equal(t, http.StatusOK, rrPrev.Code)
	assert.Contains(t, rrPrev.Body.String(), "hello orbito attachment")

	// Other non-admin member gets 403 deleting
	currentUser = otherID
	reqDelOther := httptest.NewRequest(http.MethodDelete, "/"+created.ID, nil)
	rrDelOther := httptest.NewRecorder()
	router.ServeHTTP(rrDelOther, reqDelOther)
	assert.Equal(t, http.StatusForbidden, rrDelOther.Code)

	// Creator deletes cleanly (204)
	currentUser = userID
	reqDel := httptest.NewRequest(http.MethodDelete, "/"+created.ID, nil)
	rrDel := httptest.NewRecorder()
	router.ServeHTTP(rrDel, reqDel)
	assert.Equal(t, http.StatusNoContent, rrDel.Code)
}
