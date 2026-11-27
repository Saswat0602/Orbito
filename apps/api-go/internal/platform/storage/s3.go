package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/shared/apperror"
)

const presignExpiry = 30 * time.Minute

// Manager coordinates AWS S3 and local disk storage with automatic fallback to local disk.
type Manager struct {
	local         *LocalProvider
	s3Client      *s3.Client
	presignClient *s3.PresignClient
	bucketName    string
	region        string
	usingS3       bool
	log           *slog.Logger
}

// NewManager initializes local storage and optionally activates AWS S3 if credentials and bucket are reachable.
func NewManager(ctx context.Context, cfg *config.Config, log *slog.Logger) (*Manager, error) {
	if log == nil {
		log = slog.Default()
	}
	uploadDir := "./uploads"
	bucket := "orbito-assets"
	region := "us-east-1"
	if cfg != nil {
		if cfg.UploadDest != "" {
			uploadDir = cfg.UploadDest
		}
		if cfg.AWSBucketName != "" {
			bucket = cfg.AWSBucketName
		}
		if cfg.AWSRegion != "" {
			region = cfg.AWSRegion
		}
	}

	local, err := NewLocalProvider(uploadDir)
	if err != nil {
		return nil, err
	}

	mgr := &Manager{
		local:      local,
		bucketName: bucket,
		region:     region,
		log:        log,
	}

	if cfg != nil && cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "" {
		mgr.initS3(ctx, cfg)
	}
	return mgr, nil
}

func (m *Manager) initS3(ctx context.Context, cfg *config.Config) {
	loadCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	awsCfg, err := awsconfig.LoadDefaultConfig(
		loadCtx,
		awsconfig.WithRegion(m.region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, "")),
	)
	if err != nil {
		m.log.Warn("Failed loading AWS S3 config; using local disk storage", slog.String("error", err.Error()))
		return
	}

	client := s3.NewFromConfig(awsCfg)
	if _, headErr := client.HeadBucket(loadCtx, &s3.HeadBucketInput{Bucket: &m.bucketName}); headErr != nil {
		m.log.Warn("S3 bucket unreachable; using local disk storage fallback", slog.String("bucket", m.bucketName), slog.String("error", headErr.Error()))
		return
	}

	m.s3Client = client
	m.presignClient = s3.NewPresignClient(client)
	m.usingS3 = true
	m.log.Info("AWS S3 storage initialized", slog.String("bucket", m.bucketName), slog.String("region", m.region))
}

// IsUsingS3 reports whether active storage is backed by AWS S3.
func (m *Manager) IsUsingS3() bool {
	return m.usingS3
}

// UploadDir returns the local upload directory root.
func (m *Manager) UploadDir() string {
	return m.local.UploadDir()
}

// SaveFile persists bytes in S3 when active or on local disk otherwise.
func (m *Manager) SaveFile(folder, originalName string, content []byte) (SaveResult, error) {
	if !m.usingS3 || m.s3Client == nil {
		return m.local.SaveFile(folder, originalName, content)
	}

	safeFolder, err := sanitizeFolderPath(folder)
	if err != nil {
		return SaveResult{}, err
	}
	safeName, err := sanitizePathComponent(originalName)
	if err != nil {
		return SaveResult{}, err
	}
	key := safeFolder + "/" + safeName
	contentType := detectContentType(safeName, content)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = m.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &m.bucketName,
		Key:         &key,
		Body:        bytes.NewReader(content),
		ContentType: &contentType,
	})
	if err != nil {
		return SaveResult{}, fmt.Errorf("upload object to s3: %w", err)
	}

	objectURL := fmt.Sprintf("/api/uploads/%s", key)
	if signed, sErr := m.PresignGetURL(ctx, key); sErr == nil {
		objectURL = signed
	}
	return SaveResult{
		URL:  &objectURL,
		Key:  key,
		Size: int64(len(content)),
	}, nil
}

// DeleteFile deletes an object from S3 or local disk.
func (m *Manager) DeleteFile(key string) error {
	if !m.usingS3 || m.s3Client == nil {
		return m.local.DeleteFile(key)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := m.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &m.bucketName,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("delete s3 object: %w", err)
	}
	return nil
}

// StreamLocal delegates directly to local disk streaming.
func (m *Manager) StreamLocal(w http.ResponseWriter, r *http.Request, key string, forceAttachment bool) error {
	return m.local.StreamLocal(w, r, key, forceAttachment)
}

// Stream serves a stored object from S3 (when active) or local disk with XSS-safe headers.
func (m *Manager) Stream(w http.ResponseWriter, r *http.Request, key string, forceAttachment bool) error {
	if !m.usingS3 || m.s3Client == nil {
		return m.local.StreamLocal(w, r, key, forceAttachment)
	}

	out, err := m.s3Client.GetObject(r.Context(), &s3.GetObjectInput{
		Bucket: &m.bucketName,
		Key:    &key,
	})
	if err != nil {
		return apperror.NotFound("File not found in storage")
	}
	defer func() { _ = out.Body.Close() }()

	fileName := filepath.Base(key)
	safeHeaderName := strings.NewReplacer(`"`, "", `\`, "", "\r", "", "\n", "").Replace(fileName)
	ext := strings.ToLower(filepath.Ext(fileName))
	contentType := "application/octet-stream"
	disposition := "attachment"
	if safeMime, ok := InlineSafeTypes[ext]; ok && !forceAttachment {
		contentType = safeMime
		disposition = "inline"
	} else if forceAttachment && InlineSafeTypes[ext] != "" {
		contentType = InlineSafeTypes[ext]
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`%s; filename="%s"`, disposition, safeHeaderName))
	if _, copyErr := io.Copy(w, out.Body); copyErr != nil {
		return fmt.Errorf("stream s3 object: %w", copyErr)
	}
	return nil
}

// PresignPutURL generates a 30-minute presigned S3 PUT URL for the provided object key.
func (m *Manager) PresignPutURL(ctx context.Context, key string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", apperror.BadRequest("key query parameter is required")
	}
	if m.presignClient == nil {
		return "", apperror.BadRequest("S3 storage is not configured")
	}
	res, err := m.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: &m.bucketName,
		Key:    &key,
	}, s3.WithPresignExpires(presignExpiry))
	if err != nil {
		return "", fmt.Errorf("presign put url: %w", err)
	}
	return res.URL, nil
}

// PresignGetURL generates a 30-minute presigned S3 GET URL for the provided object key.
func (m *Manager) PresignGetURL(ctx context.Context, key string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", apperror.BadRequest("key query parameter is required")
	}
	if m.presignClient == nil {
		return "", apperror.BadRequest("S3 storage is not configured")
	}
	res, err := m.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &m.bucketName,
		Key:    &key,
	}, s3.WithPresignExpires(presignExpiry))
	if err != nil {
		return "", fmt.Errorf("presign get url: %w", err)
	}
	return res.URL, nil
}

func detectContentType(fileName string, content []byte) string {
	if mType, ok := InlineSafeTypes[strings.ToLower(filepath.Ext(fileName))]; ok {
		return mType
	}
	if len(content) > 0 {
		return http.DetectContentType(content)
	}
	return "application/octet-stream"
}
