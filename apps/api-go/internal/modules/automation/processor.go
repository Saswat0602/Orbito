package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/platform/queue"
)

// RegisterProcessor binds the automation background worker handler on the queue runtime.
func RegisterProcessor(rt *queue.Runtime, svc Service) {
	rt.RegisterQueue(queue.QueueAutomation)
	rt.RegisterHandler(queue.TaskExecuteAutomationRule, func(ctx context.Context, raw []byte) error {
		var payload ExecutionJobPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return fmt.Errorf("decode automation job payload: %w", err)
		}
		return svc.ExecuteJob(ctx, payload)
	})
}

// ExecuteJob processes one queued automation rule execution and records the outcome.
func (s *service) ExecuteJob(ctx context.Context, payload ExecutionJobPayload) error {
	start := time.Now()
	rule, err := s.repo.GetRuleDetailByID(ctx, payload.RuleID)
	if err != nil {
		return err
	}
	if rule == nil || rule.Status != sqlc.RuleStatusACTIVE {
		return nil
	}

	triggerJSON, _ := json.Marshal(payload.TriggerData)
	actionResult, actErr := s.performAction(ctx, rule, payload.TriggerData)
	elapsed := safeInt32Duration(time.Since(start).Milliseconds())

	trigBy := pgtype.UUID{}
	createdBy := rule.CreatedByID
	if payload.TriggeredByID != nil && *payload.TriggeredByID != uuid.Nil {
		trigBy = pgtype.UUID{Bytes: *payload.TriggeredByID, Valid: true}
		createdBy = trigBy
	}

	if actErr != nil {
		_ = s.repo.CreateExecutionRecord(ctx, sqlc.CreateRuleExecutionRecordParams{
			ID:            uuid.New(),
			Success:       false,
			ErrorMessage:  pgtype.Text{String: actErr.Error(), Valid: true},
			ExecutionTime: elapsed,
			TriggerData:   triggerJSON,
			ActionResult:  nil,
			RuleID:        rule.ID,
			TriggeredByID: trigBy,
			CreatedByID:   createdBy,
		})
		return actErr
	}

	resJSON, _ := json.Marshal(actionResult)
	if incErr := s.repo.IncrementRuleExecution(ctx, rule.ID); incErr != nil {
		s.log.Warn("Failed to increment rule execution count", "error", incErr.Error())
	}
	if recErr := s.repo.CreateExecutionRecord(ctx, sqlc.CreateRuleExecutionRecordParams{
		ID:            uuid.New(),
		Success:       true,
		ExecutionTime: elapsed,
		TriggerData:   triggerJSON,
		ActionResult:  resJSON,
		RuleID:        rule.ID,
		TriggeredByID: trigBy,
		CreatedByID:   createdBy,
	}); recErr != nil {
		return recErr
	}
	return nil
}

func (s *service) performAction(ctx context.Context, rule *sqlc.GetAutomationRuleDetailByIDRow, triggerData map[string]any) (map[string]any, error) {
	actionCfg := decodeJSONMap(rule.ActionConfig)
	if len(actionCfg) == 0 {
		return nil, fmt.Errorf("action configuration is required")
	}

	switch rule.ActionType {
	case sqlc.ActionTypeASSIGNTASK:
		return s.execAssignTask(ctx, triggerData, actionCfg)
	case sqlc.ActionTypeCHANGESTATUS:
		return s.execChangeStatus(ctx, triggerData, actionCfg)
	case sqlc.ActionTypeADDLABEL:
		return s.execAddLabel(ctx, triggerData, actionCfg)
	case sqlc.ActionTypeSENDNOTIFICATION:
		return s.execSendNotification(ctx, triggerData, actionCfg)
	case sqlc.ActionTypeADDCOMMENT:
		return s.execAddComment(ctx, rule, triggerData, actionCfg)
	case sqlc.ActionTypeCHANGEPRIORITY:
		return s.execChangePriority(ctx, triggerData, actionCfg)
	case sqlc.ActionTypeSETDUEDATE:
		return s.execSetDueDate(ctx, triggerData, actionCfg)
	default:
		return nil, fmt.Errorf("unsupported action type: %s", rule.ActionType)
	}
}

