package task_ranks

// ReorderRequest represents the payload for PATCH /api/task-ranks/:taskId/reorder.
type ReorderRequest struct {
	ScopeType    string  `json:"scopeType" validate:"required,oneof=ORGANIZATION WORKSPACE PROJECT"`
	ScopeID      string  `json:"scopeId" validate:"required,uuid"`
	ViewType     string  `json:"viewType" validate:"required,oneof=LIST BOARD GANTT"`
	AfterTaskID  *string `json:"afterTaskId" validate:"omitempty"`
	BeforeTaskID *string `json:"beforeTaskId" validate:"omitempty"`
}

// RebalanceMessageResponse represents GET /api/task-ranks/rebalance.
type RebalanceMessageResponse struct {
	Message string `json:"message"`
}
