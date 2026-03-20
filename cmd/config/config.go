package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/cmd/cmdutil"
	"github.com/crowdy/lm-cli/internal/config"
	lmerrors "github.com/crowdy/lm-cli/internal/errors"
)

// Cmd is the config command group.
var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

func init() {
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(setCmd)
	Cmd.AddCommand(listCmd)
}

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cmdutil.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		switch key {
		case "format":
			fmt.Println(cfg.Defaults.Format)
		case "active_profile":
			fmt.Println(cfg.ActiveProfile)
		default:
			return &lmerrors.ValidationError{Message: fmt.Sprintf("unknown key %q", key)}
		}
		return nil
	},
}

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cmdutil.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		switch key {
		case "format":
			validFormats := []string{"table", "json", "yaml", "csv"}
			valid := false
			for _, f := range validFormats {
				if value == f {
					valid = true
					break
				}
			}
			if !valid {
				return &lmerrors.ValidationError{
					Field:   "format",
					Message: fmt.Sprintf("must be one of: %s", strings.Join(validFormats, ", ")),
				}
			}
			cfg.Defaults.Format = value
		case "active_profile":
			if _, ok := cfg.Profiles[value]; !ok {
				return &lmerrors.ConfigError{Message: fmt.Sprintf("profile %q not found", value)}
			}
			cfg.ActiveProfile = value
		default:
			return &lmerrors.ValidationError{Message: fmt.Sprintf("unknown key %q", key)}
		}

		return cfg.Save()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		fmt.Printf("active_profile: %s\n", cfg.ActiveProfile)
		fmt.Printf("format:         %s\n", cfg.Defaults.Format)
		fmt.Printf("profiles:\n")
		for name, profile := range cfg.Profiles {
			marker := " "
			if name == cfg.ActiveProfile {
				marker = "*"
			}
			fmt.Printf("  %s %s (channel_id: %s, token_type: %s)\n",
				marker, name, profile.ChannelID, profile.TokenType)
		}
		return nil
	},
}
