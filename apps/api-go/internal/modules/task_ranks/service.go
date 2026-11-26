package task_ranks

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

const rebalanceThreshold = 0.0001

// Service defines business operations for task rank ordering and seeding.
type Service interface {
	Reorder(ctx context.Context, taskID uuid.UUID, req ReorderRequest) error
	Rebalance(ctx context.Context, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType) error
}

type service struct {
	repo Repository
}

// NewService constructs a task_ranks service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ComputeRank calculates a midpoint rank between predecessor (afterTask) and successor (beforeTask).
func ComputeRank(predecessorRank, successorRank *float64) float64 {
	if predecessorRank == nil && successorRank == nil {
		return 1.0
	}
	if predecessorRank == nil {
		return *successorRank - 1.0
	}
	if successorRank == nil {
		return *predecessorRank + 1.0
	}
	return (*predecessorRank + *successorRank) / 2.0
}

// NeedsRebalance returns true when neighbor rank spacing falls below the floating-point safety threshold.
func NeedsRebalance(afterRank, beforeRank *float64, newRank float64) bool {
	if afterRank != nil && math.Abs(*afterRank-newRank) < rebalanceThreshold {
		return true
	}
	if beforeRank != nil && math.Abs(*beforeRank-newRank) < rebalanceThreshold {
		return true
	}
	return false
}

func (s *service) Rebalance(ctx context.Context, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType) error {
	ids, err := s.repo.ListIDsForRebalance(ctx, scopeType, scopeID, viewType)
	if err != nil {
		return fmt.Errorf("list ranks for rebalance: %w", err)
	}
	if len(ids) == 0 {
		return nil
	}
	if err := s.repo.RebalanceRowsTx(ctx, ids); err != nil {
		return fmt.Errorf("rebalance task ranks: %w", err)
	}
	return nil
}

func (s *service) ensureRanksSeeded(ctx context.Context, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType) error {
	unrankedIDs, err := s.repo.ListUnrankedTopTasks(ctx, scopeType, scopeID, viewType)
	if err != nil {
		return fmt.Errorf("list unranked tasks: %w", err)
	}
	if len(unrankedIDs) == 0 {
		return nil
	}

	baseRank, err := s.repo.GetMaxRankSingleScope(ctx, scopeType, scopeID, viewType)
	if err != nil {
		return fmt.Errorf("get max rank: %w", err)
	}

	rows := make([]sqlc.InsertTaskRankSkipConflictParams, 0, len(unrankedIDs))
	for i, tid := range unrankedIDs {
		rows = append(rows, sqlc.InsertTaskRankSkipConflictParams{
			ID:        uuid.New(),
			TaskID:    tid,
			ScopeType: scopeType,
			ScopeID:   scopeID,
			ViewType:  viewType,
			Rank:      pgtype.Float8{Float64: baseRank + float64(i+1), Valid: true},
		})
	}
	if err := s.repo.InsertRanksSkipConflictTx(ctx, rows); err != nil {
		return fmt.Errorf("seed missing task ranks: %w", err)
	}
	return nil
}

func parseOptionalUUID(raw *string, field string) (*uuid.UUID, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*raw)
	if err != nil {
		return nil, apperror.BadRequest(field + " must be a UUID")
	}
	return &id, nil
}

func (s *service) getNeighborRanks(
	ctx context.Context,
	scopeType sqlc.ScopeType,
	scopeID uuid.UUID,
	viewType sqlc.ViewType,
	afterID, beforeID *uuid.UUID,
) (afterRank *float64, beforeRank *float64, err error) {
	if afterID != nil {
		afterRank, err = s.repo.GetTaskRank(ctx, *afterID, scopeType, scopeID, viewType)
		if err != nil {
			return nil, nil, fmt.Errorf("lookup after task rank: %w", err)
		}
	}
	if beforeID != nil {
		beforeRank, err = s.repo.GetTaskRank(ctx, *beforeID, scopeType, scopeID, viewType)
		if err != nil {
			return nil, nil, fmt.Errorf("lookup before task rank: %w", err)
		}
	}

	if beforeRank != nil && afterRank == nil {
		trueAfter, tErr := s.repo.FindTrueAfterRank(ctx, scopeType, scopeID, viewType, *beforeRank)
		if tErr != nil {
			return nil, nil, fmt.Errorf("find boundary after rank: %w", tErr)
		}
		if trueAfter != nil {
			afterRank = trueAfter
		}
	} else if afterRank != nil && beforeRank == nil {
		trueBefore, tErr := s.repo.FindTrueBeforeRank(ctx, scopeType, scopeID, viewType, *afterRank)
		if tErr != nil {
			return nil, nil, fmt.Errorf("find boundary before rank: %w", tErr)
		}
		if trueBefore != nil {
			beforeRank = trueBefore
		}
	}

	return afterRank, beforeRank, nil
}

