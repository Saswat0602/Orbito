package tasks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

func generateShareToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate share token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

func (s *service) formatShareURL(token string) string {
	base := strings.TrimRight(s.frontendURL, "/")
	if base == "" {
		base = "http://localhost:3000"
	}
	return base + "/public/task/" + token
}

func (s *service) CreateShareLink(ctx context.Context, req CreateShareLinkRequest, userID uuid.UUID, isSuperAdmin bool) (*ShareLinkResponse, error) {
	taskRow, _, err := s.requireTaskAccess(ctx, req.TaskID, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	if !isSuperAdmin && !taskRow.ProjectRole.Valid && !taskRow.WsRole.Valid {
		return nil, apperror.Forbidden("You do not have access to this task")
	}

	token, err := generateShareToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().UTC().AddDate(0, 0, req.ExpiresInDays)
	share, err := s.repo.CreateShareLink(ctx, sqlc.CreatePublicTaskShareParams{
		ID:          uuid.New(),
		Token:       token,
		ExpiresAt:   pgtype.Timestamp{Time: expiresAt, Valid: true},
		TaskID:      taskRow.ID,
		CreatedByID: pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create public task share: %w", err)
	}

	return &ShareLinkResponse{
		ID:        share.ID.String(),
		Token:     share.Token,
		ShareURL:  s.formatShareURL(share.Token),
		ExpiresAt: formatTS(share.ExpiresAt),
		CreatedAt: formatTS(share.CreatedAt),
	}, nil
}

func (s *service) GetSharesForTask(ctx context.Context, taskIDOrSlug string, userID uuid.UUID, isSuperAdmin bool) ([]ShareLinkResponse, error) {
	taskRow, _, err := s.requireTaskAccess(ctx, taskIDOrSlug, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	if !isSuperAdmin && !taskRow.ProjectRole.Valid && !taskRow.WsRole.Valid {
		return nil, apperror.Forbidden("You do not have access to this task")
	}

	rows, err := s.repo.ListActiveShares(ctx, taskRow.ID)
	if err != nil {
		return nil, fmt.Errorf("list active shares: %w", err)
	}
	out := make([]ShareLinkResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, ShareLinkResponse{
			ID:        r.ID.String(),
			Token:     r.Token,
			ShareURL:  s.formatShareURL(r.Token),
			ExpiresAt: formatTS(r.ExpiresAt),
			CreatedAt: formatTS(r.CreatedAt),
		})
	}
	return out, nil
}

func (s *service) RevokeShare(ctx context.Context, shareID uuid.UUID, userID uuid.UUID, isSuperAdmin bool) error {
	row, err := s.repo.GetShareWithAccess(ctx, shareID, userID)
	if err != nil {
		return apperror.NotFound("Share not found")
	}
	if !isSuperAdmin && !row.IsProjectMember && !row.IsWsMember {
		return apperror.Forbidden("You do not have access to this share")
	}
	if err := s.repo.RevokeShare(ctx, shareID); err != nil {
		return fmt.Errorf("revoke share: %w", err)
	}
	return nil
}

