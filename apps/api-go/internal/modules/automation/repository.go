package automation

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// Repository encapsulates persistence for automation rules, executions, and task side-effects.
type Repository interface {
	CreateRule(ctx context.Context, params sqlc.CreateAutomationRuleParams) (uuid.UUID, error)
	GetRuleDetailByID(ctx context.Context, id uuid.UUID) (*sqlc.GetAutomationRuleDetailByIDRow, error)
	ListRules(ctx context.Context, projectID, workspaceID, orgID pgtype.UUID) ([]sqlc.ListAutomationRulesFilteredRow, error)
	ListRuleExecutions(ctx context.Context, ruleID uuid.UUID, limit int32) ([]sqlc.ListRuleExecutionsByRuleIDRow, error)
	UpdateRule(ctx context.Context, params sqlc.UpdateAutomationRuleByIDParams) (int64, error)
	ToggleRuleStatus(ctx context.Context, id uuid.UUID) (*sqlc.AutomationRule, error)
	DeleteRule(ctx context.Context, id uuid.UUID) (int64, error)
	GetRuleStats(ctx context.Context, ruleID uuid.UUID) (*sqlc.GetAutomationRuleStatsRowRow, error)

	ListActiveRulesByTrigger(ctx context.Context, trigger sqlc.TriggerType, projectID, workspaceID, orgID pgtype.UUID) ([]sqlc.AutomationRule, error)
	IncrementRuleExecution(ctx context.Context, id uuid.UUID) error
	CreateExecutionRecord(ctx context.Context, params sqlc.CreateRuleExecutionRecordParams) error

	GetTaskScope(ctx context.Context, taskID uuid.UUID) (*sqlc.GetTaskForAutomationActionRow, error)
	ReplaceTaskAssignees(ctx context.Context, taskID uuid.UUID, assigneeIDs []uuid.UUID) error
	UpdateTaskStatus(ctx context.Context, taskID, statusID uuid.UUID) (*sqlc.UpdateTaskStatusForAutomationRow, *sqlc.GetTaskStatusSummaryForAutomationRow, error)
	AddTaskLabel(ctx context.Context, taskID, labelID uuid.UUID) error
	AddTaskComment(ctx context.Context, taskID, authorID uuid.UUID, content string) (*sqlc.InsertTaskCommentForAutomationRow, error)
	UpdateTaskPriority(ctx context.Context, taskID uuid.UUID, priority sqlc.TaskPriority) error
	UpdateTaskDueDate(ctx context.Context, taskID uuid.UUID, dueDate time.Time) error
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository creates a PostgreSQL-backed automation repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{
		db:      db,
		queries: sqlc.New(db.Pool()),
	}
}

func (r *pgRepository) CreateRule(ctx context.Context, params sqlc.CreateAutomationRuleParams) (uuid.UUID, error) {
	id, err := r.queries.CreateAutomationRule(ctx, params)
	if err != nil {
		return uuid.Nil, apperror.Map(err)
	}
	return id, nil
}

