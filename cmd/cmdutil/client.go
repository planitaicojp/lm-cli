package cmdutil

import (
	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/internal/api"
	"github.com/crowdy/lm-cli/internal/config"
	lmerrors "github.com/crowdy/lm-cli/internal/errors"
)

// NewClient creates an API client from the cobra command context.
func NewClient(cmd *cobra.Command) (*api.Client, error) {
	profileName, _ := cmd.Flags().GetString("profile")
	if profileName == "" {
		profileName = config.EnvOr(config.EnvProfile, "")
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if profileName == "" {
		profileName = cfg.ActiveProfile
	}
	if profileName == "" {
		profileName = "default"
	}

	// Allow LM_TOKEN to bypass profile requirement
	if t := config.EnvOr(config.EnvToken, ""); t != "" {
		return api.NewClient(t)
	}

	if _, ok := cfg.Profiles[profileName]; !ok {
		return nil, &lmerrors.ConfigError{Message: "profile not found, run 'lm auth login'"}
	}

	creds, err := config.LoadCredentials()
	if err != nil {
		return nil, err
	}

	tokens, err := config.LoadTokens()
	if err != nil {
		return nil, err
	}

	token, err := api.EnsureToken(profileName, cfg, creds, tokens)
	if err != nil {
		return nil, err
	}

	return api.NewClient(token)
}
