package automation

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/platform/queue"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// QueueEnqueuer abstracts background job enqueueing.
type QueueEnqueuer interface {
	Enqueue(ctx context.Context, queueName, taskType string, payload any) error
}

// EventEmitter abstracts WebSocket broadcasts triggered by automation actions.
type EventEmitter interface {
	EmitTaskAssigned(projectID, taskID string, payload any)
	EmitTaskStatusChanged(projectID, taskID string, payload any)
	EmitTaskUpdated(projectID, taskID string, payload any)
	EmitCommentAdded(projectID, taskID string, payload any)
	EmitNotification(userID string, payload any)
}

// Notifier abstracts persisting notifications created by SEND_NOTIFICATION actions.
type Notifier interface {
	CreateAutomationNotification(ctx context.Context, userID uuid.UUID, title, message string) error
}

// Service exposes automation rule CRUD, execution stats, event dispatching, and worker execution.
type Service interface {
	Create(ctx context.Context, req CreateRuleRequest, callerID uuid.UUID) (*RuleResponse, error)
	FindAll(ctx context.Context, orgID, workspaceID, projectID string) ([]RuleResponse, error)
	FindOne(ctx context.Context, id uuid.UUID) (*RuleResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRuleRequest) (*RuleResponse, error)
	ToggleStatus(ctx context.Context, id uuid.UUID) (*RuleResponse, error)
	Remove(ctx context.Context, id uuid.UUID) error
	GetExecutionStats(ctx context.Context, id uuid.UUID) (*RuleStatsResponse, error)

	HandleTaskCreated(ctx context.Context, taskID, projectID uuid.UUID, taskData map[string]any, reporterID *uuid.UUID)
	HandleTaskUpdated(ctx context.Context, taskID, projectID uuid.UUID, updates map[string]any, userID uuid.UUID)
	HandleTaskStatusChanged(ctx context.Context, taskID, projectID uuid.UUID, statusChange map[string]any, userID uuid.UUID)
	HandleTaskAssigned(ctx context.Context, taskID, projectID uuid.UUID, assigneeID string, userID uuid.UUID)
	HandleSprintEvent(ctx context.Context, triggerType string, projectID uuid.UUID, sprintData map[string]any, userID uuid.UUID)
	HandleCommentAdded(ctx context.Context, taskID, projectID uuid.UUID, commentData map[string]any, userID uuid.UUID)
	ExecuteJob(ctx context.Context, payload ExecutionJobPayload) error
}

type service struct {
	repo     Repository
	enqueuer QueueEnqueuer
	emitter  EventEmitter
	notifier Notifier
	log      *slog.Logger
}

// NewService constructs an automation service.
func NewService(repo Repository, enqueuer QueueEnqueuer, emitter EventEmitter, notifier Notifier, log *slog.Logger) Service {
	if log == nil {
		log = slog.Default()
	}
	return &service{
		repo:     repo,
		enqueuer: enqueuer,
		emitter:  emitter,
		notifier: notifier,
		log:      log,
	}
}

