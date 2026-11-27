package editor_images

import (
	"context"

	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// Repository persists MediaAsset records for uploaded editor images.
type Repository interface {
	CreateMediaAsset(ctx context.Context, arg sqlc.CreateMediaAssetRecordParams) (sqlc.MediaAsset, error)
}

type pgRepository struct {
	queries *sqlc.Queries
}

// NewRepository creates a PostgreSQL-backed editor images repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{
		queries: sqlc.New(db.Pool()),
	}
}

func (r *pgRepository) CreateMediaAsset(ctx context.Context, arg sqlc.CreateMediaAssetRecordParams) (sqlc.MediaAsset, error) {
	row, err := r.queries.CreateMediaAssetRecord(ctx, arg)
	if err != nil {
		return sqlc.MediaAsset{}, apperror.Map(err)
	}
	return row, nil
}
