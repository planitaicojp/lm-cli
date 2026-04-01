package status

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/cmd/cmdutil"
	"github.com/crowdy/lm-cli/internal/api"
	"github.com/crowdy/lm-cli/internal/config"
	"github.com/crowdy/lm-cli/internal/model"
	"github.com/crowdy/lm-cli/internal/output"
)

// Cmd is the status command.
var Cmd = &cobra.Command{
	Use:   "status",
	Short: "Check API connectivity and bot status",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		botAPI := &api.BotAPI{Client: client}
		info, err := botAPI.GetInfo()
		if err != nil {
			return err
		}

		// Resolve token info
		tokenType := "stateless"
		tokenExpiresAt := ""
		tokenExpiresIn := ""

		profileName := resolveProfile(cmd)
		tokens, tokErr := config.LoadTokens()
		if tokErr == nil {
			if entry, ok := tokens.Get(profileName); ok {
				if entry.TokenType != "" {
					tokenType = entry.TokenType
				}
				if !entry.ExpiresAt.IsZero() {
					tokenExpiresAt = entry.ExpiresAt.Format(time.RFC3339)
					tokenExpiresIn = formatDuration(time.Until(entry.ExpiresAt))
				}
			}
		}

		// If LM_TOKEN is set, it's a direct token with unknown expiry
		if config.EnvOr(config.EnvToken, "") != "" {
			tokenType = "env"
		}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			data := model.StatusInfo{
				API:            "ok",
				BotID:          info.BasicID,
				DisplayName:    info.DisplayName,
				TokenType:      tokenType,
				TokenExpiresAt: tokenExpiresAt,
			}
			return output.New(format).Format(os.Stdout, data)
		}

		rows := []model.StatusRow{{
			API:  "ok",
			Bot:  fmt.Sprintf("%s (%s)", info.BasicID, info.DisplayName),
			Token: formatTokenStatus(tokenType, tokenExpiresIn),
		}}
		return output.New(format).Format(os.Stdout, rows)
	},
}

func resolveProfile(cmd *cobra.Command) string {
	profileName, _ := cmd.Flags().GetString("profile")
	if profileName == "" {
		profileName = config.EnvOr(config.EnvProfile, "")
	}
	if profileName == "" {
		cfg, err := config.Load()
		if err == nil && cfg.ActiveProfile != "" {
			profileName = cfg.ActiveProfile
		}
	}
	if profileName == "" {
		profileName = "default"
	}
	return profileName
}

func formatTokenStatus(tokenType, expiresIn string) string {
	if tokenType == "env" {
		return "env (LM_TOKEN)"
	}
	if tokenType == "longterm" {
		return "valid (longterm)"
	}
	if expiresIn != "" {
		return fmt.Sprintf("valid (expires in %s)", expiresIn)
	}
	return "valid"
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		return "expired"
	}
	days := int(math.Floor(d.Hours() / 24))
	hours := int(math.Floor(d.Hours())) % 24
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	minutes := int(math.Floor(d.Minutes())) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