func (s *service) Create(ctx context.Context, req CreateRuleRequest, callerID uuid.UUID) (*RuleResponse, error) {
	trig, ok := parseTriggerType(req.TriggerType)
	if !ok {
		return nil, apperror.BadRequest("Invalid triggerType")
	}
	act, ok := parseActionType(req.ActionType)
	if !ok {
		return nil, apperror.BadRequest("Invalid actionType")
	}
	status := sqlc.RuleStatusACTIVE
	if req.Status != nil && *req.Status != "" {
		parsedStatus, stOK := parseRuleStatus(*req.Status)
		if !stOK {
			return nil, apperror.BadRequest("Invalid rule status")
		}
		status = parsedStatus
	}

	creator := callerID
	if req.CreatedBy != nil && *req.CreatedBy != "" {
		parsedCreator, err := uuid.Parse(*req.CreatedBy)
		if err != nil {
			return nil, apperror.BadRequest("createdBy must be a valid UUID")
		}
		creator = parsedCreator
	}

	orgUUID, err := parseOptionalPgUUID(req.OrganizationID, "organizationId")
	if err != nil {
		return nil, err
	}
	wsUUID, err := parseOptionalPgUUID(req.WorkspaceID, "workspaceId")
	if err != nil {
		return nil, err
	}
	projUUID, err := parseOptionalPgUUID(req.ProjectID, "projectId")
	if err != nil {
		return nil, err
	}

	trigJSON, err := encodeJSONBytes(req.TriggerConfig)
	if err != nil {
		return nil, apperror.BadRequest("Invalid triggerConfig JSON")
	}
	actJSON, err := encodeJSONBytes(req.ActionConfig)
	if err != nil {
		return nil, apperror.BadRequest("Invalid actionConfig JSON")
	}

	id, err := s.repo.CreateRule(ctx, sqlc.CreateAutomationRuleParams{
		ID:             uuid.New(),
		Name:           strings.TrimSpace(req.Name),
		Description:    toPgText(req.Description),
		Status:         status,
		TriggerType:    trig,
		TriggerConfig:  trigJSON,
		ActionType:     act,
		ActionConfig:   actJSON,
		OrganizationID: orgUUID,
		WorkspaceID:    wsUUID,
		ProjectID:      projUUID,
		CreatedByID:    pgtype.UUID{Bytes: creator, Valid: creator != uuid.Nil},
	})
	if err != nil {
		return nil, err
	}
	return s.FindOne(ctx, id)
}

func (s *service) FindAll(ctx context.Context, orgID, workspaceID, projectID string) ([]RuleResponse, error) {
	var projFilter, wsFilter, orgFilter pgtype.UUID
	switch {
	case projectID != "":
		parsed, err := uuid.Parse(projectID)
		if err != nil {
			return nil, apperror.BadRequest("projectId must be a valid UUID")
		}
		projFilter = pgtype.UUID{Bytes: parsed, Valid: true}
	case workspaceID != "":
		parsed, err := uuid.Parse(workspaceID)
		if err != nil {
			return nil, apperror.BadRequest("workspaceId must be a valid UUID")
		}
		wsFilter = pgtype.UUID{Bytes: parsed, Valid: true}
	case orgID != "":
		parsed, err := uuid.Parse(orgID)
		if err != nil {
			return nil, apperror.BadRequest("organizationId must be a valid UUID")
		}
		orgFilter = pgtype.UUID{Bytes: parsed, Valid: true}
	}

	rows, err := s.repo.ListRules(ctx, projFilter, wsFilter, orgFilter)
	if err != nil {
		return nil, fmt.Errorf("list automation rules: %w", err)
	}
	out := make([]RuleResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapListRow(&rows[i]))
	}
	return out, nil
}

func (s *service) FindOne(ctx context.Context, id uuid.UUID) (*RuleResponse, error) {
	detail, err := s.repo.GetRuleDetailByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get automation rule: %w", err)
	}
	if detail == nil {
		return nil, apperror.NotFound("Automation rule not found")
	}

	execRows, execErr := s.repo.ListRuleExecutions(ctx, id, 10)
	if execErr != nil {
		return nil, fmt.Errorf("list rule executions: %w", execErr)
	}
	resp := mapDetailRow(detail, execRows)
	return &resp, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateRuleRequest) (*RuleResponse, error) {
	if _, err := s.FindOne(ctx, id); err != nil {
		return nil, err
	}

	params := sqlc.UpdateAutomationRuleByIDParams{
		ID:          id,
		Name:        toPgText(req.Name),
		Description: toPgText(req.Description),
	}
	if req.Status != nil {
		st, ok := parseRuleStatus(*req.Status)
		if !ok {
			return nil, apperror.BadRequest("Invalid rule status")
		}
		params.Status = sqlc.NullRuleStatus{RuleStatus: st, Valid: true}
	}
	if req.TriggerType != nil {
		trig, ok := parseTriggerType(*req.TriggerType)
		if !ok {
			return nil, apperror.BadRequest("Invalid triggerType")
		}
		params.TriggerType = sqlc.NullTriggerType{TriggerType: trig, Valid: true}
	}
	if req.ActionType != nil {
		act, ok := parseActionType(*req.ActionType)
		if !ok {
			return nil, apperror.BadRequest("Invalid actionType")
		}
		params.ActionType = sqlc.NullActionType{ActionType: act, Valid: true}
	}
	if req.TriggerConfig != nil {
		b, err := encodeJSONBytes(req.TriggerConfig)
		if err != nil {
			return nil, apperror.BadRequest("Invalid triggerConfig JSON")
		}
		params.TriggerConfig = b
	}
	if req.ActionConfig != nil {
		b, err := encodeJSONBytes(req.ActionConfig)
		if err != nil {
			return nil, apperror.BadRequest("Invalid actionConfig JSON")
		}
		params.ActionConfig = b
	}

	affected, err := s.repo.UpdateRule(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update automation rule: %w", err)
	}
	if affected == 0 {
		return nil, apperror.NotFound("Automation rule not found")
	}
	return s.FindOne(ctx, id)
}

