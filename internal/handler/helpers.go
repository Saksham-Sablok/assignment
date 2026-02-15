package handler

import (
	"net/http"
	"strconv"

	"github.com/services-api/internal/domain"
)

const (
	// DefaultPage is the default page number
	DefaultPage = 1
	// DefaultLimit is the default number of items per page
	DefaultLimit = 20
	// MaxLimit is the maximum number of items per page
	MaxLimit = 100
)

// ParsePaginationParams parses pagination parameters from query string
func ParsePaginationParams(r *http.Request) domain.PaginationParams {
	params := domain.PaginationParams{
		Page:  DefaultPage,
		Limit: DefaultLimit,
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			params.Limit = limit
			if params.Limit > MaxLimit {
				params.Limit = MaxLimit
			}
		}
	}

	return params
}

// ParseListParams parses list parameters from query string including filtering and sorting
func ParseListParams(r *http.Request) domain.ListParams {
	params := domain.ListParams{
		Pagination: ParsePaginationParams(r),
	}

	// Parse search parameter (searches across name and description)
	if search := r.URL.Query().Get("search"); search != "" {
		params.Search = search
	}

	// Parse name filter (exact or partial match on name)
	if name := r.URL.Query().Get("name"); name != "" {
		params.Name = name
	}

	// Parse sort field
	if sort := r.URL.Query().Get("sort"); sort != "" {
		params.Sort = sort
	} else {
		params.Sort = "created_at"
	}

	// Parse sort order
	if order := r.URL.Query().Get("order"); order != "" {
		// Normalize to "asc" or "desc"
		if order == "asc" || order == "ascending" {
			params.Order = "asc"
		} else {
			params.Order = "desc"
		}
	} else {
		params.Order = "desc"
	}

	return params
}
