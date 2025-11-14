package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the category of an error
type ErrorType string

const (
	// ErrorTypeDatabase represents database-related errors
	ErrorTypeDatabase ErrorType = "database"
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeNotFound represents resource not found errors
	ErrorTypeNotFound ErrorType = "not_found"
	// ErrorTypeInternal represents internal server errors
	ErrorTypeInternal ErrorType = "internal"
	// ErrorTypeExternal represents external service errors
	ErrorTypeExternal ErrorType = "external"
	// ErrorTypeAuthentication represents authentication errors
	ErrorTypeAuthentication ErrorType = "authentication"
	// ErrorTypeAuthorization represents authorization errors
	ErrorTypeAuthorization ErrorType = "authorization"
)

// AppError represents a structured application error
type AppError struct {
	Type    ErrorType
	Message string
	Err     error
	Context map[string]any
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value any) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// New creates a new AppError
func New(errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Context: make(map[string]any),
	}
}

// Wrap wraps an existing error with context
func Wrap(err error, errType ErrorType, message string) *AppError {
	if err == nil {
		return nil
	}

	return &AppError{
		Type:    errType,
		Message: message,
		Err:     err,
		Context: make(map[string]any),
	}
}

// Wrapf wraps an error with a formatted message
func Wrapf(err error, errType ErrorType, format string, args ...any) *AppError {
	if err == nil {
		return nil
	}

	return &AppError{
		Type:    errType,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
		Context: make(map[string]any),
	}
}

// Is checks if the error is of a specific type
func Is(err error, errType ErrorType) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == errType
	}
	return false
}

// GetContext retrieves context information from an error
func GetContext(err error, key string) (any, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		val, ok := appErr.Context[key]
		return val, ok
	}
	return nil, false
}

// Common database errors
var (
	ErrDatabaseConnection  = New(ErrorTypeDatabase, "database connection failed")
	ErrDatabaseQuery       = New(ErrorTypeDatabase, "database query failed")
	ErrDatabaseTransaction = New(ErrorTypeDatabase, "database transaction failed")
	ErrRecordNotFound      = New(ErrorTypeNotFound, "record not found")
	ErrDuplicateKey        = New(ErrorTypeDatabase, "duplicate key violation")
)

// Common validation errors
var (
	ErrInvalidInput    = New(ErrorTypeValidation, "invalid input")
	ErrMissingRequired = New(ErrorTypeValidation, "missing required field")
	ErrInvalidFormat   = New(ErrorTypeValidation, "invalid format")
)

// Common authentication/authorization errors
var (
	ErrUnauthorized = New(ErrorTypeAuthentication, "unauthorized")
	ErrForbidden    = New(ErrorTypeAuthorization, "forbidden")
	ErrTokenExpired = New(ErrorTypeAuthentication, "token expired")
	ErrInvalidToken = New(ErrorTypeAuthentication, "invalid token")
)
