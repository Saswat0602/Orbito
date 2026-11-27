package inbox

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

func (s *service) GetRules(ctx context.Context, projectID uuid.UUID) ([]RuleResponse, error) {
	inbox, err := s.repo.GetInboxByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get inbox: %w", err)
	}
	if inbox == nil {
		return nil, apperror.NotFound("Inbox not found for this project")
	}

	rows, err := s.repo.ListRulesByInboxID(ctx, inbox.ID)
	if err != nil {
		return nil, fmt.Errorf("list inbox rules: %w", err)
	}
	out := make([]RuleResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapRuleRow(&rows[i]))
	}
	return out, nil
}

func (s *service) CreateRule(ctx context.Context, projectID uuid.UUID, req CreateRuleRequest) (*RuleResponse, error) {
	inbox, err := s.repo.GetInboxByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get inbox: %w", err)
	}
	if inbox == nil {
		return nil, apperror.NotFound("Inbox not found for this project")
	}

	condJSON, err := encodeJSONMap(req.Conditions)
	if err != nil {
		return nil, apperror.BadRequest("Invalid conditions JSON")
	}
	actJSON, err := encodeJSONMap(req.Actions)
	if err != nil {
		return nil, apperror.BadRequest("Invalid actions JSON")
	}
	prio := int32(0)
	if req.Priority != nil {
		prio = *req.Priority
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	stop := req.StopOnMatch != nil && *req.StopOnMatch

	row, err := s.repo.CreateRule(ctx, sqlc.CreateInboxRuleRecordParams{
		ID:             uuid.New(),
		ProjectInboxID: inbox.ID,
		Name:           strings.TrimSpace(req.Name),
		Description:    toPgText(req.Description),
		Priority:       prio,
		Enabled:        enabled,
		Conditions:     condJSON,
		Actions:        actJSON,
		StopOnMatch:    stop,
	})
	if err != nil {
		return nil, err
	}
	res := mapRuleRow(&row)
	return &res, nil
}

func (s *service) UpdateRule(ctx context.Context, ruleID uuid.UUID, req CreateRuleRequest) (*RuleResponse, error) {
	params := sqlc.UpdateInboxRuleByIDParams{
		ID:          ruleID,
		Description: toPgText(req.Description),
	}
	if strings.TrimSpace(req.Name) != "" {
		params.Name = pgtype.Text{String: strings.TrimSpace(req.Name), Valid: true}
	}
	if req.Priority != nil {
		params.Priority = pgtype.Int4{Int32: *req.Priority, Valid: true}
	}
	if req.Enabled != nil {
		params.Enabled = pgtype.Bool{Bool: *req.Enabled, Valid: true}
	}
	if req.StopOnMatch != nil {
		params.StopOnMatch = pgtype.Bool{Bool: *req.StopOnMatch, Valid: true}
	}
	if req.Conditions != nil {
		b, err := encodeJSONMap(req.Conditions)
		if err != nil {
			return nil, apperror.BadRequest("Invalid conditions JSON")
		}
		params.Conditions = b
	}
	if req.Actions != nil {
		b, err := encodeJSONMap(req.Actions)
		if err != nil {
			return nil, apperror.BadRequest("Invalid actions JSON")
		}
		params.Actions = b
	}

	row, err := s.repo.UpdateRule(ctx, params)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, apperror.NotFound("Inbox rule not found")
	}
	res := mapRuleRow(row)
	return &res, nil
}

func (s *service) DeleteRule(ctx context.Context, ruleID uuid.UUID) error {
	affected, err := s.repo.DeleteRule(ctx, ruleID)
	if err != nil {
		return fmt.Errorf("delete inbox rule: %w", err)
	}
	if affected == 0 {
		return apperror.NotFound("Inbox rule not found")
	}
	return nil
}

// RuleOutcome summarizes aggregated field overrides from matching inbox rules.
type RuleOutcome struct {
	Matched    bool
	Priority   *sqlc.TaskPriority
	MarkAsSpam bool
	AutoReply  string
}

// EvaluateRules runs enabled rules ordered by descending priority against incoming message attributes.
func EvaluateRules(rules []sqlc.InboxRule, fromEmail, subject, body string) RuleOutcome {
	var outcome RuleOutcome
	for i := range rules {
		r := &rules[i]
		if !r.Enabled {
			continue
		}
		cond := decodeJSONMap(r.Conditions)
		if !matchRuleConditions(cond, fromEmail, subject, body) {
			continue
		}
		outcome.Matched = true
		applyRuleActions(decodeJSONMap(r.Actions), &outcome)
		if r.StopOnMatch {
			break
		}
	}
	return outcome
}

func applyRuleActions(act map[string]any, outcome *RuleOutcome) {
	if p, ok := act["setPriority"].(string); ok && p != "" {
		prio := sqlc.TaskPriority(strings.ToUpper(strings.TrimSpace(p)))
		outcome.Priority = &prio
	}
	if spam, ok := act["markAsSpam"].(bool); ok && spam {
		outcome.MarkAsSpam = true
	}
	if tpl, ok := act["autoReply"].(string); ok && tpl != "" {
		outcome.AutoReply = tpl
	}
}

func matchRuleConditions(cond map[string]any, fromEmail, subject, body string) bool {
	if len(cond) == 0 {
		return true
	}
	if anyRaw, ok := cond["any"].([]any); ok {
		for _, item := range anyRaw {
			if sub, okMap := item.(map[string]any); okMap && matchRuleConditions(sub, fromEmail, subject, body) {
				return true
			}
		}
		return false
	}
	if allRaw, ok := cond["all"].([]any); ok {
		for _, item := range allRaw {
			sub, okMap := item.(map[string]any)
			if !okMap || !matchRuleConditions(sub, fromEmail, subject, body) {
				return false
			}
		}
		return true
	}
	for field, ruleVal := range cond {
		var actual string
		switch field {
		case "from", "fromEmail":
			actual = fromEmail
		case "subject":
			actual = subject
		case "body":
			actual = body
		default:
			continue
		}
		if !evalTextMatcher(actual, ruleVal) {
			return false
		}
	}
	return true
}

func evalTextMatcher(actual string, matcher any) bool {
	actualLower := strings.ToLower(actual)
	m, ok := matcher.(map[string]any)
	if !ok {
		return strings.EqualFold(actual, fmt.Sprint(matcher))
	}
	if sub, exists := m["contains"].(string); exists && !strings.Contains(actualLower, strings.ToLower(sub)) {
		return false
	}
	if eq, exists := m["equals"].(string); exists && !strings.EqualFold(actual, eq) {
		return false
	}
	if pattern, exists := m["matches"].(string); exists && len(pattern) <= 256 {
		re, err := regexp.Compile("(?i)" + pattern)
		if err != nil || !re.MatchString(actual) {
			return false
		}
	}
	return true
}
