package gantt

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// Service defines business operations for Gantt visualization and resource allocation.
type Service interface {
	GetProjectGanttData(ctx context.Context, projectID uuid.UUID) (*ChartData, error)
	GetSprintGanttData(ctx context.Context, sprintID uuid.UUID) (*ChartData, error)
	GetResourceAllocation(ctx context.Context, projectID uuid.UUID) ([]ResourceAllocation, error)
}

type service struct {
	repo Repository
}

// NewService constructs a gantt service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

var statusProgressMap = map[string]int{
	"To Do":       0,
	"TODO":        0,
	"Backlog":     0,
	"In Progress": 50,
	"IN_PROGRESS": 50,
	"Review":      80,
	"REVIEW":      80,
	"Testing":     90,
	"TESTING":     90,
	"Done":        100,
	"DONE":        100,
	"Completed":   100,
	"COMPLETED":   100,
}

func pgTSTime(ts pgtype.Timestamp) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time.UTC()
	return &t
}

func pgTextPtr(t pgtype.Text) *string {
	if !t.Valid || t.String == "" {
		return nil
	}
	s := t.String
	return &s
}

func computeTimeline(rawDates []time.Time, boundStart, boundEnd *time.Time) Timeline {
	all := make([]time.Time, 0, len(rawDates)+2)
	all = append(all, rawDates...)
	if boundStart != nil {
		all = append(all, *boundStart)
	}
	if boundEnd != nil {
		all = append(all, *boundEnd)
	}
	if len(all) == 0 {
		now := time.Now().UTC()
		end := now.Add(30 * 24 * time.Hour)
		return Timeline{
			Start:    formatISO(now),
			End:      formatISO(end),
			Duration: 30,
		}
	}
	minT, maxT := all[0], all[0]
	for _, d := range all[1:] {
		if d.Before(minT) {
			minT = d
		}
		if d.After(maxT) {
			maxT = d
		}
	}
	days := int(math.Ceil(maxT.Sub(minT).Hours() / 24.0))
	return Timeline{
		Start:    formatISO(minT),
		End:      formatISO(maxT),
		Duration: days,
	}
}

func computeCriticalPath(tasks []*TaskItem, deps []sqlc.ListTaskDependenciesForGanttRow) []string {
	taskMap := make(map[string]*TaskItem, len(tasks))
	for _, t := range tasks {
		taskMap[t.ID] = t
	}
	adj := make(map[string][]string)
	for _, d := range deps {
		adj[d.BlockingTaskID.String()] = append(adj[d.BlockingTaskID.String()], d.DependentTaskID.String())
	}

	var findLongest func(id string, visited map[string]bool) []string
	findLongest = func(id string, visited map[string]bool) []string {
		if visited[id] {
			return nil
		}
		visited[id] = true
		defer delete(visited, id)

		bestSub := []string{}
		for _, nxt := range adj[id] {
			if _, ok := taskMap[nxt]; !ok {
				continue
			}
			sub := findLongest(nxt, visited)
			if len(sub) > len(bestSub) {
				bestSub = sub
			}
		}
		return append([]string{id}, bestSub...)
	}

	longest := []string{}
	for _, t := range tasks {
		if len(t.Dependencies) == 0 {
			p := findLongest(t.ID, make(map[string]bool))
			if len(p) > len(longest) {
				longest = p
			}
		}
	}
	return longest
}

