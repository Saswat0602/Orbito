package task_comments

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	access   map[string]sqlc.CheckCommentAccessTaskAndUserRow
	comments map[uuid.UUID]sqlc.GetTaskCommentDetailByIDRow
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		access:   make(map[string]sqlc.CheckCommentAccessTaskAndUserRow),
		comments: make(map[uuid.UUID]sqlc.GetTaskCommentDetailByIDRow),
	}
}

func (m *mockRepo) CheckAccess(_ context.Context, _ uuid.UUID, taskIdentifier string) (*sqlc.CheckCommentAccessTaskAndUserRow, error) {
	if r, ok := m.access[taskIdentifier]; ok {
		return &r, nil
	}
	return nil, nil
}

func (m *mockRepo) GetBasicByID(_ context.Context, id uuid.UUID) (*sqlc.GetCommentBasicByIDRow, error) {
	c, ok := m.comments[id]
	if !ok {
		return nil, nil
	}
	return &sqlc.GetCommentBasicByIDRow{ID: c.ID, AuthorID: c.AuthorID, TaskID: c.TaskID, Content: c.Content, ParentCommentID: c.ParentCommentID}, nil
}

func (m *mockRepo) Create(_ context.Context, id uuid.UUID, content string, taskID, authorID uuid.UUID, parentCommentID *uuid.UUID) (uuid.UUID, error) {
	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	var p pgtype.UUID
	if parentCommentID != nil {
		p = pgtype.UUID{Bytes: *parentCommentID, Valid: true}
	}
	m.comments[id] = sqlc.GetTaskCommentDetailByIDRow{
		ID: id, Content: content, TaskID: taskID, AuthorID: authorID, ParentCommentID: p,
		CreatedAt: now, UpdatedAt: now, AuthorEmail: "a@orbito.io", AuthorFirstName: "Ada", AuthorLastName: "L",
		TaskTitle: "Task 1", TaskSlug: "OR-1", ProjectID: uuid.New(), ProjectName: "Orbito", ProjectSlug: "orbito",
	}
	return id, nil
}

func (m *mockRepo) GetDetailByID(_ context.Context, id uuid.UUID) (*sqlc.GetTaskCommentDetailByIDRow, error) {
	c, ok := m.comments[id]
	if !ok {
		return nil, nil
	}
	return &c, nil
}

func (m *mockRepo) CountTopLevel(_ context.Context, taskID uuid.UUID) (int64, error) {
	var cnt int64
	for _, c := range m.comments {
		if c.TaskID == taskID && !c.ParentCommentID.Valid {
			cnt++
		}
	}
	return cnt, nil
}

func (m *mockRepo) ListTopLevelAsc(_ context.Context, taskID uuid.UUID, _, _ int32) ([]sqlc.ListTopLevelTaskCommentsAscRow, error) {
	var out []sqlc.ListTopLevelTaskCommentsAscRow
	for _, c := range m.comments {
		if c.TaskID == taskID && !c.ParentCommentID.Valid {
			out = append(out, sqlc.ListTopLevelTaskCommentsAscRow{
				ID: c.ID, Content: c.Content, TaskID: c.TaskID, AuthorID: c.AuthorID,
				CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
				AuthorEmail: c.AuthorEmail, AuthorFirstName: c.AuthorFirstName, AuthorLastName: c.AuthorLastName,
				TaskTitle: c.TaskTitle, TaskSlug: c.TaskSlug,
			})
		}
	}
	return out, nil
}

func (m *mockRepo) ListTopLevelDesc(ctx context.Context, taskID uuid.UUID, limit, offset int32) ([]sqlc.ListTopLevelTaskCommentsDescRow, error) {
	rows, _ := m.ListTopLevelAsc(ctx, taskID, limit, offset)
	out := make([]sqlc.ListTopLevelTaskCommentsDescRow, len(rows))
	for i, r := range rows {
		out[i] = sqlc.ListTopLevelTaskCommentsDescRow(r)
	}
	return out, nil
}

func (m *mockRepo) ListRepliesByParentIDs(_ context.Context, _ []uuid.UUID) ([]sqlc.ListRepliesByParentIDsRow, error) {
	return nil, nil
}

func (m *mockRepo) UpdateContent(_ context.Context, id uuid.UUID, content string) (uuid.UUID, error) {
	c := m.comments[id]
	c.Content = content
	m.comments[id] = c
	return id, nil
}

func (m *mockRepo) DeleteByID(_ context.Context, id uuid.UUID) (int64, error) {
	if _, ok := m.comments[id]; !ok {
		return 0, nil
	}
	delete(m.comments, id)
	return 1, nil
}

func TestTaskComments_SanitizeEditAuthAndRateLimit(t *testing.T) {
	repo := newMockRepo()
	taskID, authorID, otherID := uuid.New(), uuid.New(), uuid.New()
	acc := sqlc.CheckCommentAccessTaskAndUserRow{
		TaskID: taskID, ProjectID: uuid.New(), UserExists: true,
		ProjectRole: sqlc.NullRole{Role: sqlc.RoleMEMBER, Valid: true},
	}
	repo.access[taskID.String()] = acc

	svc := NewService(repo, nil, nil)
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateRequest{
		Content: `<p>Hello</p><script>alert(1)</script>`,
		TaskID:  taskID.String(),
	}, authorID)
	require.NoError(t, err)
	assert.NotContains(t, created.Content, "<script>")
	assert.Contains(t, created.Content, "<p>Hello</p>")

	// Non-author cannot edit
	cid, _ := uuid.Parse(created.ID)
	_, err = svc.Update(ctx, cid, UpdateRequest{Content: "hacked"}, otherID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "your own comments")

	// Non-admin non-author cannot delete
	err = svc.Remove(ctx, cid, otherID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "your own comments")

	// Author deletes cleanly
	require.NoError(t, svc.Remove(ctx, cid, authorID))

	// HTTP rate limit enforces 20/min (21st returns 429)
	h := NewHandler(svc)
	router := Routes(h, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxReq := ctxhelper.WithUserID(r.Context(), authorID.String())
			next.ServeHTTP(w, r.WithContext(ctxReq))
		})
	})

	payload, _ := json.Marshal(CreateRequest{Content: "hi", TaskID: taskID.String()})
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusCreated, rr.Code)
	}
	req21 := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
	rr21 := httptest.NewRecorder()
	router.ServeHTTP(rr21, req21)
	assert.Equal(t, http.StatusTooManyRequests, rr21.Code)
}
