package task_statuses

import "github.com/orbito/api-go/internal/platform/database/sqlc"

// DefaultFromProjectColor is the default color applied when creating a status from project configuration.
const DefaultFromProjectColor = "#000000"

// DefaultFromProjectCategory is the default category applied when creating a status from project configuration.
const DefaultFromProjectCategory = sqlc.StatusCategoryTODO
