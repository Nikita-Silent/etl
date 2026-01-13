package errors

import (
	"errors"
	"fmt"
)

// ErrorCode represents a categorized error code
type ErrorCode string

// Error codes for different error categories
const (
	// Database errors
	ErrCodeDBConnection    ErrorCode = "DB_CONNECTION"
	ErrCodeDBDeadlock      ErrorCode = "DB_DEADLOCK"
	ErrCodeDBSerialization ErrorCode = "DB_SERIALIZATION"
	ErrCodeDBQuery         ErrorCode = "DB_QUERY"
	ErrCodeDBTransaction   ErrorCode = "DB_TRANSACTION"

	// FTP errors
	ErrCodeFTPConnection ErrorCode = "FTP_CONNECTION"
	ErrCodeFTPDownload   ErrorCode = "FTP_DOWNLOAD"
	ErrCodeFTPUpload     ErrorCode = "FTP_UPLOAD"
	ErrCodeFTPList       ErrorCode = "FTP_LIST"

	// Parsing errors
	ErrCodeParseInvalid ErrorCode = "PARSE_INVALID"
	ErrCodeParseFormat  ErrorCode = "PARSE_FORMAT"
	ErrCodeParseHeader  ErrorCode = "PARSE_HEADER"

	// Validation errors
	ErrCodeValidation       ErrorCode = "VALIDATION"
	ErrCodeValidationDate   ErrorCode = "VALIDATION_DATE"
	ErrCodeValidationConfig ErrorCode = "VALIDATION_CONFIG"

	// Configuration errors
	ErrCodeConfig ErrorCode = "CONFIG"

	// Queue errors
	ErrCodeQueueFull ErrorCode = "QUEUE_FULL"

	// Timeout errors
	ErrCodeTimeout ErrorCode = "TIMEOUT"

	// Unknown/Internal errors
	ErrCodeInternal ErrorCode = "INTERNAL"
)

// AppError represents a structured application error
type AppError struct {
	Code    ErrorCode              // Error code for categorization
	Message string                 // Human-readable error message
	Cause   error                  // Underlying error (for wrapping)
	Context map[string]interface{} // Additional context data
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Newf creates a new AppError with formatted message
func Newf(code ErrorCode, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with an AppError
func Wrap(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// Wrapf wraps an existing error with a formatted message
func Wrapf(code ErrorCode, cause error, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for errors.Is/As compatibility
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is implements error matching for errors.Is
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	// Match by error code
	return t.Code == e.Code
}

// WithContext adds context data to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithContextMap adds multiple context values
func (e *AppError) WithContextMap(ctx map[string]interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	for k, v := range ctx {
		e.Context[k] = v
	}
	return e
}

// GetContext returns the context data
func (e *AppError) GetContext() map[string]interface{} {
	return e.Context
}

// GetCode returns the error code
func (e *AppError) GetCode() ErrorCode {
	return e.Code
}

// IsRetryable returns true if the error is retryable
func (e *AppError) IsRetryable() bool {
	switch e.Code {
	case ErrCodeDBDeadlock, ErrCodeDBSerialization, ErrCodeFTPConnection, ErrCodeTimeout:
		return true
	default:
		return false
	}
}

// IsCode checks if the error has a specific code
func IsCode(err error, code ErrorCode) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// GetCode extracts the error code from an error
func GetCode(err error) ErrorCode {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ErrCodeInternal
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.IsRetryable()
	}
	return false
}

// GetContext extracts context from an error
func GetContext(err error) map[string]interface{} {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.GetContext()
	}
	return nil
}

// Common error constructors for convenience

// NewDBError creates a database error
func NewDBError(message string, cause error) *AppError {
	return Wrap(ErrCodeDBQuery, message, cause)
}

// NewFTPError creates an FTP error
func NewFTPError(message string, cause error) *AppError {
	return Wrap(ErrCodeFTPConnection, message, cause)
}

// NewParseError creates a parsing error
func NewParseError(message string, cause error) *AppError {
	return Wrap(ErrCodeParseInvalid, message, cause)
}

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return New(ErrCodeValidation, message)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(message string) *AppError {
	return New(ErrCodeTimeout, message)
}

// NewQueueFullError creates a queue full error
func NewQueueFullError(queueName string) *AppError {
	return New(ErrCodeQueueFull, fmt.Sprintf("queue %s is full", queueName))
}