func (s *service) ToggleStatus(ctx context.Context, id uuid.UUID) (*RuleResponse, error) {
	toggled, err := s.repo.ToggleRuleStatus(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("toggle automation rule: %w", err)
	}
	if toggled == nil {
		return nil, apperror.NotFound("Automation rule not found")
	}
	return s.FindOne(ctx, id)
}

func (s *service) Remove(ctx context.Context, id uuid.UUID) error {
	affected, err := s.repo.DeleteRule(ctx, id)
	if err != nil {
		return fmt.Errorf("delete automation rule: %w", err)
	}
	if affected == 0 {
		return apperror.NotFound("Automation rule not found")
	}
	return nil
}

func (s *service) GetExecutionStats(ctx context.Context, id uuid.UUID) (*RuleStatsResponse, error) {
	rule, err := s.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	stats, err := s.repo.GetRuleStats(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get automation stats: %w", err)
	}
	failureCount := stats.TotalExecutions - stats.SuccessCount
	var successRate float64
	if stats.TotalExecutions > 0 {
		successRate = float64(stats.SuccessCount) * 100.0 / float64(stats.TotalExecutions)
	}
	return &RuleStatsResponse{
		RuleID:           rule.ID,
		RuleName:         rule.Name,
		TotalExecutions:  stats.TotalExecutions,
		SuccessCount:     stats.SuccessCount,
		FailureCount:     failureCount,
		SuccessRate:      successRate,
		AvgExecutionTime: stats.AvgExecutionTime,
		LastExecuted:     rule.LastExecuted,
	}, nil
}

func (s *service) HandleTaskCreated(ctx context.Context, taskID, projectID uuid.UUID, taskData map[string]any, reporterID *uuid.UUID) {
	s.dispatchTrigger(ctx, sqlc.TriggerTypeTASKCREATED, &projectID, map[string]any{
		"taskId": taskID.String(),
		"task":   taskData,
	}, reporterID)
}

func (s *service) HandleTaskUpdated(ctx context.Context, taskID, projectID uuid.UUID, updates map[string]any, userID uuid.UUID) {
	s.dispatchTrigger(ctx, sqlc.TriggerTypeTASKUPDATED, &projectID, map[string]any{
		"taskId":  taskID.String(),
		"updates": updates,
	}, &userID)
}

func (s *service) HandleTaskStatusChanged(ctx context.Context, taskID, projectID uuid.UUID, statusChange map[string]any, userID uuid.UUID) {
	s.dispatchTrigger(ctx, sqlc.TriggerTypeTASKSTATUSCHANGED, &projectID, map[string]any{
		"taskId":       taskID.String(),
		"statusChange": statusChange,
	}, &userID)
}

func (s *service) HandleTaskAssigned(ctx context.Context, taskID, projectID uuid.UUID, assigneeID string, userID uuid.UUID) {
	s.dispatchTrigger(ctx, sqlc.TriggerTypeTASKASSIGNED, &projectID, map[string]any{
		"taskId":     taskID.String(),
		"assigneeId": assigneeID,
	}, &userID)
}

func (s *service) HandleSprintEvent(ctx context.Context, triggerType string, projectID uuid.UUID, sprintData map[string]any, userID uuid.UUID) {
	trig, ok := parseTriggerType(triggerType)
	if !ok {
		return
	}
	s.dispatchTrigger(ctx, trig, &projectID, map[string]any{
		"projectId": projectID.String(),
		"sprint":    sprintData,
	}, &userID)
}

