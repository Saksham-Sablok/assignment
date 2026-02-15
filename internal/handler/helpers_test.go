package handler_test

import (
	"net/http/httptest"
	"testing"

	"github.com/services-api/internal/handler"
	"github.com/stretchr/testify/assert"
)

func TestParsePaginationParams(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedPage  int
		expectedLimit int
	}{
		{
			name:          "default values",
			query:         "",
			expectedPage:  1,
			expectedLimit: 20,
		},
		{
			name:          "custom page and limit",
			query:         "?page=3&limit=50",
			expectedPage:  3,
			expectedLimit: 50,
		},
		{
			name:          "page only",
			query:         "?page=5",
			expectedPage:  5,
			expectedLimit: 20,
		},
		{
			name:          "limit only",
			query:         "?limit=30",
			expectedPage:  1,
			expectedLimit: 30,
		},
		{
			name:          "limit exceeds max - capped at 100",
			query:         "?limit=200",
			expectedPage:  1,
			expectedLimit: 100,
		},
		{
			name:          "invalid page - uses default",
			query:         "?page=invalid",
			expectedPage:  1,
			expectedLimit: 20,
		},
		{
			name:          "invalid limit - uses default",
			query:         "?limit=invalid",
			expectedPage:  1,
			expectedLimit: 20,
		},
		{
			name:          "negative page - uses default",
			query:         "?page=-1",
			expectedPage:  1,
			expectedLimit: 20,
		},
		{
			name:          "zero page - uses default",
			query:         "?page=0",
			expectedPage:  1,
			expectedLimit: 20,
		},
		{
			name:          "zero limit - uses default",
			query:         "?limit=0",
			expectedPage:  1,
			expectedLimit: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/services"+tt.query, nil)
			params := handler.ParsePaginationParams(req)

			assert.Equal(t, tt.expectedPage, params.Page)
			assert.Equal(t, tt.expectedLimit, params.Limit)
		})
	}
}

func TestParseListParams(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedSearch string
		expectedName   string
		expectedSort   string
		expectedOrder  string
		expectedPage   int
		expectedLimit  int
	}{
		{
			name:           "default values",
			query:          "",
			expectedSearch: "",
			expectedName:   "",
			expectedSort:   "created_at",
			expectedOrder:  "desc",
			expectedPage:   1,
			expectedLimit:  20,
		},
		{
			name:           "search parameter",
			query:          "?search=payment",
			expectedSearch: "payment",
			expectedName:   "",
			expectedSort:   "created_at",
			expectedOrder:  "desc",
			expectedPage:   1,
			expectedLimit:  20,
		},
		{
			name:           "name filter",
			query:          "?name=payment-service",
			expectedSearch: "",
			expectedName:   "payment-service",
			expectedSort:   "created_at",
			expectedOrder:  "desc",
			expectedPage:   1,
			expectedLimit:  20,
		},
		{
			name:           "sort by name ascending",
			query:          "?sort=name&order=asc",
			expectedSearch: "",
			expectedName:   "",
			expectedSort:   "name",
			expectedOrder:  "asc",
			expectedPage:   1,
			expectedLimit:  20,
		},
		{
			name:           "sort descending with alternate keyword",
			query:          "?sort=updated_at&order=descending",
			expectedSearch: "",
			expectedName:   "",
			expectedSort:   "updated_at",
			expectedOrder:  "desc",
			expectedPage:   1,
			expectedLimit:  20,
		},
		{
			name:           "order ascending alternate keyword",
			query:          "?sort=name&order=ascending",
			expectedSearch: "",
			expectedName:   "",
			expectedSort:   "name",
			expectedOrder:  "asc",
			expectedPage:   1,
			expectedLimit:  20,
		},
		{
			name:           "all parameters combined",
			query:          "?search=api&name=gateway&sort=name&order=asc&page=2&limit=10",
			expectedSearch: "api",
			expectedName:   "gateway",
			expectedSort:   "name",
			expectedOrder:  "asc",
			expectedPage:   2,
			expectedLimit:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/services"+tt.query, nil)
			params := handler.ParseListParams(req)

			assert.Equal(t, tt.expectedSearch, params.Search)
			assert.Equal(t, tt.expectedName, params.Name)
			assert.Equal(t, tt.expectedSort, params.Sort)
			assert.Equal(t, tt.expectedOrder, params.Order)
			assert.Equal(t, tt.expectedPage, params.Pagination.Page)
			assert.Equal(t, tt.expectedLimit, params.Pagination.Limit)
		})
	}
}
