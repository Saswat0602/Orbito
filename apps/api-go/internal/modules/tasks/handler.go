package tasks

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/modules/task_ranks"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

var allowedAttachmentMimes = map[string]bool{
	"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true, "image/svg+xml": true,
	"application/pdf": true, "application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
	"application/vnd.ms-powerpoint":                                             true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"text/plain": true, "text/csv": true, "text/markdown": true, "application/json": true,
	"application/xml": true, "text/html": true, "text/css": true, "text/javascript": true,
	"application/zip": true, "application/x-rar-compressed": true, "application/x-7z-compressed": true,
	"video/mp4": true, "video/webm": true, "video/ogg": true, "video/mpeg": true,
	"video/quicktime": true, "video/x-msvideo": true, "video/x-matroska": true,
}

// Handler processes HTTP requests for tasks and task sharing.
type Handler struct {
	service Service
}

// NewHandler constructs a tasks HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

func extractAuth(r *http.Request) (uuid.UUID, bool, error) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		return uuid.Nil, false, apperror.Unauthorized("Unauthenticated")
	}
	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)
	return userID, isSuperAdmin, nil
}

// Create handles POST /api/tasks.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req CreateTaskRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Create(r.Context(), req, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

func parseMultipartCreateRequest(r *http.Request) (CreateTaskRequest, []UploadedFile, error) {
	if err := r.ParseMultipartForm(100 * 1024 * 1024); err != nil {
		return CreateTaskRequest{}, nil, apperror.BadRequest("Invalid multipart payload")
	}
	req := CreateTaskRequest{
		Title:     r.FormValue("title"),
		ProjectID: r.FormValue("projectId"),
	}
	if v := r.FormValue("description"); v != "" {
		req.Description = &v
	}
	if v := r.FormValue("type"); v != "" {
		req.Type = &v
	}
	if v := r.FormValue("priority"); v != "" {
		req.Priority = &v
	}
	if v := r.FormValue("startDate"); v != "" {
		req.StartDate = &v
	}
	if v := r.FormValue("dueDate"); v != "" {
		req.DueDate = &v
	}
	if v := r.FormValue("statusId"); v != "" {
		req.StatusID = &v
	}
	if v := r.FormValue("sprintId"); v != "" {
		req.SprintID = &v
	}
	if v := r.FormValue("parentTaskId"); v != "" {
		req.ParentTaskID = &v
	}
	if v := r.FormValue("assigneeIds"); v != "" {
		var ids []string
		if json.Unmarshal([]byte(v), &ids) == nil {
			req.AssigneeIDs = ids
		} else {
			req.AssigneeIDs = strings.Split(v, ",")
		}
	}
	if v := r.FormValue("reporterIds"); v != "" {
		var ids []string
		if json.Unmarshal([]byte(v), &ids) == nil {
			req.ReporterIDs = ids
		} else {
			req.ReporterIDs = strings.Split(v, ",")
		}
	}
	if err := httpx.Validator().Struct(req); err != nil {
		return CreateTaskRequest{}, nil, apperror.BadRequest(err.Error())
	}

	var headers = r.MultipartForm.File["attachments"]
	if len(headers) > 10 {
		return CreateTaskRequest{}, nil, apperror.BadRequest("Maximum 10 files allowed")
	}
	files := make([]UploadedFile, 0, len(headers))
	for _, fh := range headers {
		if fh.Size > 10*1024*1024 {
			return CreateTaskRequest{}, nil, apperror.BadRequest("File exceeds 10MB limit")
		}
		mimeType := fh.Header.Get("Content-Type")
		if !allowedAttachmentMimes[mimeType] {
			return CreateTaskRequest{}, nil, apperror.BadRequest("File type " + mimeType + " is not allowed")
		}
		f, oErr := fh.Open()
		if oErr != nil {
			return CreateTaskRequest{}, nil, apperror.BadRequest("Failed to read uploaded file")
		}
		data, rErr := io.ReadAll(f)
		_ = f.Close()
		if rErr != nil {
			return CreateTaskRequest{}, nil, apperror.BadRequest("Failed to read uploaded file")
		}
		files = append(files, UploadedFile{FileName: fh.Filename, MimeType: mimeType, Content: data})
	}
	return req, files, nil
}

