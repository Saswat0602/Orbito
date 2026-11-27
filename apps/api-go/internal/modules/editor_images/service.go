package editor_images

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/platform/storage"
	"github.com/orbito/api-go/internal/shared/apperror"
)

const defaultMaxEditorImageBytes int64 = 5 * 1024 * 1024

// StorageBackend abstracts file storage and S3 detection.
type StorageBackend interface {
	SaveFile(folder, originalName string, content []byte) (storage.SaveResult, error)
	IsUsingS3() bool
}

// Service handles validating, persisting, and recording editor image uploads.
type Service interface {
	UploadImage(ctx context.Context, in UploadInput, userID uuid.UUID) (*UploadResponse, error)
	MaxSize() int64
}

type service struct {
	repo         Repository
	store        StorageBackend
	folder       string
	maxSize      int64
	allowedMimes map[string]bool
	allowedShort []string
}

// NewService constructs an editor images service configured from Config.
func NewService(repo Repository, store StorageBackend, cfg *config.Config) Service {
	folder := "editor-images"
	maxSize := defaultMaxEditorImageBytes
	mimeList := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	if cfg != nil {
		if strings.TrimSpace(cfg.EditorImageUploadFolder) != "" {
			folder = strings.TrimSpace(cfg.EditorImageUploadFolder)
		}
		if cfg.MaxEditorImageSize > 0 {
			maxSize = cfg.MaxEditorImageSize
		}
		if len(cfg.AllowedImageTypes) > 0 {
			mimeList = cfg.AllowedImageTypes
		}
	}

	allowed := make(map[string]bool, len(mimeList))
	short := make([]string, 0, len(mimeList))
	for _, m := range mimeList {
		trimmed := strings.TrimSpace(strings.ToLower(m))
		if trimmed == "" {
			continue
		}
		allowed[trimmed] = true
		if parts := strings.SplitN(trimmed, "/", 2); len(parts) == 2 {
			short = append(short, parts[1])
		} else {
			short = append(short, trimmed)
		}
	}

	return &service{
		repo:         repo,
		store:        store,
		folder:       folder,
		maxSize:      maxSize,
		allowedMimes: allowed,
		allowedShort: short,
	}
}

func (s *service) MaxSize() int64 {
	return s.maxSize
}

func (s *service) UploadImage(ctx context.Context, in UploadInput, userID uuid.UUID) (*UploadResponse, error) {
	mimeLower := strings.ToLower(strings.TrimSpace(in.MimeType))
	if !s.allowedMimes[mimeLower] {
		return nil, apperror.BadRequest("Invalid file type. Allowed types: " + strings.Join(s.allowedShort, ", "))
	}
	fileLen := int64(len(in.Content))
	if fileLen > s.maxSize {
		maxMB := float64(s.maxSize) / (1024 * 1024)
		return nil, apperror.BadRequest(fmt.Sprintf("File size exceeds %.0fMB limit", maxMB))
	}

	uniqueName := generateUniqueFilename(in.OriginalName, mimeLower)
	saved, err := s.store.SaveFile(s.folder, uniqueName, in.Content)
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256(in.Content)
	hexDigest := hex.EncodeToString(sum[:])
	inCloud := s.store.IsUsingS3()

	asset, err := s.repo.CreateMediaAsset(ctx, sqlc.CreateMediaAssetRecordParams{
		ID:         uuid.New(),
		UserId:     userID,
		AssetType:  "editor-image",
		StorageKey: saved.Key,
		Filename:   in.OriginalName,
		MimeType:   mimeLower,
		Size:       clampInt32(fileLen),
		FileHash:   pgtype.Text{String: hexDigest, Valid: true},
		InCloud:    inCloud,
	})
	if err != nil {
		return nil, err
	}

	assetID := asset.ID
	key := saved.Key
	return &UploadResponse{
		Message: "Image uploaded successfully",
		ID:      &assetID,
		URL:     saved.URL,
		Key:     &key,
		Size:    saved.Size,
		InCloud: inCloud,
	}, nil
}

func generateUniqueFilename(originalName, mimeLower string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(originalName)), ".")
	if ext == "" {
		ext = strings.TrimPrefix(mimeLower, "image/")
		if ext == "" {
			ext = "png"
		}
	}
	return fmt.Sprintf("%s-%d.%s", uuid.NewString(), time.Now().UnixMilli(), ext)
}

func clampInt32(v int64) int32 {
	if v < 0 {
		return 0
	}
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(v)
}
