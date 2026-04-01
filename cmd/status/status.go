package status

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/cmd/cmdutil"
	"github.com/crowdy/lm-cli/internal/api"
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

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			data := map[string]string{
				"api":          "ok",
				"bot_id":       info.BasicID,
				"display_name": info.DisplayName,
			}
			return output.New(format).Format(os.Stdout, data)
		}

		fmt.Fprintf(os.Stdout, "API:     ok\n")
		fmt.Fprintf(os.Stdout, "Bot:     %s (%s)\n", info.BasicID, info.DisplayName)
		return nil
	},
}
