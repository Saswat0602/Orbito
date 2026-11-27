package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/queue"
)

// BulkImportJobData matches NestJS BulkTaskImportJobData.
type BulkImportJobData struct {
	Dto          BulkCreateTasksRequest `json:"dto"`
	UserID       string                 `json:"userId"`
	IsSuperAdmin bool                   `json:"isSuperAdmin,omitempty"`
}

// BulkImportProcessor handles queued bulk task creation jobs.
type BulkImportProcessor struct {
	tasksSvc Service
	log      *slog.Logger
}

// NewBulkImportProcessor constructs a bulk import worker processor.
func NewBulkImportProcessor(tasksSvc Service, log *slog.Logger) *BulkImportProcessor {
	if log == nil {
		log = slog.Default()
	}
	return &BulkImportProcessor{tasksSvc: tasksSvc, log: log}
}

// RegisterBulkImportProcessor binds the bulk-task-import worker onto the queue runtime.
func RegisterBulkImportProcessor(rt *queue.Runtime, tasksSvc Service, log *slog.Logger) {
	proc := NewBulkImportProcessor(tasksSvc, log)
	rt.RegisterQueue(queue.QueueBulkTaskImport)
	rt.RegisterHandler(queue.TaskBulkImportTasks, proc.Process)
}

// Process executes a bulk task import job payload.
func (p *BulkImportProcessor) Process(ctx context.Context, payload []byte) error {
	var data BulkImportJobData
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("decode bulk import job: %w", err)
	}
	uid, err := uuid.Parse(data.UserID)
	if err != nil {
		return fmt.Errorf("invalid bulk import userId: %w", err)
	}
	if _, err := p.tasksSvc.BulkCreate(ctx, data.Dto, uid, data.IsSuperAdmin); err != nil {
		p.log.Error("Bulk task import job failed", slog.String("userId", data.UserID), slog.String("error", err.Error()))
		return err
	}
	return nil
}
