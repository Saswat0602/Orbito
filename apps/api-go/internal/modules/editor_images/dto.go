package editor_images

import "github.com/google/uuid"

// UploadInput holds the validated multipart upload bytes and file metadata.
type UploadInput struct {
	OriginalName string
	MimeType     string
	Content      []byte
}

// UploadResponse mirrors NestJS POST /api/editor-images/upload response JSON.
type UploadResponse struct {
	Message string     `json:"message"`
	ID      *uuid.UUID `json:"id"`
	URL     *string    `json:"url"`
	Key     *string    `json:"key"`
	Size    int64      `json:"size"`
	InCloud bool       `json:"inCloud"`
}