func (s *service) Reorder(ctx context.Context, taskID uuid.UUID, req ReorderRequest) error {
	scopeID, err := uuid.Parse(req.ScopeID)
	if err != nil {
		return apperror.BadRequest("scopeId must be a UUID")
	}
	afterID, err := parseOptionalUUID(req.AfterTaskID, "afterTaskId")
	if err != nil {
		return err
	}
	beforeID, err := parseOptionalUUID(req.BeforeTaskID, "beforeTaskId")
	if err != nil {
		return err
	}

	scopeType := sqlc.ScopeType(req.ScopeType)
	viewType := sqlc.ViewType(req.ViewType)

	if seedErr := s.ensureRanksSeeded(ctx, scopeType, scopeID, viewType); seedErr != nil {
		return seedErr
	}

	afterRank, beforeRank, err := s.getNeighborRanks(ctx, scopeType, scopeID, viewType, afterID, beforeID)
	if err != nil {
		return err
	}

	newRank := ComputeRank(afterRank, beforeRank)
	if err := s.repo.UpsertTaskRank(ctx, taskID, scopeType, scopeID, viewType, newRank); err != nil {
		return fmt.Errorf("upsert task rank: %w", err)
	}

	if NeedsRebalance(afterRank, beforeRank, newRank) {
		if err := s.Rebalance(ctx, scopeType, scopeID, viewType); err != nil {
			return err
		}
	}
	return nil
}

// SeedForTasksBatchWithQueries seeds rank rows for one or more newly created tasks inside an active DB transaction.
func SeedForTasksBatchWithQueries(
	ctx context.Context,
	q *sqlc.Queries,
	taskIDs []uuid.UUID,
	projectID, workspaceID, orgID uuid.UUID,
) error {
	if len(taskIDs) == 0 {
		return nil
	}

	existingMax, err := q.GetMaxTaskRanksForScopes(ctx, sqlc.GetMaxTaskRanksForScopesParams{
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
		OrgID:       orgID,
	})
	if err != nil {
		return fmt.Errorf("get max task ranks: %w", err)
	}

	maxMap := make(map[string]float64, len(existingMax))
	for _, r := range existingMax {
		key := string(r.ScopeType) + ":" + r.ScopeID.String() + ":" + string(r.ViewType)
		maxMap[key] = r.MaxRank
	}

	type scopeRef struct {
		scopeType sqlc.ScopeType
		scopeID   uuid.UUID
	}
	scopes := []scopeRef{
		{scopeType: sqlc.ScopeTypePROJECT, scopeID: projectID},
		{scopeType: sqlc.ScopeTypeWORKSPACE, scopeID: workspaceID},
		{scopeType: sqlc.ScopeTypeORGANIZATION, scopeID: orgID},
	}
	views := []sqlc.ViewType{sqlc.ViewTypeLIST, sqlc.ViewTypeGANTT, sqlc.ViewTypeBOARD}

	for _, tid := range taskIDs {
		for _, sc := range scopes {
			for _, vw := range views {
				key := string(sc.scopeType) + ":" + sc.scopeID.String() + ":" + string(vw)
				nextRank := maxMap[key] + 1
				if err := q.InsertTaskRankSkipConflict(ctx, sqlc.InsertTaskRankSkipConflictParams{
					ID:        uuid.New(),
					TaskID:    tid,
					ScopeType: sc.scopeType,
					ScopeID:   sc.scopeID,
					ViewType:  vw,
					Rank:      pgtype.Float8{Float64: nextRank, Valid: true},
				}); err != nil {
					return fmt.Errorf("seed task rank: %w", err)
				}
				maxMap[key] = nextRank
			}
		}
	}
	return nil
}
