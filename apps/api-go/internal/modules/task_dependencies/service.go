package task_dependencies

import (
	"context"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// Service defines business operations for task dependencies.
type Service interface {
	Create(ctx context.Context, req CreateRequest, callerID uuid.UUID) (*DependencyResponse, error)
	CreateBulk(ctx context.Context, req BulkCreateRequest, callerID uuid.UUID) ([]DependencyResponse, error)
	FindAll(ctx context.Context, projectID *uuid.UUID) ([]DependencyResponse, error)
	FindOne(ctx context.Context, id uuid.UUID) (*DependencyResponse, error)
	GetTaskDependencies(ctx context.Context, taskIdentifier string) (*TaskDependenciesSummaryResponse, error)
	GetBlockedTasks(ctx context.Context, taskIdentifier string) ([]DependencyResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*DependencyResponse, error)
	Remove(ctx context.Context, id uuid.UUID) error
	RemoveByTasks(ctx context.Context, dependentIdentifier, blockingIdentifier string) error
	GetDependencyStats(ctx context.Context, projectID uuid.UUID) (*DependencyStatsResponse, error)
}

type service struct {
	repo Repository
}

// NewService constructs a task_dependencies service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) resolveTaskID(ctx context.Context, identifier string) (uuid.UUID, error) {
	id, found, err := s.repo.FindTaskID(ctx, identifier)
	if err != nil {
		return uuid.Nil, apperror.Internal("Failed to lookup task")
	}
	if !found {
		return uuid.Nil, apperror.NotFound("Task not found")
	}
	return id, nil
}

func (s *service) validateNoCycles(ctx context.Context, dependentTaskID, blockingTaskID uuid.UUID) error {
	visited := make(map[uuid.UUID]bool)

	var existsPath func(currentID, targetID uuid.UUID) (bool, error)
	existsPath = func(currentID, targetID uuid.UUID) (bool, error) {
		if currentID == targetID {
			return true, nil
		}
		visited[currentID] = true

		nextIDs, err := s.repo.ListBlockingIDs(ctx, currentID)
		if err != nil {
			return false, err
		}
		for _, nid := range nextIDs {
			if !visited[nid] {
				hasPath, pathErr := existsPath(nid, targetID)
				if pathErr != nil {
					return false, pathErr
				}
				if hasPath {
					return true, nil
				}
			}
		}
		return false, nil
	}

	hasCycle, err := existsPath(blockingTaskID, dependentTaskID)
	if err != nil {
		return apperror.Internal("Failed to validate dependency graph")
	}
	if hasCycle {
		return apperror.BadRequest("Creating this dependency would result in a circular dependency")
	}
	return nil
}

func (s *service) Create(ctx context.Context, req CreateRequest, callerID uuid.UUID) (*DependencyResponse, error) {
	if req.DependentTaskID == req.BlockingTaskID {
		return nil, apperror.BadRequest("A task cannot depend on itself")
	}

	dependentID, err := s.resolveTaskID(ctx, req.DependentTaskID)
	if err != nil {
		return nil, err
	}
	blockingID, err := s.resolveTaskID(ctx, req.BlockingTaskID)
	if err != nil {
		return nil, err
	}
	if dependentID == blockingID {
		return nil, apperror.BadRequest("A task cannot depend on itself")
	}

	exists, err := s.repo.CheckExists(ctx, dependentID, blockingID)
	if err != nil {
		return nil, apperror.Internal("Failed to check existing dependency")
	}
	if exists {
		return nil, apperror.Conflict("Dependency relationship already exists")
	}

	if cycleErr := s.validateNoCycles(ctx, dependentID, blockingID); cycleErr != nil {
		return nil, cycleErr
	}

	depType := sqlc.DependencyTypeBLOCKS
	if req.Type != nil && *req.Type != "" {
		depType = sqlc.DependencyType(*req.Type)
	}

	createdBy := &callerID
	if req.CreatedBy != nil && *req.CreatedBy != "" {
		parsed, pErr := uuid.Parse(*req.CreatedBy)
		if pErr != nil {
			return nil, apperror.BadRequest("createdBy must be a UUID")
		}
		createdBy = &parsed
	}

	newID, err := s.repo.Create(ctx, uuid.New(), depType, dependentID, blockingID, createdBy)
	if err != nil {
		return nil, apperror.Internal("Failed to create task dependency")
	}

	detail, err := s.repo.GetByID(ctx, newID)
	if err != nil || detail == nil {
		return nil, apperror.Internal("Failed to fetch created dependency")
	}
	mapped := mapDetailRow(detail)
	return &mapped, nil
}

func (s *service) CreateBulk(ctx context.Context, req BulkCreateRequest, callerID uuid.UUID) ([]DependencyResponse, error) {
	results := make([]DependencyResponse, 0, len(req.Dependencies))
	for _, item := range req.Dependencies {
		created, err := s.Create(ctx, item, callerID)
		if err != nil {
			continue
		}
		results = append(results, *created)
	}
	return results, nil
}

func (s *service) FindAll(ctx context.Context, projectID *uuid.UUID) ([]DependencyResponse, error) {
	rows, err := s.repo.ListAll(ctx, projectID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch task dependencies")
	}
	out := make([]DependencyResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapListRow(r))
	}
	return out, nil
}

