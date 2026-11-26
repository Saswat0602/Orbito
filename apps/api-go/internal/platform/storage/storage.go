package storage

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/orbito/api-go/internal/shared/apperror"
)

// SaveResult represents the stored file metadata returned after upload.
type SaveResult struct {
	URL  *string
	Key  string
	Size int64
}

// InlineSafeTypes maps safe file extensions to Content-Type values that cannot run scripts on our origin.
var InlineSafeTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".bmp":  "image/bmp",
	".ico":  "image/x-icon",
	".pdf":  "application/pdf",
	".txt":  "text/plain",
	".csv":  "text/csv",
	".mp4":  "video/mp4",
	".webm": "video/webm",
	".ogv":  "video/ogg",
	".mov":  "video/quicktime",
	".mp3":  "audio/mpeg",
	".oga":  "audio/ogg",
	".wav":  "audio/wav",
}

// Provider defines file persistence operations used across uploads, attachments, and avatars.
type Provider interface {
	SaveFile(folder, originalName string, content []byte) (SaveResult, error)
	DeleteFile(key string) error
	StreamLocal(w http.ResponseWriter, r *http.Request, key string, forceAttachment bool) error
	UploadDir() string
}

// LocalProvider stores and serves files under the configured local upload directory.
type LocalProvider struct {
	uploadDir string
}

// NewLocalProvider initializes a local disk storage provider ensuring uploadDir exists.
func NewLocalProvider(uploadDir string) (*LocalProvider, error) {
	if strings.TrimSpace(uploadDir) == "" {
		uploadDir = "./uploads"
	}
	absDir, err := filepath.Abs(uploadDir)
	if err != nil {
		return nil, fmt.Errorf("resolve upload dir: %w", err)
	}
	if err := os.MkdirAll(absDir, 0o750); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &LocalProvider{uploadDir: absDir}, nil
}

// UploadDir returns the resolved absolute root directory for uploads.
func (p *LocalProvider) UploadDir() string {
	return p.uploadDir
}

func sanitizePathComponent(component string) (string, error) {
	if component == "" {
		return "", apperror.BadRequest("Invalid path component")
	}
	cleaned := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(component, "..", ""), "/", ""), "\\", "")
	if cleaned != component {
		return "", apperror.BadRequest("Invalid characters in path")
	}
	return cleaned, nil
}

func sanitizeFolderPath(folder string) (string, error) {
	if folder == "" {
		return "", apperror.BadRequest("Invalid folder path")
	}
	if strings.Contains(folder, "..") {
		return "", apperror.BadRequest("Invalid characters in folder path")
	}
	rawParts := strings.Split(folder, "/")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	if len(parts) == 0 {
		return "", apperror.BadRequest("Empty folder path")
	}
	return strings.Join(parts, "/"), nil
}

func (p *LocalProvider) resolveWithinBase(relPath string) (string, error) {
	trimmed := strings.TrimPrefix(relPath, "/")
	target := filepath.Clean(filepath.Join(p.uploadDir, filepath.FromSlash(trimmed)))
	basePrefix := p.uploadDir + string(filepath.Separator)
	if target != p.uploadDir && !strings.HasPrefix(target, basePrefix) {
		return "", apperror.NotFound("Invalid file path")
	}
	return target, nil
}

// SaveFile writes bytes safely inside uploadDir/folder/originalName.
func (p *LocalProvider) SaveFile(folder, originalName string, content []byte) (SaveResult, error) {
	safeFolder, err := sanitizeFolderPath(folder)
	if err != nil {
		return SaveResult{}, err
	}
	safeName, err := sanitizePathComponent(originalName)
	if err != nil {
		return SaveResult{}, err
	}

	key := safeFolder + "/" + safeName
	targetDir, err := p.resolveWithinBase(safeFolder)
	if err != nil {
		return SaveResult{}, err
	}
	if mkErr := os.MkdirAll(targetDir, 0o750); mkErr != nil {
		return SaveResult{}, fmt.Errorf("create upload subfolder: %w", mkErr)
	}
	filePath, err := p.resolveWithinBase(key)
	if err != nil {
		return SaveResult{}, err
	}
	if wrErr := os.WriteFile(filePath, content, 0o600); wrErr != nil {
		return SaveResult{}, fmt.Errorf("write upload file: %w", wrErr)
	}

	url := "/" + key
	return SaveResult{
		URL:  &url,
		Key:  key,
		Size: int64(len(content)),
	}, nil
}

// DeleteFile removes a locally stored file if it exists.
func (p *LocalProvider) DeleteFile(key string) error {
	filePath, err := p.resolveWithinBase(key)
	if err != nil {
		return err
	}
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

// StreamLocal streams a local file with XSS-safe Content-Type and Content-Disposition headers.
func (p *LocalProvider) StreamLocal(w http.ResponseWriter, _ *http.Request, key string, forceAttachment bool) error {
	filePath, err := p.resolveWithinBase(key)
	if err != nil {
		return err
	}
	// #nosec G304 -- path verified against basePrefix via resolveWithinBase
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return apperror.NotFound("File not found on server")
		}
		return fmt.Errorf("open stored file: %w", err)
	}
	defer func() { _ = f.Close() }()

	fileName := filepath.Base(filePath)
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

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("stream file: %w", err)
	}
	return nil
}
