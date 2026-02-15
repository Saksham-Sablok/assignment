package domain

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// DefaultPaginationParams returns default pagination parameters
func DefaultPaginationParams() PaginationParams {
	return PaginationParams{
		Page:  1,
		Limit: 20,
	}
}

// Offset calculates the offset for database queries
func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}

// PaginatedResult holds paginated query results
type PaginatedResult[T any] struct {
	Data       []T                `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

// PaginationMetadata holds pagination metadata for responses
type PaginationMetadata struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

// NewPaginatedResult creates a new paginated result
func NewPaginatedResult[T any](data []T, total int64, params PaginationParams) *PaginatedResult[T] {
	totalPages := int(total) / params.Limit
	if int(total)%params.Limit > 0 {
		totalPages++
	}

	return &PaginatedResult[T]{
		Data: data,
		Pagination: PaginationMetadata{
			Total:      total,
			Page:       params.Page,
			Limit:      params.Limit,
			TotalPages: totalPages,
		},
	}
}

// ListParams holds parameters for listing services
type ListParams struct {
	Search     string           `json:"search,omitempty"`
	Name       string           `json:"name,omitempty"`
	Sort       string           `json:"sort,omitempty"`
	Order      string           `json:"order,omitempty"`
	Pagination PaginationParams `json:"pagination"`
}

// DefaultListParams returns default list parameters
func DefaultListParams() ListParams {
	return ListParams{
		Sort:       "created_at",
		Order:      "desc",
		Pagination: DefaultPaginationParams(),
	}
}

// ValidSortFields returns the valid sort fields
func ValidSortFields() []string {
	return []string{"name", "created_at", "updated_at"}
}

// IsValidSortField checks if the sort field is valid
func IsValidSortField(field string) bool {
	for _, f := range ValidSortFields() {
		if f == field {
			return true
		}
	}
	return false
}