// CreateWithAttachments handles POST /api/tasks/create-task-attachment.
func (h *Handler) CreateWithAttachments(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	req, files, err := parseMultipartCreateRequest(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.CreateWithAttachments(r.Context(), req, files, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

func parseTaskFilterQuery(r *http.Request) TaskFilterQuery {
	q := r.URL.Query()
	split := func(k string) []string {
		v := strings.TrimSpace(q.Get(k))
		return parseCSV(&v)
	}
	var parentPtr *string
	if q.Has("parentTaskId") {
		v := q.Get("parentTaskId")
		parentPtr = &v
	}
	var fromT, toT *time.Time
	if f := q.Get("from"); f != "" {
		if parsed, ok := parseISO8601(f); ok {
			fromT = &parsed
		}
	}
	if t := q.Get("to"); t != "" {
		if parsed, ok := parseISO8601(t); ok {
			toT = &parsed
		}
	}
	return TaskFilterQuery{
		OrganizationID: q.Get("organizationId"),
		ProjectIDs:     split("projectId"),
		WorkspaceIDs:   split("workspaceId"),
		SprintID:       q.Get("sprintId"),
		ParentTaskID:   parentPtr,
		Priorities:     split("priorities"),
		Statuses:       split("statuses"),
		Types:          split("types"),
		AssigneeIDs:    split("assigneeIds"),
		ReporterIDs:    split("reporterIds"),
		Search:         q.Get("search"),
		SortBy:         q.Get("sortBy"),
		SortOrder:      q.Get("sortOrder"),
		ViewType:       httpx.QueryParam(r, "viewType", "LIST"),
		Page:           httpx.QueryParamInt(r, "page", 1),
		Limit:          httpx.QueryParamInt(r, "limit", 20),
		GroupBy:        q.Get("groupBy"),
		From:           fromT,
		To:             toT,
		DateField:      httpx.QueryParam(r, "dateField", "dueDate"),
	}
}

// FindAll handles GET /api/tasks.
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.FindAll(r.Context(), parseTaskFilterQuery(r), userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetAllTasks handles GET /api/tasks/all-tasks.
func (h *Handler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetTasks(r.Context(), parseTaskFilterQuery(r), userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetGroupedTasks handles GET /api/tasks/grouped.
func (h *Handler) GetGroupedTasks(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	q := r.URL.Query()
	groupBy := q.Get("groupBy")
	if groupBy == "" {
		httpx.RespondError(w, r, apperror.BadRequest("groupBy should not be empty"))
		return
	}
	split := func(k string) []string {
		v := strings.TrimSpace(q.Get(k))
		return parseCSV(&v)
	}
	query := GroupedTasksQuery{
		OrganizationID: q.Get("organizationId"),
		WorkspaceIDs:   split("workspaceId"),
		ProjectIDs:     split("projectId"),
		SprintID:       q.Get("sprintId"),
		Priorities:     split("priorities"),
		Statuses:       split("statuses"),
		Types:          split("types"),
		AssigneeIDs:    split("assigneeIds"),
		ReporterIDs:    split("reporterIds"),
		Search:         q.Get("search"),
		GroupBy:        groupBy,
		GroupKey:       q.Get("groupKey"),
		Page:           httpx.QueryParamInt(r, "page", 1),
		LimitPerGroup:  httpx.QueryParamInt(r, "limitPerGroup", 20),
	}
	res, err := h.service.GetTasksGrouped(r.Context(), query, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetTasksByStatus handles GET /api/tasks/by-status.
func (h *Handler) GetTasksByStatus(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	q := r.URL.Query()
	slug := q.Get("slug")
	if slug == "" {
		httpx.RespondError(w, r, apperror.BadRequest("slug should not be empty"))
		return
	}
	includeSub, _ := strconv.ParseBool(q.Get("includeSubtasks"))
	res, err := h.service.GetTasksGroupedByStatus(r.Context(), ByStatusQuery{
		Slug:            slug,
		IncludeSubtasks: includeSub,
		StatusID:        q.Get("statusId"),
		SprintID:        q.Get("sprintId"),
		Page:            httpx.QueryParamInt(r, "page", 1),
		Limit:           httpx.QueryParamInt(r, "limit", 25),
	}, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetTodaysTasks handles GET /api/tasks/today.
func (h *Handler) GetTodaysTasks(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	page := max(1, httpx.QueryParamInt(r, "page", 1))
	limit := min(100, max(1, httpx.QueryParamInt(r, "limit", 10)))
	res, err := h.service.FindTodaysTasks(r.Context(), r.URL.Query().Get("organizationId"), page, limit, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetTasksByOrganization handles GET /api/tasks/organization/{orgId}.
func (h *Handler) GetTasksByOrganization(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	orgID, err := httpx.URLParamUUID(r, "orgId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	page := max(1, httpx.QueryParamInt(r, "page", 1))
	limit := min(100, max(1, httpx.QueryParamInt(r, "limit", 10)))
	res, err := h.service.FindByOrganization(r.Context(), orgID, r.URL.Query().Get("priority"), r.URL.Query().Get("search"), page, limit, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindOne handles GET /api/tasks/{id} & GET /api/tasks/slug/{slug}.
func (h *Handler) FindOne(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	idOrSlug := chi.URLParam(r, "id")
	if idOrSlug == "" {
		idOrSlug = chi.URLParam(r, "slug")
	}
	res, err := h.service.FindOne(r.Context(), idOrSlug, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindByKey handles GET /api/tasks/key/{key}.
func (h *Handler) FindByKey(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.FindByKey(r.Context(), chi.URLParam(r, "key"), userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Update handles PATCH /api/tasks/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	rawBytes, rErr := io.ReadAll(r.Body)
	if rErr != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Failed to read body"))
		return
	}
	var req UpdateTaskRequest
	if bErr := decodeStrictJSON(rawBytes, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	var rawMap map[string]json.RawMessage
	if json.Unmarshal(rawBytes, &rawMap) == nil {
		if v, ok := rawMap["sprintId"]; ok && string(v) == "null" {
			req.ClearSprintID = true
		}
		if v, ok := rawMap["parentTaskId"]; ok && string(v) == "null" {
			req.ClearParentTaskID = true
		}
		if v, ok := rawMap["completedAt"]; ok && string(v) == "null" {
			req.ClearCompletedAt = true
		}
	}

	res, err := h.service.Update(r.Context(), chi.URLParam(r, "id"), req, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

func decodeStrictJSON(raw []byte, target any) error {
	if len(raw) == 0 {
		return apperror.BadRequest("Request body must not be empty")
	}
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return apperror.BadRequest(err.Error())
	}
	if err := httpx.Validator().Struct(target); err != nil {
		return apperror.BadRequest(err.Error())
	}
	return nil
}

// Reorder handles PATCH /api/tasks/{id}/reorder.
func (h *Handler) Reorder(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req task_ranks.ReorderRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	if err := h.service.ReorderTask(r.Context(), chi.URLParam(r, "id"), req, userID, isSuperAdmin); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UpdateStatus handles PATCH /api/tasks/{id}/status.
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req UpdateStatusRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Update(r.Context(), chi.URLParam(r, "id"), UpdateTaskRequest{StatusID: &req.StatusID}, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// UpdateAssignees handles PATCH /api/tasks/{id}/assignees.
func (h *Handler) UpdateAssignees(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req UpdateAssigneesRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Update(r.Context(), chi.URLParam(r, "id"), UpdateTaskRequest{AssigneeIDs: &req.AssigneeIDs}, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// UnassignTask handles PATCH /api/tasks/{id}/unassign.
func (h *Handler) UnassignTask(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	empty := []string{}
	res, err := h.service.Update(r.Context(), chi.URLParam(r, "id"), UpdateTaskRequest{AssigneeIDs: &empty}, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// UpdatePriority handles PATCH /api/tasks/{id}/priority.
func (h *Handler) UpdatePriority(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req UpdatePriorityRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Update(r.Context(), chi.URLParam(r, "id"), UpdateTaskRequest{Priority: &req.Priority}, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// UpdateDueDate handles PATCH /api/tasks/{id}/due-date.
func (h *Handler) UpdateDueDate(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req UpdateDueDateRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Update(r.Context(), chi.URLParam(r, "id"), UpdateTaskRequest{DueDate: &req.DueDate}, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// AddComment handles POST /api/tasks/{id}/comments.
func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req AddCommentRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.AddComment(r.Context(), chi.URLParam(r, "id"), req.Comment, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// Remove handles DELETE /api/tasks/{id}.
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.Remove(r.Context(), chi.URLParam(r, "id"), userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// BulkCreate handles POST /api/tasks/bulk-create.
func (h *Handler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req BulkCreateTasksRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.BulkCreate(r.Context(), req, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// BulkDelete handles POST /api/tasks/bulk-delete.
func (h *Handler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req BulkDeleteTasksRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.BulkDeleteTasks(r.Context(), req, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// BulkStatusUpdate handles POST /api/tasks/bulk-status-update.
func (h *Handler) BulkStatusUpdate(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req BulkUpdateTasksStatusRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.BulkUpdateTasksStatus(r.Context(), req, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// BulkAssign handles POST /api/tasks/bulk-assign.
func (h *Handler) BulkAssign(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req BulkAssignTasksRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.BulkAssignTasks(r.Context(), req, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// CompleteOccurrence handles POST /api/tasks/{id}/complete-occurrence.
func (h *Handler) CompleteOccurrence(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.CompleteOccurrenceAndGenerateNext(r.Context(), chi.URLParam(r, "id"), userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// AddRecurrence handles POST /api/tasks/{id}/recurrence.
func (h *Handler) AddRecurrence(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req RecurrenceConfigRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.AddRecurrence(r.Context(), chi.URLParam(r, "id"), req, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// UpdateRecurrence handles PATCH /api/tasks/{id}/recurrence.
func (h *Handler) UpdateRecurrence(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req RecurrenceConfigRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.UpdateRecurrenceConfig(r.Context(), chi.URLParam(r, "id"), req, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// StopRecurrence handles DELETE /api/tasks/{id}/recurrence.
func (h *Handler) StopRecurrence(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.StopRecurrence(r.Context(), chi.URLParam(r, "id"), userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetRecurringTasks handles GET /api/tasks/recurring/project/{projectId}.
func (h *Handler) GetRecurringTasks(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	projectID, err := httpx.URLParamUUID(r, "projectId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetRecurringTasks(r.Context(), projectID, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// CreateShare handles POST /api/task-shares.
func (h *Handler) CreateShare(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req CreateShareLinkRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.CreateShareLink(r.Context(), req, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// GetSharesForTask handles GET /api/task-shares/task/{taskId}.
func (h *Handler) GetSharesForTask(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetSharesForTask(r.Context(), chi.URLParam(r, "taskId"), userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// RevokeShare handles DELETE /api/task-shares/{shareId}.
func (h *Handler) RevokeShare(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	shareID, err := httpx.URLParamUUID(r, "shareId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if err := h.service.RevokeShare(r.Context(), shareID, userID, isSuperAdmin); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]string{"message": "Share link revoked successfully"})
}
