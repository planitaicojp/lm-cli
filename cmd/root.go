package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/cmd/auth"
	cmdconfig "github.com/crowdy/lm-cli/cmd/config"
	"github.com/crowdy/lm-cli/cmd/audience"
	"github.com/crowdy/lm-cli/cmd/bot"
	"github.com/crowdy/lm-cli/cmd/content"
	"github.com/crowdy/lm-cli/cmd/group"
	"github.com/crowdy/lm-cli/cmd/insight"
	"github.com/crowdy/lm-cli/cmd/message"
	"github.com/crowdy/lm-cli/cmd/richmenu"
	"github.com/crowdy/lm-cli/cmd/status"
	"github.com/crowdy/lm-cli/cmd/user"
	"github.com/crowdy/lm-cli/cmd/skill"
	"github.com/crowdy/lm-cli/cmd/webhook"
	"github.com/crowdy/lm-cli/internal/api"
	"github.com/crowdy/lm-cli/internal/config"
	lmerrors "github.com/crowdy/lm-cli/internal/errors"
)

var (
	version = "dev"

	flagProfile string
	flagFormat  string
	flagNoInput bool
	flagQuiet   bool
	flagVerbose bool
	flagNoColor bool
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:           "lm",
	Short:         "LINE Messaging API CLI",
	Long:          "Command-line interface for the LINE Messaging API",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		api.UserAgent = "crowdy/lm-cli/" + version
		if flagVerbose {
			api.SetDebugLevel(api.DebugVerbose)
		}
		if flagNoInput {
			_ = os.Setenv(config.EnvNoInput, "1")
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagProfile, "profile", "", "config profile to use")
	rootCmd.PersistentFlags().StringVar(&flagFormat, "format", "", "output format: table, json, yaml, csv")
	rootCmd.PersistentFlags().BoolVar(&flagNoInput, "no-input", false, "disable interactive prompts")
	rootCmd.PersistentFlags().BoolVar(&flagQuiet, "quiet", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "verbose HTTP logging")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "disable color output")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(auth.Cmd)
	rootCmd.AddCommand(cmdconfig.Cmd)
	rootCmd.AddCommand(message.Cmd)
	rootCmd.AddCommand(bot.Cmd)
	rootCmd.AddCommand(user.Cmd)
	rootCmd.AddCommand(group.Cmd)
	rootCmd.AddCommand(richmenu.Cmd)
	rootCmd.AddCommand(webhook.Cmd)
	rootCmd.AddCommand(audience.Cmd)
	rootCmd.AddCommand(insight.Cmd)
	rootCmd.AddCommand(content.Cmd)
	rootCmd.AddCommand(status.Cmd)
	rootCmd.AddCommand(skill.Cmd)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(lmerrors.GetExitCode(err))
	}
}

// GetProfile returns the active profile name.
func GetProfile() string {
	if flagProfile != "" {
		return flagProfile
	}
	if p := config.EnvOr(config.EnvProfile, ""); p != "" {
		return p
	}
	cfg, err := config.Load()
	if err != nil {
		return "default"
	}
	if cfg.ActiveProfile != "" {
		return cfg.ActiveProfile
	}
	return "default"
}

// GetFormat returns the output format.
func GetFormat() string {
	if flagFormat != "" {
		return flagFormat
	}
	if f := config.EnvOr(config.EnvFormat, ""); f != "" {
		return f
	}
	cfg, err := config.Load()
	if err != nil {
		return config.DefaultFormat
	}
	if cfg.Defaults.Format != "" {
		return cfg.Defaults.Format
	}
	return config.DefaultFormat
}

// IsQuiet returns whether quiet mode is enabled.
func IsQuiet() bool {
	return flagQuiet
}