func (r *pgRepository) GetRuleDetailByID(ctx context.Context, id uuid.UUID) (*sqlc.GetAutomationRuleDetailByIDRow, error) {
	row, err := r.queries.GetAutomationRuleDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListRules(ctx context.Context, projectID, workspaceID, orgID pgtype.UUID) ([]sqlc.ListAutomationRulesFilteredRow, error) {
	return r.queries.ListAutomationRulesFiltered(ctx, sqlc.ListAutomationRulesFilteredParams{
		ProjectID:      projectID,
		WorkspaceID:    workspaceID,
		OrganizationID: orgID,
	})
}

func (r *pgRepository) ListRuleExecutions(ctx context.Context, ruleID uuid.UUID, limit int32) ([]sqlc.ListRuleExecutionsByRuleIDRow, error) {
	return r.queries.ListRuleExecutionsByRuleID(ctx, sqlc.ListRuleExecutionsByRuleIDParams{
		RuleID: ruleID,
		Limit:  limit,
	})
}

func (r *pgRepository) UpdateRule(ctx context.Context, params sqlc.UpdateAutomationRuleByIDParams) (int64, error) {
	return r.queries.UpdateAutomationRuleByID(ctx, params)
}

func (r *pgRepository) ToggleRuleStatus(ctx context.Context, id uuid.UUID) (*sqlc.AutomationRule, error) {
	row, err := r.queries.ToggleAutomationRuleStatus(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) DeleteRule(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.DeleteAutomationRuleByID(ctx, id)
}

func (r *pgRepository) GetRuleStats(ctx context.Context, ruleID uuid.UUID) (*sqlc.GetAutomationRuleStatsRowRow, error) {
	row, err := r.queries.GetAutomationRuleStatsRow(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListActiveRulesByTrigger(ctx context.Context, trigger sqlc.TriggerType, projectID, workspaceID, orgID pgtype.UUID) ([]sqlc.AutomationRule, error) {
	return r.queries.ListActiveAutomationRulesByTrigger(ctx, sqlc.ListActiveAutomationRulesByTriggerParams{
		TriggerType:    trigger,
		ProjectID:      projectID,
		WorkspaceID:    workspaceID,
		OrganizationID: orgID,
	})
}

func (r *pgRepository) IncrementRuleExecution(ctx context.Context, id uuid.UUID) error {
	return r.queries.IncrementAutomationRuleExecution(ctx, id)
}

func (r *pgRepository) CreateExecutionRecord(ctx context.Context, params sqlc.CreateRuleExecutionRecordParams) error {
	_, err := r.queries.CreateRuleExecutionRecord(ctx, params)
	return err
}

func (r *pgRepository) GetTaskScope(ctx context.Context, taskID uuid.UUID) (*sqlc.GetTaskForAutomationActionRow, error) {
	row, err := r.queries.GetTaskForAutomationAction(ctx, taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ReplaceTaskAssignees(ctx context.Context, taskID uuid.UUID, assigneeIDs []uuid.UUID) error {
	return r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		if err := q.DeleteTaskAssigneesForAutomation(ctx, taskID); err != nil {
			return err
		}
		for _, uid := range assigneeIDs {
			if err := q.InsertTaskAssigneeForAutomation(ctx, sqlc.InsertTaskAssigneeForAutomationParams{
				TaskID: taskID,
				UserID: uid,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *pgRepository) UpdateTaskStatus(ctx context.Context, taskID, statusID uuid.UUID) (*sqlc.UpdateTaskStatusForAutomationRow, *sqlc.GetTaskStatusSummaryForAutomationRow, error) {
	updated, err := r.queries.UpdateTaskStatusForAutomation(ctx, sqlc.UpdateTaskStatusForAutomationParams{
		ID:       taskID,
		StatusID: statusID,
	})
	if err != nil {
		return nil, nil, apperror.Map(err)
	}
	statusRow, stErr := r.queries.GetTaskStatusSummaryForAutomation(ctx, statusID)
	if stErr != nil {
		return &updated, nil, nil
	}
	return &updated, &statusRow, nil
}

func (r *pgRepository) AddTaskLabel(ctx context.Context, taskID, labelID uuid.UUID) error {
	return r.queries.InsertTaskLabelForAutomation(ctx, sqlc.InsertTaskLabelForAutomationParams{
		TaskID:  taskID,
		LabelID: labelID,
	})
}

func (r *pgRepository) AddTaskComment(ctx context.Context, taskID, authorID uuid.UUID, content string) (*sqlc.InsertTaskCommentForAutomationRow, error) {
	row, err := r.queries.InsertTaskCommentForAutomation(ctx, sqlc.InsertTaskCommentForAutomationParams{
		ID:       uuid.New(),
		TaskID:   taskID,
		AuthorID: authorID,
		Content:  content,
	})
	if err != nil {
		return nil, apperror.Map(err)
	}
	return &row, nil
}

func (r *pgRepository) UpdateTaskPriority(ctx context.Context, taskID uuid.UUID, priority sqlc.TaskPriority) error {
	_, err := r.queries.UpdateTaskPriorityForAutomation(ctx, sqlc.UpdateTaskPriorityForAutomationParams{
		ID:       taskID,
		Priority: priority,
	})
	return err
}

func (r *pgRepository) UpdateTaskDueDate(ctx context.Context, taskID uuid.UUID, dueDate time.Time) error {
	_, err := r.queries.UpdateTaskDueDateForAutomation(ctx, sqlc.UpdateTaskDueDateForAutomationParams{
		ID:      taskID,
		DueDate: pgtype.Timestamp{Time: dueDate.UTC(), Valid: true},
	})
	return err
}

func decodeJSONMap(raw []byte) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func encodeJSONBytes(m map[string]any) ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}
