package config

import "testing"

func TestIsYes(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want bool
	}{
		{"empty", "", false},
		{"1", "1", true},
		{"true", "true", true},
		{"false", "false", false},
		{"other", "yes", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LM_YES", tt.val)
			if got := IsYes(); got != tt.want {
				t.Errorf("IsYes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNoInput(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want bool
	}{
		{"empty", "", false},
		{"1", "1", true},
		{"true", "true", true},
		{"false", "false", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LM_NO_INPUT", tt.val)
			if got := IsNoInput(); got != tt.want {
				t.Errorf("IsNoInput() = %v, want %v", got, tt.want)
			}
		})
	}
}