func (s *service) FindOne(ctx context.Context, id uuid.UUID) (*DependencyResponse, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch task dependency")
	}
	if row == nil {
		return nil, apperror.NotFound("Task dependency not found")
	}
	mapped := mapDetailRow(row)
	return &mapped, nil
}

func (s *service) GetTaskDependencies(ctx context.Context, taskIdentifier string) (*TaskDependenciesSummaryResponse, error) {
	taskID, err := s.resolveTaskID(ctx, taskIdentifier)
	if err != nil {
		return nil, err
	}

	depsOnRows, err := s.repo.ListDependsOn(ctx, taskID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch task dependencies")
	}
	blocksRows, err := s.repo.ListBlocks(ctx, taskID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch blocked tasks")
	}

	dependsOn := make([]DependencyResponse, 0, len(depsOnRows))
	for _, r := range depsOnRows {
		dependsOn = append(dependsOn, mapDependsOnRow(r))
	}
	blocks := make([]DependencyResponse, 0, len(blocksRows))
	for _, r := range blocksRows {
		blocks = append(blocks, mapBlocksRow(r))
	}

	return &TaskDependenciesSummaryResponse{
		DependsOn: dependsOn,
		Blocks:    blocks,
	}, nil
}

func (s *service) GetBlockedTasks(ctx context.Context, taskIdentifier string) ([]DependencyResponse, error) {
	taskID, err := s.resolveTaskID(ctx, taskIdentifier)
	if err != nil {
		return nil, err
	}
	blocksRows, err := s.repo.ListBlocks(ctx, taskID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch blocked tasks")
	}
	out := make([]DependencyResponse, 0, len(blocksRows))
	for _, r := range blocksRows {
		out = append(out, mapBlocksRow(r))
	}
	return out, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*DependencyResponse, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch task dependency")
	}
	if existing == nil {
		return nil, apperror.NotFound("Task dependency not found")
	}

	var newDepID, newBlkID *uuid.UUID
	checkCycles := false
	resolvedDep := existing.DependentTaskID
	resolvedBlk := existing.BlockingTaskID

	if req.DependentTaskID != nil && *req.DependentTaskID != "" {
		idVal, resErr := s.resolveTaskID(ctx, *req.DependentTaskID)
		if resErr != nil {
			return nil, resErr
		}
		resolvedDep = idVal
		newDepID = &idVal
		checkCycles = true
	}
	if req.BlockingTaskID != nil && *req.BlockingTaskID != "" {
		idVal, resErr := s.resolveTaskID(ctx, *req.BlockingTaskID)
		if resErr != nil {
			return nil, resErr
		}
		resolvedBlk = idVal
		newBlkID = &idVal
		checkCycles = true
	}

	if checkCycles {
		if resolvedDep == resolvedBlk {
			return nil, apperror.BadRequest("A task cannot depend on itself")
		}
		if cycleErr := s.validateNoCycles(ctx, resolvedDep, resolvedBlk); cycleErr != nil {
			return nil, cycleErr
		}
	}

	var depType *sqlc.DependencyType
	if req.Type != nil && *req.Type != "" {
		dt := sqlc.DependencyType(*req.Type)
		depType = &dt
	}

	if _, updErr := s.repo.Update(ctx, id, depType, newDepID, newBlkID); updErr != nil {
		return nil, apperror.Internal("Failed to update task dependency")
	}

	updated, err := s.repo.GetByID(ctx, id)
	if err != nil || updated == nil {
		return nil, apperror.Internal("Failed to load updated dependency")
	}
	mapped := mapDetailRow(updated)
	return &mapped, nil
}

func (s *service) Remove(ctx context.Context, id uuid.UUID) error {
	affected, err := s.repo.DeleteByID(ctx, id)
	if err != nil {
		return apperror.Internal("Failed to remove task dependency")
	}
	if affected == 0 {
		return apperror.NotFound("Task dependency not found")
	}
	return nil
}

func (s *service) RemoveByTasks(ctx context.Context, dependentIdentifier, blockingIdentifier string) error {
	depID, err := s.resolveTaskID(ctx, dependentIdentifier)
	if err != nil {
		return err
	}
	blkID, err := s.resolveTaskID(ctx, blockingIdentifier)
	if err != nil {
		return err
	}

	affected, err := s.repo.DeleteByPair(ctx, depID, blkID)
	if err != nil {
		return apperror.Internal("Failed to remove task dependency")
	}
	if affected == 0 {
		return apperror.NotFound("Task dependency not found")
	}
	return nil
}

func (s *service) GetDependencyStats(ctx context.Context, projectID uuid.UUID) (*DependencyStatsResponse, error) {
	rows, err := s.repo.ListAll(ctx, &projectID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch dependency stats")
	}

	blockedSet := make(map[uuid.UUID]struct{}, len(rows))
	for _, r := range rows {
		blockedSet[r.DependentTaskID] = struct{}{}
	}

	return &DependencyStatsResponse{
		TotalDependencies: len(rows),
		BlockedTasks:      len(blockedSet),
		CriticalPath:      []string{},
	}, nil
}
