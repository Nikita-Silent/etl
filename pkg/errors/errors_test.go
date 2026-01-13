package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(ErrCodeDBConnection, "connection failed")

	if err.Code != ErrCodeDBConnection {
		t.Errorf("expected code %s, got %s", ErrCodeDBConnection, err.Code)
	}
	if err.Message != "connection failed" {
		t.Errorf("expected message 'connection failed', got %s", err.Message)
	}
	if err.Cause != nil {
		t.Errorf("expected nil cause, got %v", err.Cause)
	}
	if err.Context == nil {
		t.Error("expected non-nil context map")
	}
}

func TestNewf(t *testing.T) {
	err := Newf(ErrCodeDBQuery, "query failed: %s", "timeout")

	expectedMsg := "query failed: timeout"
	if err.Message != expectedMsg {
		t.Errorf("expected message %s, got %s", expectedMsg, err.Message)
	}
	if err.Code != ErrCodeDBQuery {
		t.Errorf("expected code %s, got %s", ErrCodeDBQuery, err.Code)
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := Wrap(ErrCodeFTPDownload, "download failed", originalErr)

	if err.Code != ErrCodeFTPDownload {
		t.Errorf("expected code %s, got %s", ErrCodeFTPDownload, err.Code)
	}
	if err.Message != "download failed" {
		t.Errorf("expected message 'download failed', got %s", err.Message)
	}
	if err.Cause != originalErr {
		t.Errorf("expected cause to be original error, got %v", err.Cause)
	}
}

func TestWrapf(t *testing.T) {
	originalErr := errors.New("connection timeout")
	err := Wrapf(ErrCodeFTPConnection, originalErr, "failed to connect to %s", "ftp.example.com")

	expectedMsg := "failed to connect to ftp.example.com"
	if err.Message != expectedMsg {
		t.Errorf("expected message %s, got %s", expectedMsg, err.Message)
	}
	if err.Cause != originalErr {
		t.Errorf("expected cause to be original error, got %v", err.Cause)
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name:     "error without cause",
			err:      New(ErrCodeValidation, "invalid input"),
			expected: "[VALIDATION] invalid input",
		},
		{
			name:     "error with cause",
			err:      Wrap(ErrCodeDBQuery, "query failed", errors.New("timeout")),
			expected: "[DB_QUERY] query failed: timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	originalErr := errors.New("original")
	wrappedErr := Wrap(ErrCodeDBConnection, "wrapped", originalErr)

	unwrapped := wrappedErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}

	// Test error without cause
	noWrapErr := New(ErrCodeValidation, "no wrap")
	if noWrapErr.Unwrap() != nil {
		t.Error("Unwrap() should return nil for error without cause")
	}
}

func TestIs(t *testing.T) {
	err1 := New(ErrCodeDBDeadlock, "deadlock")
	err2 := New(ErrCodeDBDeadlock, "another deadlock")
	err3 := New(ErrCodeFTPConnection, "ftp error")

	if !err1.Is(err2) {
		t.Error("errors with same code should match with Is()")
	}
	if err1.Is(err3) {
		t.Error("errors with different codes should not match with Is()")
	}

	// Test with non-AppError
	stdErr := errors.New("standard error")
	if err1.Is(stdErr) {
		t.Error("AppError should not match standard error")
	}
}

func TestErrorsIs(t *testing.T) {
	baseErr := New(ErrCodeDBDeadlock, "deadlock")
	wrappedErr := Wrap(ErrCodeDBQuery, "query failed", baseErr)

	// errors.Is should work with wrapped errors
	if !errors.Is(wrappedErr, baseErr) {
		t.Error("errors.Is should match wrapped AppError")
	}

	targetErr := New(ErrCodeDBDeadlock, "target")
	if !errors.Is(wrappedErr, targetErr) {
		t.Error("errors.Is should match by error code")
	}
}

func TestErrorsAs(t *testing.T) {
	originalErr := errors.New("original")
	appErr := Wrap(ErrCodeFTPDownload, "download failed", originalErr)

	var target *AppError
	if !errors.As(appErr, &target) {
		t.Error("errors.As should work with AppError")
	}
	if target.Code != ErrCodeFTPDownload {
		t.Errorf("extracted error code = %v, want %v", target.Code, ErrCodeFTPDownload)
	}
}

func TestWithContext(t *testing.T) {
	err := New(ErrCodeDBQuery, "query failed")
	err = err.WithContext("query", "SELECT * FROM users")

	if err.Context["query"] != "SELECT * FROM users" {
		t.Errorf("context value = %v, want 'SELECT * FROM users'", err.Context["query"])
	}

	// Test method chaining
	err2 := New(ErrCodeFTPDownload, "download failed").
		WithContext("file", "data.txt").
		WithContext("size", 1024)

	if err2.Context["file"] != "data.txt" {
		t.Errorf("context file = %v, want 'data.txt'", err2.Context["file"])
	}
	if err2.Context["size"] != 1024 {
		t.Errorf("context size = %v, want 1024", err2.Context["size"])
	}
}

func TestWithContextMap(t *testing.T) {
	err := New(ErrCodeParseInvalid, "parse failed")
	ctx := map[string]interface{}{
		"line":   42,
		"column": 10,
		"file":   "data.txt",
	}
	err = err.WithContextMap(ctx)

	if err.Context["line"] != 42 {
		t.Errorf("context line = %v, want 42", err.Context["line"])
	}
	if err.Context["column"] != 10 {
		t.Errorf("context column = %v, want 10", err.Context["column"])
	}
	if err.Context["file"] != "data.txt" {
		t.Errorf("context file = %v, want 'data.txt'", err.Context["file"])
	}
}

func TestGetContext(t *testing.T) {
	err := New(ErrCodeValidation, "validation failed").
		WithContext("field", "email")

	ctx := err.GetContext()
	if ctx["field"] != "email" {
		t.Errorf("GetContext()['field'] = %v, want 'email'", ctx["field"])
	}
}

func TestGetCode(t *testing.T) {
	err := New(ErrCodeTimeout, "timeout")
	if code := err.GetCode(); code != ErrCodeTimeout {
		t.Errorf("GetCode() = %v, want %v", code, ErrCodeTimeout)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		code      ErrorCode
		retryable bool
	}{
		{"deadlock", ErrCodeDBDeadlock, true},
		{"serialization", ErrCodeDBSerialization, true},
		{"ftp connection", ErrCodeFTPConnection, true},
		{"timeout", ErrCodeTimeout, true},
		{"validation", ErrCodeValidation, false},
		{"parse error", ErrCodeParseInvalid, false},
		{"db query", ErrCodeDBQuery, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.code, "test error")
			if got := err.IsRetryable(); got != tt.retryable {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.retryable)
			}
		})
	}
}

