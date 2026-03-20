package bot

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/cmd/cmdutil"
	"github.com/crowdy/lm-cli/internal/api"
	"github.com/crowdy/lm-cli/internal/model"
	"github.com/crowdy/lm-cli/internal/output"
)

// Cmd is the bot command group.
var Cmd = &cobra.Command{
	Use:   "bot",
	Short: "Bot information and quota",
}

func init() {
	Cmd.AddCommand(infoCmd)
	Cmd.AddCommand(quotaCmd)
	Cmd.AddCommand(consumptionCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show bot profile information",
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

		rows := []model.BotInfoRow{{
			UserID:      info.UserID,
			BasicID:     info.BasicID,
			DisplayName: info.DisplayName,
			ChatMode:    info.ChatMode,
		}}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			return output.New(format).Format(os.Stdout, info)
		}
		return output.New(format).Format(os.Stdout, rows)
	},
}

var quotaCmd = &cobra.Command{
	Use:   "quota",
	Short: "Show message quota",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		botAPI := &api.BotAPI{Client: client}
		quota, err := botAPI.GetQuota()
		if err != nil {
			return err
		}

		rows := []model.QuotaRow{{Type: quota.Type, Value: quota.Value}}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			return output.New(format).Format(os.Stdout, quota)
		}
		return output.New(format).Format(os.Stdout, rows)
	},
}

var consumptionCmd = &cobra.Command{
	Use:   "consumption",
	Short: "Show message consumption (total usage this month)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		botAPI := &api.BotAPI{Client: client}
		consumption, err := botAPI.GetConsumption()
		if err != nil {
			return err
		}

		rows := []model.ConsumptionRow{{TotalUsage: consumption.TotalUsage}}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			return output.New(format).Format(os.Stdout, consumption)
		}
		return output.New(format).Format(os.Stdout, rows)
	},
}
