package notifications

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/sanitize"
)

// EmailSender queues transactional emails asynchronously without blocking HTTP handlers.
type EmailSender interface {
	SendAsync(to, subject, template string, data map[string]any)
}

var (
	usernameMentionRegex = regexp.MustCompile(`(?:^|\s|>|\[)@([\w.-]+)\b`)
	memberURLMentionRe   = regexp.MustCompile(`(?i)/members/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)
	bracketMentionRe     = regexp.MustCompile(`(?i)@\[mention:([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\]`)
)

// CommentHooks implements task_comments.NotificationHooks to dispatch mention and participant alerts.
type CommentHooks struct {
	svc         Service
	repo        Repository
	email       EmailSender
	frontendURL string
}

// NewCommentHooks constructs comment notification hooks backed by notifications.Service.
func NewCommentHooks(svc Service, repo Repository, email EmailSender, frontendURL string) *CommentHooks {
	if frontendURL == "" {
		frontendURL = "http://localhost:3001"
	}
	return &CommentHooks{
		svc:         svc,
		repo:        repo,
		email:       email,
		frontendURL: strings.TrimRight(frontendURL, "/"),
	}
}

func extractMentions(content string) (usernames []string, userIDs []uuid.UUID) {
	uSet := make(map[string]struct{})
	for _, m := range usernameMentionRegex.FindAllStringSubmatch(content, -1) {
		if len(m) > 1 && m[1] != "" {
			uSet[m[1]] = struct{}{}
		}
	}
	for u := range uSet {
		usernames = append(usernames, u)
	}

	idSet := make(map[uuid.UUID]struct{})
	for _, re := range []*regexp.Regexp{memberURLMentionRe, bracketMentionRe} {
		for _, m := range re.FindAllStringSubmatch(content, -1) {
			if len(m) > 1 {
				if parsed, err := uuid.Parse(m[1]); err == nil {
					idSet[parsed] = struct{}{}
				}
			}
		}
	}
	for id := range idSet {
		userIDs = append(userIDs, id)
	}
	return usernames, userIDs
}

func diffStringSlice(current, previous []string) []string {
	prev := make(map[string]struct{}, len(previous))
	for _, p := range previous {
		prev[p] = struct{}{}
	}
	var out []string
	for _, c := range current {
		if _, ok := prev[c]; !ok {
			out = append(out, c)
		}
	}
	return out
}

func diffUUIDSlice(current, previous []uuid.UUID) []uuid.UUID {
	prev := make(map[uuid.UUID]struct{}, len(previous))
	for _, p := range previous {
		prev[p] = struct{}{}
	}
	var out []uuid.UUID
	for _, c := range current {
		if _, ok := prev[c]; !ok {
			out = append(out, c)
		}
	}
	return out
}

// OnCommentCreated dispatches @mention notifications/emails and notifies task participants.
func (h *CommentHooks) OnCommentCreated(ctx context.Context, _ uuid.UUID, taskID, authorID uuid.UUID, content string) {
	h.handle(ctx, taskID, authorID, content, "")
}

// OnCommentUpdated dispatches notifications only for newly added @mentions.
func (h *CommentHooks) OnCommentUpdated(ctx context.Context, _ uuid.UUID, taskID, authorID uuid.UUID, oldContent, newContent string) {
	h.handle(ctx, taskID, authorID, newContent, oldContent)
}

func (h *CommentHooks) handle(ctx context.Context, taskID, authorID uuid.UUID, content, oldContent string) {
	if h.repo == nil || h.svc == nil {
		return
	}

	taskCtx, err := h.repo.GetTaskContext(ctx, taskID)
	if err != nil || taskCtx == nil {
		return
	}

	authorFirst := "Someone"
	authorFull := "Someone"
	if author, aErr := h.repo.GetUserBrief(ctx, authorID); aErr == nil && author != nil {
		if author.FirstName != "" {
			authorFirst = author.FirstName
		}
		authorFull = strings.TrimSpace(author.FirstName + " " + author.LastName)
		if authorFull == "" {
			authorFull = authorFirst
		}
	}

	notified := map[uuid.UUID]struct{}{authorID: {}}
	usernames, userIDs := extractMentions(content)
	if oldContent != "" {
		oldUsernames, oldIDs := extractMentions(oldContent)
		usernames = diffStringSlice(usernames, oldUsernames)
		userIDs = diffUUIDSlice(userIDs, oldIDs)
	}

	entityType := "Task"
	actionURL := "/tasks/" + taskCtx.Slug
	orgID := taskCtx.OrganizationID
	mentionPrio := sqlc.NotificationPriorityHIGH

	if len(usernames) > 0 || len(userIDs) > 0 {
		mentioned, mErr := h.repo.FindMentionedUsers(ctx, usernames, userIDs)
		if mErr == nil {
			for _, u := range mentioned {
				if _, dup := notified[u.ID]; dup {
					continue
				}
				notified[u.ID] = struct{}{}

				_, _ = h.svc.CreateNotification(ctx, CreateInput{
					Title:          "You were mentioned",
					Message:        fmt.Sprintf(`%s mentioned you in a comment on "%s"`, authorFirst, taskCtx.Title),
					Type:           sqlc.NotificationTypeMENTION,
					UserID:         u.ID,
					OrganizationID: &orgID,
					Priority:       &mentionPrio,
					EntityType:     &entityType,
					EntityID:       &taskCtx.ID,
					ActionURL:      &actionURL,
					CreatedBy:      &authorID,
				})

				if h.email != nil && u.Email != "" {
					mentionedName := strings.TrimSpace(u.FirstName + " " + u.LastName)
					if mentionedName == "" && u.Username.Valid {
						mentionedName = u.Username.String
					}
					h.email.SendAsync(
						u.Email,
						"You were mentioned in task: "+sanitize.Text(taskCtx.Title),
						"mention",
						map[string]any{
							"mentioner":     map[string]any{"name": authorFull},
							"mentionedUser": map[string]any{"name": mentionedName},
							"entity":        map[string]any{"title": taskCtx.Title, "key": taskCtx.Slug},
							"textContent":   strings.ReplaceAll(sanitize.Text(content), "&nbsp;", " "),
							"entityUrl":     h.frontendURL + "/tasks/" + taskCtx.Slug,
						},
					)
				}
			}
		}
	}

	if oldContent == "" {
		participants, pErr := h.repo.ListTaskParticipants(ctx, taskID)
		if pErr == nil {
			medPrio := sqlc.NotificationPriorityMEDIUM
			for _, p := range participants {
				if _, dup := notified[p.UserID]; dup {
					continue
				}
				notified[p.UserID] = struct{}{}

				_, _ = h.svc.CreateNotification(ctx, CreateInput{
					Title:          "New Comment",
					Message:        fmt.Sprintf(`%s commented on "%s"`, authorFirst, taskCtx.Title),
					Type:           sqlc.NotificationTypeTASKCOMMENTED,
					UserID:         p.UserID,
					OrganizationID: &orgID,
					Priority:       &medPrio,
					EntityType:     &entityType,
					EntityID:       &taskCtx.ID,
					ActionURL:      &actionURL,
					CreatedBy:      &authorID,
				})
			}
		}
	}
}