func TestIsCode(t *testing.T) {
	err := New(ErrCodeDBConnection, "connection failed")

	if !IsCode(err, ErrCodeDBConnection) {
		t.Error("IsCode should return true for matching code")
	}
	if IsCode(err, ErrCodeFTPConnection) {
		t.Error("IsCode should return false for non-matching code")
	}

	// Test with wrapped error
	wrappedErr := Wrap(ErrCodeDBQuery, "query failed", err)
	if !IsCode(wrappedErr, ErrCodeDBQuery) {
		t.Error("IsCode should work with wrapped errors")
	}

	// Test with standard error
	stdErr := errors.New("standard error")
	if IsCode(stdErr, ErrCodeInternal) {
		t.Error("IsCode should return false for standard errors")
	}
}

func TestGetCodeFunc(t *testing.T) {
	err := New(ErrCodeFTPDownload, "download failed")

	if code := GetCode(err); code != ErrCodeFTPDownload {
		t.Errorf("GetCode() = %v, want %v", code, ErrCodeFTPDownload)
	}

	// Test with wrapped error
	wrappedErr := Wrap(ErrCodeDBQuery, "query failed", err)
	if code := GetCode(wrappedErr); code != ErrCodeDBQuery {
		t.Errorf("GetCode() on wrapped error = %v, want %v", code, ErrCodeDBQuery)
	}

	// Test with standard error
	stdErr := errors.New("standard error")
	if code := GetCode(stdErr); code != ErrCodeInternal {
		t.Errorf("GetCode() on standard error = %v, want %v", code, ErrCodeInternal)
	}
}

func TestIsRetryableFunc(t *testing.T) {
	retryableErr := New(ErrCodeDBDeadlock, "deadlock")
	if !IsRetryable(retryableErr) {
		t.Error("IsRetryable should return true for deadlock error")
	}

	nonRetryableErr := New(ErrCodeValidation, "validation failed")
	if IsRetryable(nonRetryableErr) {
		t.Error("IsRetryable should return false for validation error")
	}

	// Test with standard error
	stdErr := errors.New("standard error")
	if IsRetryable(stdErr) {
		t.Error("IsRetryable should return false for standard errors")
	}
}

func TestGetContextFunc(t *testing.T) {
	err := New(ErrCodeParseInvalid, "parse failed").
		WithContext("line", 42)

	ctx := GetContext(err)
	if ctx == nil {
		t.Fatal("GetContext should return non-nil context")
	}
	if ctx["line"] != 42 {
		t.Errorf("GetContext()['line'] = %v, want 42", ctx["line"])
	}

	// Test with standard error
	stdErr := errors.New("standard error")
	if ctx := GetContext(stdErr); ctx != nil {
		t.Error("GetContext should return nil for standard errors")
	}
}

func TestNewDBError(t *testing.T) {
	cause := errors.New("connection timeout")
	err := NewDBError("database operation failed", cause)

	if err.Code != ErrCodeDBQuery {
		t.Errorf("NewDBError code = %v, want %v", err.Code, ErrCodeDBQuery)
	}
	if err.Cause != cause {
		t.Error("NewDBError should wrap cause")
	}
}

