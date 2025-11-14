# Error Handling Package

A structured error handling package for Stormkit that provides consistent error wrapping, context management, and error classification.

## Features

- **Structured Errors**: Categorize errors by type (Database, Validation, NotFound, etc.)
- **Error Wrapping**: Preserve error chains with `%w` formatting
- **Context Management**: Attach contextual information to errors
- **Type Checking**: Use `Is()` to check error types across the chain
- **Common Errors**: Pre-defined sentinel errors for common scenarios

## Error Types

```go
const (
    ErrorTypeDatabase        // Database-related errors
    ErrorTypeValidation      // Validation errors
    ErrorTypeNotFound        // Resource not found errors
    ErrorTypeInternal        // Internal server errors
    ErrorTypeExternal        // External service errors
    ErrorTypeAuthentication  // Authentication errors
    ErrorTypeAuthorization   // Authorization errors
)
```

## Usage Examples

### Creating New Errors

```go
import "github.com/stormkit-io/stormkit-io/src/lib/errors"

// Simple error
err := errors.New(errors.ErrorTypeValidation, "invalid email format")

// Error with context
err := errors.New(errors.ErrorTypeNotFound, "user not found").
    WithContext("user_id", 123).
    WithContext("email", "user@example.com")
```

### Wrapping Errors

```go
// Wrap existing error
if err != nil {
    return errors.Wrap(err, errors.ErrorTypeDatabase, "failed to fetch user")
}

// Wrap with formatted message
if err != nil {
    return errors.Wrapf(
        err,
        errors.ErrorTypeDatabase,
        "failed to fetch user with id=%d from table=%s",
        userID,
        tableName,
    )
}
```

### Checking Error Types

```go
err := someFunction()
if errors.Is(err, errors.ErrorTypeDatabase) {
    // Handle database errors specifically
    log.Error("Database operation failed")
}
```

### Retrieving Context

```go
if userID, ok := errors.GetContext(err, "user_id"); ok {
    log.Printf("Error occurred for user: %v", userID)
}
```

### Using Sentinel Errors

```go
import "github.com/stormkit-io/stormkit-io/src/lib/errors"

// Database errors
if err != nil {
    return errors.ErrDatabaseConnection
}

// Validation errors
if email == "" {
    return errors.ErrMissingRequired.WithContext("field", "email")
}

// Not found errors
if user == nil {
    return errors.ErrRecordNotFound.WithContext("user_id", userID)
}
```

## Real-World Examples

### Database Operations

```go
func (s *Store) UserByID(userID types.ID) (*User, error) {
    var wr bytes.Buffer
    
    data := map[string]any{
        "where": "u.user_id = $1 AND u.deleted_at IS NULL",
        "limit": 1,
    }
    
    if err := s.selectTmpl.Execute(&wr, data); err != nil {
        return nil, errors.Wrapf(
            err,
            errors.ErrorTypeInternal,
            "failed to execute template for user_id=%d",
            userID,
        )
    }
    
    users, err := s.selectUsers(context.TODO(), wr.String(), userID)
    if err != nil {
        return nil, errors.Wrapf(
            err,
            errors.ErrorTypeDatabase,
            "failed to fetch user with user_id=%d",
            userID,
        )
    }
    
    if len(users) == 0 {
        return nil, errors.New(errors.ErrorTypeNotFound, "user not found").
            WithContext("user_id", userID)
    }
    
    return users[0], nil
}
```

### API Handlers

```go
func handlerDeployGet(req *app.RequestContext) *shttp.Response {
    id := utils.StringToID(req.Vars()["deploymentId"])
    depl, err := deploy.NewStore().DeploymentByIDWithLogs(req.Context(), id, req.App.ID)
    
    if err != nil {
        wrappedErr := errors.Wrapf(
            err,
            errors.ErrorTypeDatabase,
            "failed to fetch deployment with id=%d for app_id=%d",
            id,
            req.App.ID,
        )
        return shttp.UnexpectedError(wrappedErr)
    }
    
    if depl == nil {
        return shttp.NotFound()
    }
    
    return &shttp.Response{
        Data: map[string]any{
            "deploy": depl,
        },
    }
}
```

### File Operations

```go
func openFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        if stderrors.Is(err, os.ErrNotExist) {
            return errors.Wrapf(
                err,
                errors.ErrorTypeNotFound,
                "file not found: %s",
                filename,
            )
        }
        return errors.Wrapf(
            err,
            errors.ErrorTypeInternal,
            "failed to open file: %s",
            filename,
        )
    }
    defer file.Close()
    
    // Process file...
    return nil
}
```

## Best Practices

1. **Always wrap errors with context**: Include relevant IDs, names, or parameters
2. **Use appropriate error types**: Choose the most specific error type
3. **Don't wrap errors multiple times**: Wrap once at the boundary where you have context
4. **Include structured logging**: Log errors with their full context
5. **Use sentinel errors for common cases**: Prefer predefined errors when applicable

## Migration Guide

### Before
```go
if err != nil {
    return err  // No context
}
```

### After
```go
if err != nil {
    return errors.Wrapf(
        err,
        errors.ErrorTypeDatabase,
        "failed to fetch user with id=%d",
        userID,
    )
}
```

## Testing

```go
func TestErrorWrapping(t *testing.T) {
    originalErr := fmt.Errorf("connection timeout")
    wrapped := errors.Wrap(originalErr, errors.ErrorTypeDatabase, "query failed")
    
    // Check error type
    assert.True(t, errors.Is(wrapped, errors.ErrorTypeDatabase))
    
    // Check error chain
    assert.True(t, stderrors.Is(wrapped, originalErr))
    
    // Check context
    wrapped = wrapped.WithContext("query", "SELECT * FROM users")
    val, ok := errors.GetContext(wrapped, "query")
    assert.True(t, ok)
    assert.Equal(t, "SELECT * FROM users", val)
}
```

## Integration with Existing Code

The error package is designed to work alongside existing error handling:

```go
import (
    stderrors "errors"  // Standard library
    "github.com/stormkit-io/stormkit-io/src/lib/errors"  // Our package
)

// Can still use standard error checking
if stderrors.Is(err, os.ErrNotExist) {
    // Wrap with our structured error
    return errors.Wrap(err, errors.ErrorTypeNotFound, "file not found")
}
```
