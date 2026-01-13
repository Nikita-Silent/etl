package validation

import (
	"testing"
	"time"
)

func TestRequiredValidator(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{
			name:      "valid string",
			value:     "test",
			wantError: false,
		},
		{
			name:      "empty string",
			value:     "",
			wantError: true,
		},
		{
			name:      "nil value",
			value:     nil,
			wantError: true,
		},
		{
			name:      "valid number",
			value:     123,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := Required("test_field")
			err := validator.Validate(tt.value)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestDateFormatValidator(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		format    string
		wantError bool
	}{
		{
			name:      "valid date",
			value:     "2024-12-01",
			format:    "2006-01-02",
			wantError: false,
		},
		{
			name:      "invalid date format",
			value:     "12/01/2024",
			format:    "2006-01-02",
			wantError: true,
		},
		{
			name:      "empty string (allowed)",
			value:     "",
			format:    "2006-01-02",
			wantError: false,
		},
		{
			name:      "not a string",
			value:     123,
			format:    "2006-01-02",
			wantError: true,
		},
		{
			name:      "invalid date",
			value:     "2024-13-01",
			format:    "2006-01-02",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := DateFormat("date", tt.format)
			err := validator.Validate(tt.value)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestNotInFutureValidator(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	tests := []struct {
		name      string
		value     interface{}
		format    string
		wantError bool
	}{
		{
			name:      "today (valid)",
			value:     today,
			format:    "2006-01-02",
			wantError: false,
		},
		{
			name:      "yesterday (valid)",
			value:     yesterday,
			format:    "2006-01-02",
			wantError: false,
		},
		{
			name:      "tomorrow (invalid)",
			value:     tomorrow,
			format:    "2006-01-02",
			wantError: true,
		},
		{
			name:      "empty string (allowed)",
			value:     "",
			format:    "2006-01-02",
			wantError: false,
		},
		{
			name:      "invalid format",
			value:     "12/01/2024",
			format:    "2006-01-02",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NotInFuture("date", tt.format)
			err := validator.Validate(tt.value)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestKassaCodeValidator(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{
			name:      "valid kassa code",
			value:     "P13",
			wantError: false,
		},
		{
			name:      "valid kassa code with folder",
			value:     "P13/P13",
			wantError: false,
		},
		{
			name:      "empty string (allowed)",
			value:     "",
			wantError: false,
		},
		{
			name:      "not a string",
			value:     123,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := KassaCode("kassa_code")
			err := validator.Validate(tt.value)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestCompositeValidator(t *testing.T) {
	tests := []struct {
		name       string
		validators []Validator
		value      interface{}
		wantError  bool
	}{
		{
			name: "all validators pass",
			validators: []Validator{
				Required("date"),
				DateFormat("date", "2006-01-02"),
			},
			value:     "2024-12-01",
			wantError: false,
		},
		{
			name: "required validator fails",
			validators: []Validator{
				Required("date"),
				DateFormat("date", "2006-01-02"),
			},
			value:     "",
			wantError: true,
		},
		{
			name: "date format validator fails",
			validators: []Validator{
				Required("date"),
				DateFormat("date", "2006-01-02"),
			},
			value:     "invalid-date",
			wantError: true,
		},
		{
			name: "complex validation chain",
			validators: []Validator{
				Required("date"),
				DateFormat("date", "2006-01-02"),
				NotInFuture("date", "2006-01-02"),
			},
			value:     time.Now().Format("2006-01-02"),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewComposite(tt.validators...)
			err := validator.Validate(tt.value)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	t.Run("create validation error", func(t *testing.T) {
		err := NewValidationError("test_field", "test message")
		if err.Field != "test_field" {
			t.Errorf("Expected field 'test_field', got '%s'", err.Field)
		}
		if err.Message != "test message" {
			t.Errorf("Expected message 'test message', got '%s'", err.Message)
		}
	})

	t.Run("error string representation", func(t *testing.T) {
		err := NewValidationError("test_field", "test message")
		expected := "test_field: test message"
		if err.Error() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("is validation error", func(t *testing.T) {
		err := NewValidationError("test_field", "test message")
		if !IsValidationError(err) {
			t.Error("Expected IsValidationError to return true")
		}
	})
}