func TestNewFTPError(t *testing.T) {
	cause := errors.New("connection refused")
	err := NewFTPError("FTP operation failed", cause)

	if err.Code != ErrCodeFTPConnection {
		t.Errorf("NewFTPError code = %v, want %v", err.Code, ErrCodeFTPConnection)
	}
	if err.Cause != cause {
		t.Error("NewFTPError should wrap cause")
	}
}

func TestNewParseError(t *testing.T) {
	cause := errors.New("invalid format")
	err := NewParseError("parsing failed", cause)

	if err.Code != ErrCodeParseInvalid {
		t.Errorf("NewParseError code = %v, want %v", err.Code, ErrCodeParseInvalid)
	}
	if err.Cause != cause {
		t.Error("NewParseError should wrap cause")
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("invalid input")

	if err.Code != ErrCodeValidation {
		t.Errorf("NewValidationError code = %v, want %v", err.Code, ErrCodeValidation)
	}
	if err.Cause != nil {
		t.Error("NewValidationError should not have a cause")
	}
}

func TestNewTimeoutError(t *testing.T) {
	err := NewTimeoutError("operation timed out")

	if err.Code != ErrCodeTimeout {
		t.Errorf("NewTimeoutError code = %v, want %v", err.Code, ErrCodeTimeout)
	}
	if !err.IsRetryable() {
		t.Error("timeout errors should be retryable")
	}
}

func TestNewQueueFullError(t *testing.T) {
	err := NewQueueFullError("load-queue")

	if err.Code != ErrCodeQueueFull {
		t.Errorf("NewQueueFullError code = %v, want %v", err.Code, ErrCodeQueueFull)
	}
	expectedMsg := "queue load-queue is full"
	if err.Message != expectedMsg {
		t.Errorf("NewQueueFullError message = %v, want %v", err.Message, expectedMsg)
	}
}

func TestErrorChaining(t *testing.T) {
	// Create a chain of errors
	baseErr := errors.New("base error")
	level1 := Wrap(ErrCodeDBConnection, "level 1", baseErr)
	level2 := Wrap(ErrCodeDBQuery, "level 2", level1)

	// Test Unwrap chain
	if unwrapped := errors.Unwrap(level2); unwrapped != level1 {
		t.Errorf("first Unwrap() = %v, want level1", unwrapped)
	}

	// Test errors.Is works through chain
	if !errors.Is(level2, level1) {
		t.Error("errors.Is should work through error chain")
	}

	// Test GetCode gets the outermost code
	if code := GetCode(level2); code != ErrCodeDBQuery {
		t.Errorf("GetCode() = %v, want %v", code, ErrCodeDBQuery)
	}
}

func TestContextImmutability(t *testing.T) {
	err := New(ErrCodeValidation, "validation failed")
	err.WithContext("field", "email")

	// Context should be mutable through the same reference
	err.WithContext("value", "invalid@")

	if len(err.Context) != 2 {
		t.Errorf("context length = %d, want 2", len(err.Context))
	}
}

func TestNilContextHandling(t *testing.T) {
	err := &AppError{
		Code:    ErrCodeInternal,
		Message: "test",
		Context: nil,
	}

	// WithContext should initialize nil context
	err.WithContext("key", "value")
	if err.Context == nil {
		t.Error("WithContext should initialize nil context")
	}
	if err.Context["key"] != "value" {
		t.Error("WithContext should set value after initializing nil context")
	}

	// WithContextMap should also initialize nil context
	err2 := &AppError{
		Code:    ErrCodeInternal,
		Message: "test2",
		Context: nil,
	}
	err2.WithContextMap(map[string]interface{}{"foo": "bar"})
	if err2.Context == nil {
		t.Error("WithContextMap should initialize nil context")
	}
}

func ExampleAppError() {
	// Create a simple error
	err := New(ErrCodeValidation, "invalid email format")
	fmt.Println(err)

	// Create error with context
	err2 := New(ErrCodeDBQuery, "query failed").
		WithContext("query", "SELECT * FROM users").
		WithContext("duration", "5s")
	fmt.Println(err2)

	// Wrap an existing error
	originalErr := errors.New("connection timeout")
	err3 := Wrap(ErrCodeFTPConnection, "failed to connect to FTP server", originalErr)
	fmt.Println(err3)

	// Output:
	// [VALIDATION] invalid email format
	// [DB_QUERY] query failed
	// [FTP_CONNECTION] failed to connect to FTP server: connection timeout
}

func ExampleIsRetryable() {
	err1 := New(ErrCodeDBDeadlock, "deadlock detected")
	fmt.Println("Deadlock retryable:", err1.IsRetryable())

	err2 := New(ErrCodeValidation, "invalid input")
	fmt.Println("Validation retryable:", err2.IsRetryable())

	// Output:
	// Deadlock retryable: true
	// Validation retryable: false
}
