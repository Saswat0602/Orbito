package workflows

import (
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// DefaultStatusDefinition defines the template for initial task statuses in a workflow.
type DefaultStatusDefinition struct {
	Name      string
	Color     string
	Category  sqlc.StatusCategory
	Position  int32
	IsDefault bool
}

// DefaultTaskStatuses defines standard 4-state workflow statuses matching NestJS constants.
var DefaultTaskStatuses = []DefaultStatusDefinition{
	{
		Name:      "To Do",
		Color:     "#6366f1",
		Category:  sqlc.StatusCategoryTODO,
		Position:  1,
		IsDefault: true,
	},
	{
		Name:      "In Progress",
		Color:     "#f59e0b",
		Category:  sqlc.StatusCategoryINPROGRESS,
		Position:  2,
		IsDefault: false,
	},
	{
		Name:      "In Review",
		Color:     "#8b5cf6",
		Category:  sqlc.StatusCategoryINPROGRESS,
		Position:  3,
		IsDefault: false,
	},
	{
		Name:      "Done",
		Color:     "#10b981",
		Category:  sqlc.StatusCategoryDONE,
		Position:  4,
		IsDefault: false,
	},
}
