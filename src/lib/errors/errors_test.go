package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name: "error with wrapped error",
			err: &AppError{
				Type:    ErrorTypeDatabase,
				Message: "query failed",
				Err:     errors.New("connection timeout"),
			},
			expected: "[database] query failed: connection timeout",
		},
		{
			name: "error without wrapped error",
			err: &AppError{
				Type:    ErrorTypeValidation,
				Message: "invalid email",
			},
			expected: "[validation] invalid email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := Wrap(originalErr, ErrorTypeDatabase, "database operation failed")

	assert.NotNil(t, wrapped)
	assert.Equal(t, ErrorTypeDatabase, wrapped.Type)
	assert.Equal(t, "database operation failed", wrapped.Message)
	assert.True(t, errors.Is(wrapped, originalErr))
}

func TestWrapf(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := Wrapf(originalErr, ErrorTypeDatabase, "failed to fetch user with id=%d", 123)

	assert.NotNil(t, wrapped)
	assert.Contains(t, wrapped.Error(), "id=123")
}

func TestWrap_NilError(t *testing.T) {
	wrapped := Wrap(nil, ErrorTypeDatabase, "test")
	assert.Nil(t, wrapped)
}

func TestIs(t *testing.T) {
	err := New(ErrorTypeDatabase, "test error")

	assert.True(t, Is(err, ErrorTypeDatabase))
	assert.False(t, Is(err, ErrorTypeValidation))

	stdErr := errors.New("standard error")
	assert.False(t, Is(stdErr, ErrorTypeDatabase))
}

func TestWithContext(t *testing.T) {
	err := New(ErrorTypeDatabase, "query failed").
		WithContext("userID", 123).
		WithContext("table", "users")

	assert.NotNil(t, err.Context)
	assert.Equal(t, 123, err.Context["userID"])
	assert.Equal(t, "users", err.Context["table"])
}

func TestGetContext(t *testing.T) {
	err := New(ErrorTypeDatabase, "test").WithContext("key", "value")

	val, ok := GetContext(err, "key")
	assert.True(t, ok)
	assert.Equal(t, "value", val)

	val, ok = GetContext(err, "nonexistent")
	assert.False(t, ok)
	assert.Nil(t, val)
}
