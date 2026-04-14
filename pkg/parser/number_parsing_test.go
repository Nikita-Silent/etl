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