func (s *service) buildTaskItems(
	rows []sqlc.ListTasksForProjectGanttRow,
	assignees []sqlc.ListTaskAssigneesForGanttRow,
	depMap map[string][]string,
	fallbackStart, fallbackEnd *time.Time,
) ([]*TaskItem, []time.Time) {
	assMap := make(map[string][]AssigneeBrief)
	for _, a := range assignees {
		assMap[a.TaskID.String()] = append(assMap[a.TaskID.String()], AssigneeBrief{
			ID:        a.UserID.String(),
			FirstName: a.FirstName,
			LastName:  a.LastName,
			Avatar:    pgTextPtr(a.Avatar),
		})
	}

	items := make([]*TaskItem, 0, len(rows))
	var dates []time.Time
	for _, r := range rows {
		st := pgTSTime(r.StartDate)
		if st == nil {
			st = fallbackStart
		}
		en := pgTSTime(r.DueDate)
		if en == nil {
			en = fallbackEnd
		}
		if st != nil && en != nil {
			dates = append(dates, *st, *en)
		}

		idStr := r.ID.String()
		deps := depMap[idStr]
		if deps == nil {
			deps = []string{}
		}
		var parentPtr *string
		if r.ParentTaskID.Valid {
			p := uuid.UUID(r.ParentTaskID.Bytes).String()
			parentPtr = &p
		}

		items = append(items, &TaskItem{
			ID:           idStr,
			Title:        r.Title,
			Start:        formatISOPtr(st),
			End:          formatISOPtr(en),
			Progress:     statusProgressMap[r.StatusName],
			Dependencies: deps,
			Assignees:    assMap[idStr],
			Priority:     string(r.Priority),
			Status: StatusBrief{
				Name:  r.StatusName,
				Color: r.StatusColor,
			},
			Type:   string(r.Type),
			Key:    r.Slug,
			Parent: parentPtr,
		})
	}
	return items, dates
}

func (s *service) GetProjectGanttData(ctx context.Context, projectID uuid.UUID) (*ChartData, error) {
	project, err := s.repo.GetProjectBasic(ctx, projectID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch project")
	}
	if project == nil {
		return nil, apperror.NotFound("Project not found")
	}

	taskRows, err := s.repo.ListProjectTasks(ctx, projectID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch project tasks")
	}
	sprintRows, err := s.repo.ListSprintsByProject(ctx, projectID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch project sprints")
	}
	depRows, err := s.repo.ListDependenciesByProject(ctx, projectID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch task dependencies")
	}

	taskIDs := make([]uuid.UUID, len(taskRows))
	for i, r := range taskRows {
		taskIDs[i] = r.ID
	}
	assignees, err := s.repo.ListAssigneesByTaskIDs(ctx, taskIDs)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch task assignees")
	}

	depMap := make(map[string][]string)
	for _, d := range depRows {
		k := d.DependentTaskID.String()
		depMap[k] = append(depMap[k], d.BlockingTaskID.String())
	}

	flatTasks, rawDates := s.buildTaskItems(taskRows, assignees, depMap, nil, nil)

	taskByID := make(map[string]*TaskItem, len(flatTasks))
	for _, t := range flatTasks {
		taskByID[t.ID] = t
	}
	roots := make([]*TaskItem, 0, len(flatTasks))
	for _, t := range flatTasks {
		if t.Parent != nil {
			if p, ok := taskByID[*t.Parent]; ok {
				p.Children = append(p.Children, t)
				continue
			}
		} else {
			roots = append(roots, t)
		}
	}

	timeline := computeTimeline(rawDates, pgTSTime(project.StartDate), pgTSTime(project.EndDate))
	criticalPath := computeCriticalPath(flatTasks, depRows)

	type datedMilestone struct {
		t time.Time
		m Milestone
	}
	var ms []datedMilestone
	for _, sp := range sprintRows {
		if sp.StartDate.Valid {
			t := sp.StartDate.Time.UTC()
			ms = append(ms, datedMilestone{
				t: t,
				m: Milestone{ID: sp.ID.String(), Title: sp.Name + " Start", Date: formatISOPtr(&t), Type: "sprint_start"},
			})
		}
		if sp.EndDate.Valid {
			t := sp.EndDate.Time.UTC()
			ms = append(ms, datedMilestone{
				t: t,
				m: Milestone{ID: sp.ID.String() + "_end", Title: sp.Name + " End", Date: formatISOPtr(&t), Type: "sprint_end"},
			})
		}
	}
	sort.Slice(ms, func(i, j int) bool { return ms[i].t.Before(ms[j].t) })
	milestones := make([]Milestone, 0, len(ms))
	for _, item := range ms {
		milestones = append(milestones, item.m)
	}

	return &ChartData{
		Tasks:        roots,
		Timeline:     timeline,
		CriticalPath: criticalPath,
		Milestones:   milestones,
	}, nil
}

