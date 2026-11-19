package settings

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines data access methods for system and user settings.
type Repository interface {
	GetByUserAndKey(ctx context.Context, userID uuid.UUID, key string) (*sqlc.Setting, error)
	GetGlobalByKey(ctx context.Context, key string) (*sqlc.Setting, error)
	UpsertUserSetting(ctx context.Context, params sqlc.UpsertUserSettingParams) (*sqlc.Setting, error)
	CreateGlobalSetting(ctx context.Context, params sqlc.CreateGlobalSettingParams) (*sqlc.Setting, error)
	UpdateGlobalSetting(ctx context.Context, params sqlc.UpdateGlobalSettingByIDParams) (*sqlc.Setting, error)
	ListUserSettings(ctx context.Context, userID uuid.UUID, category string) ([]sqlc.Setting, error)
	ListGlobalSettings(ctx context.Context, category string) ([]sqlc.Setting, error)
	DeleteUserSetting(ctx context.Context, userID uuid.UUID, key string) error
	DeleteGlobalSetting(ctx context.Context, key string) error
	CountUserSetting(ctx context.Context, userID uuid.UUID, key string) (int64, error)
	CountGlobalSetting(ctx context.Context, key string) (int64, error)
	WithTx(ctx context.Context, fn func(repo Repository) error) error
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository creates a new PostgreSQL settings repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{
		db:      db,
		queries: db.Queries(),
	}
}

func (r *pgRepository) GetByUserAndKey(ctx context.Context, userID uuid.UUID, key string) (*sqlc.Setting, error) {
	row, err := r.queries.GetSettingByKeyAndUser(ctx, sqlc.GetSettingByKeyAndUserParams{
		Key:    key,
		UserId: pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetGlobalByKey(ctx context.Context, key string) (*sqlc.Setting, error) {
	row, err := r.queries.GetGlobalSettingByKey(ctx, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) UpsertUserSetting(ctx context.Context, params sqlc.UpsertUserSettingParams) (*sqlc.Setting, error) {
	row, err := r.queries.UpsertUserSetting(ctx, params)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) CreateGlobalSetting(ctx context.Context, params sqlc.CreateGlobalSettingParams) (*sqlc.Setting, error) {
	row, err := r.queries.CreateGlobalSetting(ctx, params)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) UpdateGlobalSetting(ctx context.Context, params sqlc.UpdateGlobalSettingByIDParams) (*sqlc.Setting, error) {
	row, err := r.queries.UpdateGlobalSettingByID(ctx, params)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListUserSettings(ctx context.Context, userID uuid.UUID, category string) ([]sqlc.Setting, error) {
	return r.queries.ListUserSettings(ctx, sqlc.ListUserSettingsParams{
		UserId:  pgtype.UUID{Bytes: userID, Valid: true},
		Column2: category,
	})
}

func (r *pgRepository) ListGlobalSettings(ctx context.Context, category string) ([]sqlc.Setting, error) {
	return r.queries.ListGlobalSettings(ctx, category)
}

func (r *pgRepository) DeleteUserSetting(ctx context.Context, userID uuid.UUID, key string) error {
	return r.queries.DeleteUserSetting(ctx, sqlc.DeleteUserSettingParams{
		UserId: pgtype.UUID{Bytes: userID, Valid: true},
		Key:    key,
	})
}

func (r *pgRepository) DeleteGlobalSetting(ctx context.Context, key string) error {
	return r.queries.DeleteGlobalSetting(ctx, key)
}

func (r *pgRepository) CountUserSetting(ctx context.Context, userID uuid.UUID, key string) (int64, error) {
	return r.queries.CountUserSetting(ctx, sqlc.CountUserSettingParams{
		UserId: pgtype.UUID{Bytes: userID, Valid: true},
		Key:    key,
	})
}

func (r *pgRepository) CountGlobalSetting(ctx context.Context, key string) (int64, error) {
	return r.queries.CountGlobalSetting(ctx, key)
}

func (r *pgRepository) WithTx(ctx context.Context, fn func(repo Repository) error) error {
	return r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		txRepo := &pgRepository{
			db:      r.db,
			queries: q,
		}
		return fn(txRepo)
	})
}
