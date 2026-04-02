package config

import "os"

// Environment variable names
const (
	EnvProfile   = "LM_PROFILE"
	EnvToken     = "LM_TOKEN"
	EnvChannelID = "LM_CHANNEL_ID"
	EnvSecret    = "LM_CHANNEL_SECRET"
	EnvFormat    = "LM_FORMAT"
	EnvConfigDir = "LM_CONFIG_DIR"
	EnvNoInput   = "LM_NO_INPUT"
	EnvEndpoint  = "LM_ENDPOINT"
	EnvDebug     = "LM_DEBUG"
	EnvYes       = "LM_YES"
)

// EnvOr returns the environment variable value if set, otherwise the fallback.
func EnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// IsYes returns true if auto-confirm mode is requested.
func IsYes() bool {
	v := os.Getenv(EnvYes)
	return v == "1" || v == "true"
}

// IsNoInput returns true if non-interactive mode is requested.
func IsNoInput() bool {
	v := os.Getenv(EnvNoInput)
	return v == "1" || v == "true"
}
