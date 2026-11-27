package automation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/platform/queue"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepo struct {
	rules        map[uuid.UUID]sqlc.GetAutomationRuleDetailByIDRow
	executions   []sqlc.CreateRuleExecutionRecordParams
	statusSet    uuid.UUID
	assignedIDs  []uuid.UUID
	commentAdded string
}

func (s *stubRepo) CreateRule(_ context.Context, params sqlc.CreateAutomationRuleParams) (uuid.UUID, error) {
	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	s.rules[params.ID] = sqlc.GetAutomationRuleDetailByIDRow{
		ID:             params.ID,
		Name:           params.Name,
		Description:    params.Description,
		Status:         params.Status,
		TriggerType:    params.TriggerType,
		TriggerConfig:  params.TriggerConfig,
		ActionType:     params.ActionType,
		ActionConfig:   params.ActionConfig,
		OrganizationID: params.OrganizationID,
		WorkspaceID:    params.WorkspaceID,
		ProjectID:      params.ProjectID,
		CreatedByID:    params.CreatedByID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return params.ID, nil
}

func (s *stubRepo) GetRuleDetailByID(_ context.Context, id uuid.UUID) (*sqlc.GetAutomationRuleDetailByIDRow, error) {
	r, ok := s.rules[id]
	if !ok {
		return nil, nil
	}
	return &r, nil
}

func (s *stubRepo) ListRules(_ context.Context, _, _, _ pgtype.UUID) ([]sqlc.ListAutomationRulesFilteredRow, error) {
	return nil, nil
}

func (s *stubRepo) ListRuleExecutions(_ context.Context, _ uuid.UUID, _ int32) ([]sqlc.ListRuleExecutionsByRuleIDRow, error) {
	return nil, nil
}

func (s *stubRepo) UpdateRule(_ context.Context, params sqlc.UpdateAutomationRuleByIDParams) (int64, error) {
	if _, ok := s.rules[params.ID]; !ok {
		return 0, nil
	}
	return 1, nil
}

func (s *stubRepo) ToggleRuleStatus(_ context.Context, id uuid.UUID) (*sqlc.AutomationRule, error) {
	r, ok := s.rules[id]
	if !ok {
		return nil, nil
	}
	if r.Status == sqlc.RuleStatusACTIVE {
		r.Status = sqlc.RuleStatusINACTIVE
	} else {
		r.Status = sqlc.RuleStatusACTIVE
	}
	s.rules[id] = r
	return &sqlc.AutomationRule{ID: id, Status: r.Status}, nil
}

func (s *stubRepo) DeleteRule(_ context.Context, id uuid.UUID) (int64, error) {
	if _, ok := s.rules[id]; !ok {
		return 0, nil
	}
	delete(s.rules, id)
	return 1, nil
}

func (s *stubRepo) GetRuleStats(_ context.Context, _ uuid.UUID) (*sqlc.GetAutomationRuleStatsRowRow, error) {
	return &sqlc.GetAutomationRuleStatsRowRow{TotalExecutions: 2, SuccessCount: 1, AvgExecutionTime: 15}, nil
}

func (s *stubRepo) ListActiveRulesByTrigger(_ context.Context, trigger sqlc.TriggerType, projectID, _, _ pgtype.UUID) ([]sqlc.AutomationRule, error) {
	var out []sqlc.AutomationRule
	for _, r := range s.rules {
		if r.Status == sqlc.RuleStatusACTIVE && r.TriggerType == trigger && (!projectID.Valid || r.ProjectID == projectID) {
			out = append(out, sqlc.AutomationRule{
				ID:            r.ID,
				Status:        r.Status,
				TriggerType:   r.TriggerType,
				TriggerConfig: r.TriggerConfig,
				ActionType:    r.ActionType,
				ActionConfig:  r.ActionConfig,
				ProjectID:     r.ProjectID,
			})
		}
	}
	return out, nil
}

func (s *stubRepo) IncrementRuleExecution(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubRepo) CreateExecutionRecord(_ context.Context, p sqlc.CreateRuleExecutionRecordParams) error {
	s.executions = append(s.executions, p)
	return nil
}
func (s *stubRepo) GetTaskScope(_ context.Context, taskID uuid.UUID) (*sqlc.GetTaskForAutomationActionRow, error) {
	return &sqlc.GetTaskForAutomationActionRow{ID: taskID, ProjectID: uuid.New()}, nil
}
func (s *stubRepo) ReplaceTaskAssignees(_ context.Context, _ uuid.UUID, ids []uuid.UUID) error {
	s.assignedIDs = ids
	return nil
}
func (s *stubRepo) UpdateTaskStatus(_ context.Context, taskID, statusID uuid.UUID) (*sqlc.UpdateTaskStatusForAutomationRow, *sqlc.GetTaskStatusSummaryForAutomationRow, error) {
	s.statusSet = statusID
	return &sqlc.UpdateTaskStatusForAutomationRow{ID: taskID, ProjectID: uuid.New(), StatusID: statusID}, &sqlc.GetTaskStatusSummaryForAutomationRow{ID: statusID, Name: "Done"}, nil
}
func (s *stubRepo) AddTaskLabel(_ context.Context, _, _ uuid.UUID) error { return nil }
func (s *stubRepo) AddTaskComment(_ context.Context, taskID, authorID uuid.UUID, content string) (*sqlc.InsertTaskCommentForAutomationRow, error) {
	s.commentAdded = content
	return &sqlc.InsertTaskCommentForAutomationRow{ID: uuid.New(), TaskID: taskID, AuthorID: authorID, Content: content}, nil
}
func (s *stubRepo) UpdateTaskPriority(_ context.Context, _ uuid.UUID, _ sqlc.TaskPriority) error {
	return nil
}
func (s *stubRepo) UpdateTaskDueDate(_ context.Context, _ uuid.UUID, _ time.Time) error { return nil }

type stubEnqueuer struct {
	payloads []ExecutionJobPayload
}

func (e *stubEnqueuer) Enqueue(_ context.Context, _, _ string, payload any) error {
	raw, _ := json.Marshal(payload)
	var job ExecutionJobPayload
	_ = json.Unmarshal(raw, &job)
	e.payloads = append(e.payloads, job)
	return nil
}

func TestAutomationTriggerDispatchAndExecuteJob(t *testing.T) {
	repo := &stubRepo{rules: make(map[uuid.UUID]sqlc.GetAutomationRuleDetailByIDRow)}
	enq := &stubEnqueuer{}
	svc := NewService(repo, enq, nil, nil, nil)

	userID := uuid.New()
	projID := uuid.New()
	statusID := uuid.New()
	projStr := projID.String()

	rule, err := svc.Create(context.Background(), CreateRuleRequest{
		Name:          "Move bugs to Triage",
		TriggerType:   string(sqlc.TriggerTypeTASKCREATED),
		TriggerConfig: map[string]any{"taskType": "BUG"},
		ActionType:    string(sqlc.ActionTypeCHANGESTATUS),
		ActionConfig:  map[string]any{"statusId": statusID.String()},
		ProjectID:     &projStr,
	}, userID)
	require.NoError(t, err)
	assert.Equal(t, "ACTIVE", rule.Status)

	taskID := uuid.New()
	svc.HandleTaskCreated(context.Background(), taskID, projID, map[string]any{"type": "FEATURE"}, &userID)
	assert.Empty(t, enq.payloads, "non-matching taskType must not trigger")

	svc.HandleTaskCreated(context.Background(), taskID, projID, map[string]any{"type": "BUG"}, &userID)
	require.Len(t, enq.payloads, 1)

	require.NoError(t, svc.ExecuteJob(context.Background(), enq.payloads[0]))
	assert.Equal(t, statusID, repo.statusSet)
	require.Len(t, repo.executions, 1)
	assert.True(t, repo.executions[0].Success)

	toggled, err := svc.ToggleStatus(context.Background(), rule.ID)
	require.NoError(t, err)
	assert.Equal(t, "INACTIVE", toggled.Status)

	stats, err := svc.GetExecutionStats(context.Background(), rule.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), stats.TotalExecutions)
	assert.Equal(t, 50.0, stats.SuccessRate)
}

func TestAutomationHandlerMountAndStats(t *testing.T) {
	repo := &stubRepo{rules: make(map[uuid.UUID]sqlc.GetAutomationRuleDetailByIDRow)}
	svc := NewService(repo, nil, nil, nil, nil)
	h := NewHandler(svc)

	userID := uuid.New()
	rule, err := svc.Create(context.Background(), CreateRuleRequest{
		Name:        "Test rule",
		TriggerType: string(sqlc.TriggerTypeTASKCREATED),
		ActionType:  string(sqlc.ActionTypeSENDNOTIFICATION),
	}, userID)
	require.NoError(t, err)

	r := chi.NewRouter()
	MountRoutes(r, h, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(w, req.WithContext(ctxhelper.WithUserID(req.Context(), userID.String())))
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/automation/rules/"+rule.ID.String()+"/stats", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	delReq := httptest.NewRequest(http.MethodDelete, "/automation/rules/"+rule.ID.String(), nil)
	delRec := httptest.NewRecorder()
	r.ServeHTTP(delRec, delReq)
	assert.Equal(t, http.StatusNoContent, delRec.Code)
	assert.Equal(t, queue.QueueAutomation, "automation")
}
