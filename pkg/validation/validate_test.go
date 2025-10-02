package validation

import (
	"testing"
)

func TestValidatePosition(t *testing.T) {
	tests := []struct {
		name      string
		position  string
		wantError bool
	}{
		{"Valid QB", "QB", false},
		{"Valid RB", "RB", false},
		{"Valid WR", "WR", false},
		{"Valid TE", "TE", false},
		{"Valid K", "K", false},
		{"Valid DEF", "DEF", false},
		{"Invalid position", "INVALID", true},
		{"Empty string", "", true},
		{"Lowercase valid", "qb", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePosition(tt.position)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidatePosition(%q) error = %v, wantError %v", tt.position, err, tt.wantError)
			}
		})
	}
}

func TestValidateStatus(t *testing.T) {
	tests := []struct {
		name      string
		status    string
		wantError bool
	}{
		{"Valid active", "active", false},
		{"Valid inactive", "inactive", false},
		{"Valid injured", "injured", false},
		{"Invalid status", "retired", true},
		{"Empty string", "", true},
		{"Uppercase", "ACTIVE", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatus(tt.status)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateStatus(%q) error = %v, wantError %v", tt.status, err, tt.wantError)
			}
		})
	}
}

func TestValidateLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		want  int
	}{
		{"Valid limit", 25, 25},
		{"Zero limit returns default", 0, 50},
		{"Negative limit returns default", -10, 50},
		{"Over max returns max", 150, 100},
		{"Max limit", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateLimit(tt.limit)
			if got != tt.want {
				t.Errorf("ValidateLimit(%d) = %d, want %d", tt.limit, got, tt.want)
			}
		})
	}
}

func TestValidateOffset(t *testing.T) {
	tests := []struct {
		name   string
		offset int
		want   int
	}{
		{"Valid offset", 50, 50},
		{"Zero offset", 0, 0},
		{"Negative offset returns zero", -10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateOffset(tt.offset)
			if got != tt.want {
				t.Errorf("ValidateOffset(%d) = %d, want %d", tt.offset, got, tt.want)
			}
		})
	}
}

func TestParseIntParam(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		defaultValue int
		want         int
	}{
		{"Valid integer", "42", 0, 42},
		{"Empty string returns default", "", 10, 10},
		{"Invalid string returns default", "abc", 20, 20},
		{"Negative integer", "-5", 0, -5},
		{"Zero", "0", 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseIntParam(tt.value, tt.defaultValue)
			if got != tt.want {
				t.Errorf("ParseIntParam(%q, %d) = %d, want %d", tt.value, tt.defaultValue, got, tt.want)
			}
		})
	}
}
