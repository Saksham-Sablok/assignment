package domain

import "errors"

// Common errors
var (
	ErrNotFound            = errors.New("resource not found")
	ErrDuplicateVersion    = errors.New("version already exists for this service")
	ErrValidation          = errors.New("validation error")
	ErrNameRequired        = errors.New("name is required")
	ErrDescriptionRequired = errors.New("description is required")
	ErrVersionRequired     = errors.New("version is required")
	ErrNameTooLong         = errors.New("name must be at most 255 characters")
	ErrDescriptionTooLong  = errors.New("description must be at most 1000 characters")
	ErrInvalidSortField    = errors.New("invalid sort field")
	ErrInvalidID           = errors.New("invalid ID format")
)

// ValidationError wraps validation errors with details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return e.Message
}