func (s *service) execAssignTask(ctx context.Context, triggerData, cfg map[string]any) (map[string]any, error) {
	taskID, err := extractTaskUUID(triggerData)
	if err != nil {
		return nil, err
	}
	assigneeIDs := extractUUIDSlice(cfg, "assigneeId")
	if len(assigneeIDs) == 0 {
		assigneeIDs = extractUUIDSlice(cfg, "assigneeIds")
	}
	if err := s.repo.ReplaceTaskAssignees(ctx, taskID, assigneeIDs); err != nil {
		return nil, err
	}
	taskScope, _ := s.repo.GetTaskScope(ctx, taskID)
	if s.emitter != nil && taskScope != nil {
		for _, aid := range assigneeIDs {
			s.emitter.EmitTaskAssigned(taskScope.ProjectID.String(), taskID.String(), map[string]any{"assigneeId": aid.String()})
		}
	}
	return map[string]any{"success": true, "taskId": taskID.String(), "assigneeIds": uuidsToStrings(assigneeIDs)}, nil
}

func (s *service) execChangeStatus(ctx context.Context, triggerData, cfg map[string]any) (map[string]any, error) {
	taskID, err := extractTaskUUID(triggerData)
	if err != nil {
		return nil, err
	}
	statusID, err := extractConfigUUID(cfg, "statusId")
	if err != nil {
		return nil, err
	}
	updated, statusRow, err := s.repo.UpdateTaskStatus(ctx, taskID, statusID)
	if err != nil {
		return nil, err
	}
	if s.emitter != nil && updated != nil {
		s.emitter.EmitTaskStatusChanged(updated.ProjectID.String(), taskID.String(), map[string]any{
			"statusId": statusID.String(),
			"status":   statusRow,
		})
	}
	res := map[string]any{"success": true, "taskId": taskID.String(), "statusId": statusID.String()}
	if statusRow != nil {
		res["statusName"] = statusRow.Name
	}
	return res, nil
}

func (s *service) execAddLabel(ctx context.Context, triggerData, cfg map[string]any) (map[string]any, error) {
	taskID, err := extractTaskUUID(triggerData)
	if err != nil {
		return nil, err
	}
	labelID, err := extractConfigUUID(cfg, "labelId")
	if err != nil {
		return nil, err
	}
	if err := s.repo.AddTaskLabel(ctx, taskID, labelID); err != nil {
		return nil, err
	}
	taskScope, _ := s.repo.GetTaskScope(ctx, taskID)
	if s.emitter != nil && taskScope != nil {
		s.emitter.EmitTaskUpdated(taskScope.ProjectID.String(), taskID.String(), map[string]any{"labelAdded": labelID.String()})
	}
	return map[string]any{"success": true, "taskId": taskID.String(), "labelId": labelID.String()}, nil
}

func (s *service) execSendNotification(ctx context.Context, triggerData, cfg map[string]any) (map[string]any, error) {
	userID, err := extractConfigUUID(cfg, "userId")
	if err != nil {
		return nil, err
	}
	msg, _ := cfg["message"].(string)
	if msg == "" {
		msg = "Automation rule triggered"
	}
	if s.notifier != nil {
		if nErr := s.notifier.CreateAutomationNotification(ctx, userID, "Automation Notification", msg); nErr != nil {
			s.log.Warn("Failed to persist automation notification", "error", nErr.Error())
		}
	}
	if s.emitter != nil {
		s.emitter.EmitNotification(userID.String(), map[string]any{
			"type":    "automation",
			"message": msg,
			"data":    triggerData,
		})
	}
	return map[string]any{"success": true, "userId": userID.String(), "message": msg}, nil
}

func (s *service) execAddComment(ctx context.Context, rule *sqlc.GetAutomationRuleDetailByIDRow, triggerData, cfg map[string]any) (map[string]any, error) {
	taskID, err := extractTaskUUID(triggerData)
	if err != nil {
		return nil, err
	}
	content, _ := cfg["content"].(string)
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("comment content is required")
	}
	if !rule.CreatedByID.Valid {
		return nil, fmt.Errorf("rule creator is required to post automation comment")
	}
	authorID := uuid.UUID(rule.CreatedByID.Bytes)
	comment, err := s.repo.AddTaskComment(ctx, taskID, authorID, content)
	if err != nil {
		return nil, err
	}
	taskScope, _ := s.repo.GetTaskScope(ctx, taskID)
	if s.emitter != nil && taskScope != nil {
		s.emitter.EmitCommentAdded(taskScope.ProjectID.String(), taskID.String(), map[string]any{
			"commentId": comment.ID.String(),
			"content":   comment.Content,
		})
	}
	return map[string]any{"success": true, "taskId": taskID.String(), "commentId": comment.ID.String()}, nil
}

