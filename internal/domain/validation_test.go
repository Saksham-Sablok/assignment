package domain_test

import (
	"testing"

	"github.com/services-api/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestIsValidSortField(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected bool
	}{
		{
			name:     "valid field - name",
			field:    "name",
			expected: true,
		},
		{
			name:     "valid field - created_at",
			field:    "created_at",
			expected: true,
		},
		{
			name:     "valid field - updated_at",
			field:    "updated_at",
			expected: true,
		},
		{
			name:     "invalid field - id",
			field:    "id",
			expected: false,
		},
		{
			name:     "invalid field - description",
			field:    "description",
			expected: false,
		},
		{
			name:     "invalid field - empty",
			field:    "",
			expected: false,
		},
		{
			name:     "invalid field - random",
			field:    "random_field",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain.IsValidSortField(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidSortFields(t *testing.T) {
	fields := domain.ValidSortFields()
	assert.Contains(t, fields, "name")
	assert.Contains(t, fields, "created_at")
	assert.Contains(t, fields, "updated_at")
	assert.Len(t, fields, 3)
}

func TestPaginationParams_Offset(t *testing.T) {
	tests := []struct {
		name     string
		params   domain.PaginationParams
		expected int
	}{
		{
			name: "page 1 limit 20",
			params: domain.PaginationParams{
				Page:  1,
				Limit: 20,
			},
			expected: 0,
		},
		{
			name: "page 2 limit 20",
			params: domain.PaginationParams{
				Page:  2,
				Limit: 20,
			},
			expected: 20,
		},
		{
			name: "page 3 limit 10",
			params: domain.PaginationParams{
				Page:  3,
				Limit: 10,
			},
			expected: 20,
		},
		{
			name: "page 5 limit 15",
			params: domain.PaginationParams{
				Page:  5,
				Limit: 15,
			},
			expected: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.params.Offset()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewPaginatedResult(t *testing.T) {
	tests := []struct {
		name          string
		dataLen       int
		total         int64
		params        domain.PaginationParams
		expectedPages int
	}{
		{
			name:    "exact pages",
			dataLen: 20,
			total:   100,
			params: domain.PaginationParams{
				Page:  1,
				Limit: 20,
			},
			expectedPages: 5,
		},
		{
			name:    "partial last page",
			dataLen: 10,
			total:   55,
			params: domain.PaginationParams{
				Page:  3,
				Limit: 20,
			},
			expectedPages: 3,
		},
		{
			name:    "single page",
			dataLen: 5,
			total:   5,
			params: domain.PaginationParams{
				Page:  1,
				Limit: 20,
			},
			expectedPages: 1,
		},
		{
			name:    "empty result",
			dataLen: 0,
			total:   0,
			params: domain.PaginationParams{
				Page:  1,
				Limit: 20,
			},
			expectedPages: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]string, tt.dataLen)
			result := domain.NewPaginatedResult(data, tt.total, tt.params)

			assert.Len(t, result.Data, tt.dataLen)
			assert.Equal(t, tt.total, result.Pagination.Total)
			assert.Equal(t, tt.params.Page, result.Pagination.Page)
			assert.Equal(t, tt.params.Limit, result.Pagination.Limit)
			assert.Equal(t, tt.expectedPages, result.Pagination.TotalPages)
		})
	}
}

func TestDefaultPaginationParams(t *testing.T) {
	params := domain.DefaultPaginationParams()
	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 20, params.Limit)
}

func TestDefaultListParams(t *testing.T) {
	params := domain.DefaultListParams()
	assert.Equal(t, "created_at", params.Sort)
	assert.Equal(t, "desc", params.Order)
	assert.Equal(t, 1, params.Pagination.Page)
	assert.Equal(t, 20, params.Pagination.Limit)
}

func TestValidationError(t *testing.T) {
	err := domain.ValidationError{
		Field:   "name",
		Message: "name is required",
	}

	assert.Equal(t, "name is required", err.Error())
}
