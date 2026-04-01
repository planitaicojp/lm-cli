package errors

import (
	"errors"
	"fmt"
)

// ExitCoder is implemented by errors that carry a process exit code.
type ExitCoder interface {
	ExitCode() int
}

// APIError represents an error returned by the LINE Messaging API.
type APIError struct {
	StatusCode int
	Message    string
	Details    []APIErrorDetail
}

// APIErrorDetail is a single validation detail in a LINE API error response.
type APIErrorDetail struct {
	Message  string `json:"message"`
	Property string `json:"property"`
}

func (e *APIError) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("API error (HTTP %d): %s (%s: %s)",
			e.StatusCode, e.Message, e.Details[0].Property, e.Details[0].Message)
	}
	return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
}

func (e *APIError) ExitCode() int {
	return ExitAPI
}

// AuthError represents an authentication or authorization failure.
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("auth error: %s", e.Message)
}

func (e *AuthError) ExitCode() int {
	return ExitAuth
}

// ConfigError represents a configuration problem.
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error: %s", e.Message)
}

func (e *ConfigError) ExitCode() int {
	return ExitGeneral
}

// NotFoundError indicates that a requested resource was not found.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

func (e *NotFoundError) ExitCode() int {
	return ExitNotFound
}

// ValidationError represents invalid user input.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

func (e *ValidationError) ExitCode() int {
	return ExitValidation
}

// NetworkError wraps an underlying network-level error.
type NetworkError struct {
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %v", e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

func (e *NetworkError) ExitCode() int {
	return ExitNetwork
}

// RateLimitError represents a 429 Too Many Requests error.
type RateLimitError struct {
	RetryAfter string
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter != "" {
		return fmt.Sprintf("rate limit exceeded, retry after %s", e.RetryAfter)
	}
	return "rate limit exceeded"
}

func (e *RateLimitError) ExitCode() int {
	return ExitRateLimit
}

// CancelledError represents a user-cancelled operation.
type CancelledError struct{}

func (e *CancelledError) Error() string {
	return "cancelled"
}

func (e *CancelledError) ExitCode() int {
	return ExitCancelled
}

// GetExitCode returns the exit code for the given error.
// Supports wrapped errors (fmt.Errorf("%w", ...)).
func GetExitCode(err error) int {
	if err == nil {
		return ExitOK
	}
	var ec ExitCoder
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	return ExitGeneral
}
