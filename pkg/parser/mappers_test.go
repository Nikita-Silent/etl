package parser

import "testing"

func TestParseFloatWithComma(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{name: "dot_decimal", input: "1.25", want: 1.25},
		{name: "comma_decimal", input: "1,25", want: 1.25},
		{name: "integer", input: "42", want: 42},
		{name: "negative", input: "-2,5", want: -2.5},
		{name: "mixed_separators", input: "1,234.56", want: 1.234},
		{name: "thousands_separator", input: "1,000", wantErr: true},
		{name: "multiple_commas", input: "1,234,567", wantErr: true},
		{name: "invalid_chars", input: "abc", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFloatWithComma(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseFloatWithComma() expected error, got nil")
				}
				if got != 0 {
					t.Fatalf("parseFloatWithComma() = %v, want 0 on error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseFloatWithComma() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("parseFloatWithComma() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeParseInt(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "empty", input: "", want: 0},
		{name: "zero", input: "0", want: 0},
		{name: "positive", input: "42", want: 42},
		{name: "negative", input: "-5", want: -5},
		{name: "invalid", input: "abc", want: 0},
		{name: "float_string", input: "12.3", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := safeParseInt(tt.input); got != tt.want {
				t.Fatalf("safeParseInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeParseInt64(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{name: "empty", input: "", want: 0},
		{name: "zero", input: "0", want: 0},
		{name: "positive", input: "9223372036854775807", want: 9223372036854775807},
		{name: "negative", input: "-10", want: -10},
		{name: "overflow", input: "9223372036854775808", want: 0},
		{name: "invalid", input: "nope", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := safeParseInt64(tt.input); got != tt.want {
				t.Fatalf("safeParseInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeParseFloat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{name: "empty", input: "", want: 0},
		{name: "dot_decimal", input: "1.25", want: 1.25},
		{name: "comma_decimal", input: "1,25", want: 1.25},
		{name: "negative", input: "-2,5", want: -2.5},
		{name: "invalid", input: "abc", want: 0},
		{name: "thousands_separator", input: "1,000", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := safeParseFloat(tt.input); got != tt.want {
				t.Fatalf("safeParseFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}