func (s *service) GetSprintGanttData(ctx context.Context, sprintID uuid.UUID) (*ChartData, error) {
	sprint, err := s.repo.GetSprintBasic(ctx, sprintID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch sprint")
	}
	if sprint == nil {
		return nil, apperror.NotFound("Sprint not found")
	}

	rawTasks, err := s.repo.ListSprintTasks(ctx, sprintID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch sprint tasks")
	}
	taskRows := make([]sqlc.ListTasksForProjectGanttRow, len(rawTasks))
	taskIDs := make([]uuid.UUID, len(rawTasks))
	for i, r := range rawTasks {
		taskRows[i] = sqlc.ListTasksForProjectGanttRow(r)
		taskIDs[i] = r.ID
	}

	assignees, err := s.repo.ListAssigneesByTaskIDs(ctx, taskIDs)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch task assignees")
	}
	depRows, err := s.repo.ListDependenciesByProject(ctx, sprint.ProjectID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch task dependencies")
	}
	depMap := make(map[string][]string)
	for _, d := range depRows {
		k := d.DependentTaskID.String()
		depMap[k] = append(depMap[k], d.BlockingTaskID.String())
	}

	spStart := pgTSTime(sprint.StartDate)
	spEnd := pgTSTime(sprint.EndDate)
	ganttTasks, rawDates := s.buildTaskItems(taskRows, assignees, depMap, spStart, spEnd)

	timeline := computeTimeline(rawDates, spStart, spEnd)
	criticalPath := computeCriticalPath(ganttTasks, nil)

	milestones := []Milestone{
		{ID: sprint.ID.String(), Title: sprint.Name + " Start", Date: formatISOPtr(spStart), Type: "sprint_start"},
		{ID: sprint.ID.String() + "_end", Title: sprint.Name + " End", Date: formatISOPtr(spEnd), Type: "sprint_end"},
	}

	return &ChartData{
		Tasks:        ganttTasks,
		Timeline:     timeline,
		CriticalPath: criticalPath,
		Milestones:   milestones,
	}, nil
}

func (s *service) GetResourceAllocation(ctx context.Context, projectID uuid.UUID) ([]ResourceAllocation, error) {
	tasks, err := s.repo.ListResourceTasks(ctx, projectID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch resource allocation tasks")
	}
	if len(tasks) == 0 {
		return []ResourceAllocation{}, nil
	}

	taskIDs := make([]uuid.UUID, len(tasks))
	for i, t := range tasks {
		taskIDs[i] = t.ID
	}
	assignees, err := s.repo.ListAssigneesByTaskIDs(ctx, taskIDs)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch task assignees")
	}

	byTask := make(map[uuid.UUID][]sqlc.ListTaskAssigneesForGanttRow)
	for _, a := range assignees {
		byTask[a.TaskID] = append(byTask[a.TaskID], a)
	}

	resMap := make(map[string]*ResourceAllocation)
	order := make([]string, 0)

	for _, t := range tasks {
		taskAssignees := byTask[t.ID]
		if len(taskAssignees) == 0 {
			continue
		}
		sp := 1
		if t.StoryPoints.Valid && t.StoryPoints.Int32 != 0 {
			sp = int(t.StoryPoints.Int32)
		}
		share := float64(sp) / float64(len(taskAssignees))
		item := ResourceTask{
			ID:          t.ID.String(),
			Title:       t.Title,
			Start:       formatISOPtr(pgTSTime(t.StartDate)),
			End:         formatISOPtr(pgTSTime(t.DueDate)),
			StoryPoints: sp,
		}

		for _, a := range taskAssignees {
			uid := a.UserID.String()
			alloc, ok := resMap[uid]
			if !ok {
				order = append(order, uid)
				alloc = &ResourceAllocation{
					Assignee: ResourceAssignee{
						ID:     uid,
						Name:   a.FirstName + " " + a.LastName,
						Avatar: pgTextPtr(a.Avatar),
					},
					Tasks: []ResourceTask{},
				}
				resMap[uid] = alloc
			}
			alloc.Tasks = append(alloc.Tasks, item)
			alloc.Workload += share
		}
	}

	out := make([]ResourceAllocation, 0, len(order))
	for _, uid := range order {
		out = append(out, *resMap[uid])
	}
	return out, nil
}

