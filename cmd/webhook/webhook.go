package webhook

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/cmd/cmdutil"
	"github.com/crowdy/lm-cli/internal/api"
	"github.com/crowdy/lm-cli/internal/model"
	"github.com/crowdy/lm-cli/internal/output"
)

// Cmd is the webhook command group.
var Cmd = &cobra.Command{
	Use:   "webhook",
	Short: "Webhook endpoint management",
}

func init() {
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(setCmd)
	Cmd.AddCommand(testCmd)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get webhook endpoint",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		webhookAPI := &api.WebhookAPI{Client: client}
		info, err := webhookAPI.Get()
		if err != nil {
			return err
		}

		rows := []model.WebhookInfoRow{{
			WebhookEndpoint: info.WebhookEndpoint,
			Active:          info.Active,
		}}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			return output.New(format).Format(os.Stdout, info)
		}
		return output.New(format).Format(os.Stdout, rows)
	},
}

var setCmd = &cobra.Command{
	Use:   "set <url>",
	Short: "Set webhook endpoint URL",
	Args:  cmdutil.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		webhookAPI := &api.WebhookAPI{Client: client}
		if err := webhookAPI.Set(args[0]); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Webhook endpoint set to %s\n", args[0])
		return nil
	},
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test webhook endpoint",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		webhookAPI := &api.WebhookAPI{Client: client}
		resp, err := webhookAPI.Test()
		if err != nil {
			return err
		}

		rows := []model.WebhookTestRow{{
			Success:    resp.Success,
			StatusCode: resp.StatusCode,
			Reason:     resp.Reason,
		}}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			return output.New(format).Format(os.Stdout, resp)
		}
		return output.New(format).Format(os.Stdout, rows)
	},
}
