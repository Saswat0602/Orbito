package pagination

import (
	"math"
)

// Pagination defines the standard 5-field pagination metadata matching NestJS.
type Pagination struct {
	CurrentPage int   `json:"currentPage"`
	TotalPages  int   `json:"totalPages"`
	TotalCount  int64 `json:"totalCount"`
	HasNextPage bool  `json:"hasNextPage"`
	HasPrevPage bool  `json:"hasPrevPage"`
}

// Params holds the parsed query parameters for pagination.
type Params struct {
	Page  int
	Limit int
}

// New constructs a standard Pagination envelope given currentPage, limit, and totalCount.
func New(page, limit int, totalCount int64) Pagination {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	totalPages := 0
	if totalCount > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(limit)))
	}

	hasNextPage := page < totalPages
	hasPrevPage := page > 1 && (totalPages == 0 || page <= totalPages+1)

	return Pagination{
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalCount:  totalCount,
		HasNextPage: hasNextPage,
		HasPrevPage: hasPrevPage,
	}
}

// Offset calculates the SQL OFFSET value from page and limit (0-indexed).
func Offset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	return (page - 1) * limit
}

// NormalizeLimit ensures the limit is within valid boundaries (1 to maxLimit).
func NormalizeLimit(limit, defaultLimit, maxLimit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}