func (s *service) HandleCommentAdded(ctx context.Context, taskID, projectID uuid.UUID, commentData map[string]any, userID uuid.UUID) {
	s.dispatchTrigger(ctx, sqlc.TriggerTypeCOMMENTADDED, &projectID, map[string]any{
		"taskId":    taskID.String(),
		"projectId": projectID.String(),
		"comment":   commentData,
	}, &userID)
}

func (s *service) dispatchTrigger(ctx context.Context, trig sqlc.TriggerType, projectID *uuid.UUID, triggerData map[string]any, actorID *uuid.UUID) {
	if s.enqueuer == nil {
		return
	}
	var projFilter pgtype.UUID
	if projectID != nil && *projectID != uuid.Nil {
		projFilter = pgtype.UUID{Bytes: *projectID, Valid: true}
	}
	rules, err := s.repo.ListActiveRulesByTrigger(ctx, trig, projFilter, pgtype.UUID{}, pgtype.UUID{})
	if err != nil {
		s.log.Warn("Failed to query active automation rules", slog.String("trigger", string(trig)), slog.String("error", err.Error()))
		return
	}
	for i := range rules {
		rule := rules[i]
		if !checkTriggerConfig(&rule, triggerData) {
			continue
		}
		payload := ExecutionJobPayload{
			RuleID:        rule.ID,
			TriggerType:   string(trig),
			TriggerData:   triggerData,
			TriggeredByID: actorID,
		}
		if enqErr := s.enqueuer.Enqueue(ctx, queue.QueueAutomation, queue.TaskExecuteAutomationRule, payload); enqErr != nil {
			s.log.Warn("Failed to enqueue automation rule execution", slog.String("ruleId", rule.ID.String()), slog.String("error", enqErr.Error()))
		}
	}
}

func parseTriggerType(raw string) (sqlc.TriggerType, bool) {
	val := sqlc.TriggerType(strings.TrimSpace(raw))
	switch val {
	case sqlc.TriggerTypeTASKCREATED,
		sqlc.TriggerTypeTASKUPDATED,
		sqlc.TriggerTypeTASKSTATUSCHANGED,
		sqlc.TriggerTypeTASKASSIGNED,
		sqlc.TriggerTypeTASKDUEDATEAPPROACHING,
		sqlc.TriggerTypeTASKOVERDUE,
		sqlc.TriggerTypeSPRINTSTARTED,
		sqlc.TriggerTypeSPRINTCOMPLETED,
		sqlc.TriggerTypePROJECTCREATED,
		sqlc.TriggerTypeCOMMENTADDED:
		return val, true
	default:
		return "", false
	}
}

func parseActionType(raw string) (sqlc.ActionType, bool) {
	val := sqlc.ActionType(strings.TrimSpace(raw))
	switch val {
	case sqlc.ActionTypeASSIGNTASK,
		sqlc.ActionTypeCHANGESTATUS,
		sqlc.ActionTypeADDLABEL,
		sqlc.ActionTypeREMOVELABEL,
		sqlc.ActionTypeSETDUEDATE,
		sqlc.ActionTypeSENDNOTIFICATION,
		sqlc.ActionTypeSENDEMAIL,
		sqlc.ActionTypeADDCOMMENT,
		sqlc.ActionTypeMOVETOSPRINT,
		sqlc.ActionTypeCHANGEPRIORITY:
		return val, true
	default:
		return "", false
	}
}

func parseRuleStatus(raw string) (sqlc.RuleStatus, bool) {
	val := sqlc.RuleStatus(strings.TrimSpace(raw))
	switch val {
	case sqlc.RuleStatusACTIVE, sqlc.RuleStatusINACTIVE, sqlc.RuleStatusDRAFT:
		return val, true
	default:
		return "", false
	}
}

func parseOptionalPgUUID(raw *string, fieldName string) (pgtype.UUID, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return pgtype.UUID{}, nil
	}
	parsed, err := uuid.Parse(strings.TrimSpace(*raw))
	if err != nil {
		return pgtype.UUID{}, apperror.BadRequest(fieldName + " must be a valid UUID")
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}, nil
}

func toPgText(raw *string) pgtype.Text {
	if raw == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *raw, Valid: true}
}

func safeInt32Duration(ms int64) int32 {
	if ms < 0 {
		return 0
	}
	if ms > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(ms)
}
