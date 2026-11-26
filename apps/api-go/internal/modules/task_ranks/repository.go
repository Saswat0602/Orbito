package task_ranks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines persistence operations for task ranking and Lexorank reordering.
type Repository interface {
	ListIDsForRebalance(ctx context.Context, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType) ([]uuid.UUID, error)
	RebalanceRowsTx(ctx context.Context, ids []uuid.UUID) error
	ListUnrankedTopTasks(ctx context.Context, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType) ([]uuid.UUID, error)
	GetMaxRankSingleScope(ctx context.Context, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType) (float64, error)
	InsertRanksSkipConflictTx(ctx context.Context, params []sqlc.InsertTaskRankSkipConflictParams) error
	GetTaskRank(ctx context.Context, taskID uuid.UUID, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType) (*float64, error)
	FindTrueAfterRank(ctx context.Context, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType, beforeRank float64) (*float64, error)
	FindTrueBeforeRank(ctx context.Context, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType, afterRank float64) (*float64, error)
	UpsertTaskRank(ctx context.Context, taskID uuid.UUID, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType, rank float64) error
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL task_ranks repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{db: db, queries: db.Queries()}
}

func (r *pgRepository) ListIDsForRebalance(ctx context.Context, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType) ([]uuid.UUID, error) {
	return r.queries.GetTaskRanksForRebalance(ctx, sqlc.GetTaskRanksForRebalanceParams{
		ScopeType: scopeType,
		ScopeID:   scopeID,
		ViewType:  viewType,
	})
}

func (r *pgRepository) RebalanceRowsTx(ctx context.Context, ids []uuid.UUID) error {
	return r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		for idx, id := range ids {
			if err := q.UpdateTaskRankValueByID(ctx, sqlc.UpdateTaskRankValueByIDParams{
				ID:   id,
				Rank: pgtype.Float8{Float64: float64(idx + 1), Valid: true},
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *pgRepository) ListUnrankedTopTasks(ctx context.Context, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType) ([]uuid.UUID, error) {
	switch scopeType {
	case sqlc.ScopeTypePROJECT:
		return r.queries.ListUnrankedTopTasksInProject(ctx, sqlc.ListUnrankedTopTasksInProjectParams{
			ProjectID: scopeID,
			ScopeType: scopeType,
			ScopeID:   scopeID,
			ViewType:  viewType,
		})
	case sqlc.ScopeTypeWORKSPACE:
		return r.queries.ListUnrankedTopTasksInWorkspace(ctx, sqlc.ListUnrankedTopTasksInWorkspaceParams{
			WorkspaceID: scopeID,
			ScopeType:   scopeType,
			ScopeID:     scopeID,
			ViewType:    viewType,
		})
	default:
		return r.queries.ListUnrankedTopTasksInOrg(ctx, sqlc.ListUnrankedTopTasksInOrgParams{
			OrganizationID: scopeID,
			ScopeType:      scopeType,
			ScopeID:        scopeID,
			ViewType:       viewType,
		})
	}
}

func (r *pgRepository) GetMaxRankSingleScope(ctx context.Context, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType) (float64, error) {
	return r.queries.GetMaxTaskRankSingleScope(ctx, sqlc.GetMaxTaskRankSingleScopeParams{
		ScopeType: scopeType,
		ScopeID:   scopeID,
		ViewType:  viewType,
	})
}

func (r *pgRepository) InsertRanksSkipConflictTx(ctx context.Context, params []sqlc.InsertTaskRankSkipConflictParams) error {
	return r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		for _, item := range params {
			if err := q.InsertTaskRankSkipConflict(ctx, item); err != nil {
				return err
			}
		}
		return nil
	})
}

func float8Ptr(v pgtype.Float8) *float64 {
	if !v.Valid {
		return nil
	}
	f := v.Float64
	return &f
}

func (r *pgRepository) GetTaskRank(ctx context.Context, taskID uuid.UUID, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType) (*float64, error) {
	v, err := r.queries.GetTaskRankByUniqueKey(ctx, sqlc.GetTaskRankByUniqueKeyParams{
		TaskID:    taskID,
		ScopeType: scopeType,
		ScopeID:   scopeID,
		ViewType:  viewType,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return float8Ptr(v), nil
}

func (r *pgRepository) FindTrueAfterRank(ctx context.Context, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType, beforeRank float64) (*float64, error) {
	v, err := r.queries.FindTrueAfterRank(ctx, sqlc.FindTrueAfterRankParams{
		ScopeType: scopeType,
		ScopeID:   scopeID,
		ViewType:  viewType,
		Rank:      pgtype.Float8{Float64: beforeRank, Valid: true},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return float8Ptr(v), nil
}

func (r *pgRepository) FindTrueBeforeRank(ctx context.Context, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType, afterRank float64) (*float64, error) {
	v, err := r.queries.FindTrueBeforeRank(ctx, sqlc.FindTrueBeforeRankParams{
		ScopeType: scopeType,
		ScopeID:   scopeID,
		ViewType:  viewType,
		Rank:      pgtype.Float8{Float64: afterRank, Valid: true},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return float8Ptr(v), nil
}

func (r *pgRepository) UpsertTaskRank(ctx context.Context, taskID uuid.UUID, scopeType sqlc.ScopeType, scopeID uuid.UUID, viewType sqlc.ViewType, rank float64) error {
	return r.queries.UpsertTaskRank(ctx, sqlc.UpsertTaskRankParams{
		ID:        uuid.New(),
		TaskID:    taskID,
		ScopeType: scopeType,
		ScopeID:   scopeID,
		ViewType:  viewType,
		Rank:      pgtype.Float8{Float64: rank, Valid: true},
	})
}