func (s *service) execChangePriority(ctx context.Context, triggerData, cfg map[string]any) (map[string]any, error) {
	taskID, err := extractTaskUUID(triggerData)
	if err != nil {
		return nil, err
	}
	prioRaw, _ := cfg["priority"].(string)
	prio := sqlc.TaskPriority(strings.ToUpper(strings.TrimSpace(prioRaw)))
	switch prio {
	case sqlc.TaskPriorityLOWEST, sqlc.TaskPriorityLOW, sqlc.TaskPriorityMEDIUM, sqlc.TaskPriorityHIGH, sqlc.TaskPriorityHIGHEST:
	default:
		return nil, fmt.Errorf("invalid priority %q", prioRaw)
	}
	if err := s.repo.UpdateTaskPriority(ctx, taskID, prio); err != nil {
		return nil, err
	}
	taskScope, _ := s.repo.GetTaskScope(ctx, taskID)
	if s.emitter != nil && taskScope != nil {
		s.emitter.EmitTaskUpdated(taskScope.ProjectID.String(), taskID.String(), map[string]any{"priority": string(prio)})
	}
	return map[string]any{"success": true, "taskId": taskID.String(), "priority": string(prio)}, nil
}

func (s *service) execSetDueDate(ctx context.Context, triggerData, cfg map[string]any) (map[string]any, error) {
	taskID, err := extractTaskUUID(triggerData)
	if err != nil {
		return nil, err
	}
	rawDate, _ := cfg["dueDate"].(string)
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(rawDate))
	if err != nil {
		parsed, err = time.Parse("2006-01-02", strings.TrimSpace(rawDate))
		if err != nil {
			return nil, fmt.Errorf("invalid dueDate %q", rawDate)
		}
	}
	if err := s.repo.UpdateTaskDueDate(ctx, taskID, parsed); err != nil {
		return nil, err
	}
	taskScope, _ := s.repo.GetTaskScope(ctx, taskID)
	if s.emitter != nil && taskScope != nil {
		s.emitter.EmitTaskUpdated(taskScope.ProjectID.String(), taskID.String(), map[string]any{"dueDate": parsed.UTC()})
	}
	return map[string]any{"success": true, "taskId": taskID.String(), "dueDate": parsed.UTC()}, nil
}

func checkTriggerConfig(rule *sqlc.AutomationRule, data map[string]any) bool {
	cfg := decodeJSONMap(rule.TriggerConfig)
	if len(cfg) == 0 {
		return true
	}
	taskMap, _ := data["task"].(map[string]any)
	checks := [][2]string{
		{"taskType", "type"},
		{"priority", "priority"},
		{"projectId", "projectId"},
		{"assigneeId", "assigneeId"},
	}
	for _, pair := range checks {
		want, ok := cfg[pair[0]]
		if !ok || want == nil || want == "" {
			continue
		}
		if taskMap == nil || !reflect.DeepEqual(taskMap[pair[1]], want) {
			return false
		}
	}
	return true
}

func extractTaskUUID(data map[string]any) (uuid.UUID, error) {
	raw, _ := data["taskId"].(string)
	if raw == "" {
		return uuid.Nil, fmt.Errorf("taskId is required in trigger data")
	}
	return uuid.Parse(raw)
}

func extractConfigUUID(cfg map[string]any, key string) (uuid.UUID, error) {
	raw, _ := cfg[key].(string)
	if raw == "" {
		return uuid.Nil, fmt.Errorf("%s is required in action config", key)
	}
	return uuid.Parse(raw)
}

func extractUUIDSlice(cfg map[string]any, key string) []uuid.UUID {
	val, exists := cfg[key]
	if !exists || val == nil {
		return nil
	}
	switch v := val.(type) {
	case string:
		if id, err := uuid.Parse(v); err == nil {
			return []uuid.UUID{id}
		}
	case []string:
		return parseUUIDList(v)
	case []any:
		strs := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				strs = append(strs, s)
			}
		}
		return parseUUIDList(strs)
	}
	return nil
}

func parseUUIDList(items []string) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(items))
	for _, s := range items {
		if id, err := uuid.Parse(strings.TrimSpace(s)); err == nil {
			out = append(out, id)
		}
	}
	return out
}

func uuidsToStrings(ids []uuid.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}
