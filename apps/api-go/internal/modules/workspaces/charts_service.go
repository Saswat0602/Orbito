package workspaces

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// ChartsService calculates aggregated analytics and metrics for workspaces.
type ChartsService interface {
	GetKPIMetrics(ctx context.Context, workspaceID uuid.UUID) (*WorkspaceKPIMetrics, error)
	GetChartData(ctx context.Context, workspaceID uuid.UUID) (*WorkspaceChartDataResponse, error)
}

type chartsService struct {
	repo Repository
}

// NewChartsService creates a workspace analytics service.
func NewChartsService(repo Repository) ChartsService {
	return &chartsService{repo: repo}
}

func (c *chartsService) GetKPIMetrics(ctx context.Context, workspaceID uuid.UUID) (*WorkspaceKPIMetrics, error) {
	totalProjects, err := c.repo.CountProjects(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	activeProjects, err := c.repo.CountProjectsByStatus(ctx, workspaceID, sqlc.ProjectStatusACTIVE)
	if err != nil {
		return nil, err
	}

	completedProjects, err := c.repo.CountProjectsByStatus(ctx, workspaceID, sqlc.ProjectStatusCOMPLETED)
	if err != nil {
		return nil, err
	}

	totalTasks, err := c.repo.CountTasks(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	completedTasks, err := c.repo.CountCompletedTasks(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	overdueTasks, err := c.repo.CountOverdueTasks(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var completionRate float64
	if totalTasks > 0 {
		rate := (float64(completedTasks) / float64(totalTasks)) * 100
		completionRate = math.Round(rate*100) / 100
	}

	return &WorkspaceKPIMetrics{
		TotalProjects:     totalProjects,
		ActiveProjects:    activeProjects,
		CompletedProjects: completedProjects,
		TotalTasks:        totalTasks,
		CompletedTasks:    completedTasks,
		OverdueTasks:      overdueTasks,
		CompletionRate:    completionRate,
	}, nil
}

func (c *chartsService) GetChartData(ctx context.Context, workspaceID uuid.UUID) (*WorkspaceChartDataResponse, error) {
	kpi, err := c.GetKPIMetrics(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	projStatuses, err := c.repo.ProjectStatusStats(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	projectStatusData := make([]map[string]any, 0, len(projStatuses))
	for _, ps := range projStatuses {
		projectStatusData = append(projectStatusData, map[string]any{
			"status": string(ps.Status),
			"_count": map[string]any{"status": ps.Count},
		})
	}

	taskPriorities, err := c.repo.TaskPriorityStats(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	taskPriorityData := make([]map[string]any, 0, len(taskPriorities))
	for _, tp := range taskPriorities {
		taskPriorityData = append(taskPriorityData, map[string]any{
			"priority": string(tp.Priority),
			"_count":   map[string]any{"priority": tp.Count},
		})
	}

	return &WorkspaceChartDataResponse{
		KPIMetrics:    kpi,
		ProjectStatus: projectStatusData,
		TaskPriority:  taskPriorityData,
	}, nil
}
