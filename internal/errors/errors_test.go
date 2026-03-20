package errors

import (
	"testing"
)

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{"nil", nil, ExitOK},
		{"api_error", &APIError{StatusCode: 400, Message: "bad request"}, ExitAPI},
		{"auth_error", &AuthError{Message: "unauthorized"}, ExitAuth},
		{"not_found", &NotFoundError{Resource: "user", ID: "U123"}, ExitNotFound},
		{"validation", &ValidationError{Field: "text", Message: "required"}, ExitValidation},
		{"network", &NetworkError{Err: nil}, ExitNetwork},
		{"rate_limit", &RateLimitError{}, ExitRateLimit},
		{"config", &ConfigError{Message: "not found"}, ExitGeneral},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetExitCode(tt.err)
			if got != tt.wantCode {
				t.Errorf("GetExitCode(%T) = %d, want %d", tt.err, got, tt.wantCode)
			}
		})
	}
}

func TestAPIError_WithDetails(t *testing.T) {
	err := &APIError{
		StatusCode: 400,
		Message:    "The request body has 1 error(s)",
		Details: []APIErrorDetail{
			{Message: "May not be empty", Property: "messages[0].text"},
		},
	}

	s := err.Error()
	if s == "" {
		t.Error("expected non-empty error string")
	}
	// Should contain the detail info
	expected := "messages[0].text"
	found := false
	for i := 0; i <= len(s)-len(expected); i++ {
		if s[i:i+len(expected)] == expected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("error string %q should contain %q", s, expected)
	}
}
