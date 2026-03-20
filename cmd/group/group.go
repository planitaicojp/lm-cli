package group

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/cmd/cmdutil"
	"github.com/crowdy/lm-cli/internal/api"
	"github.com/crowdy/lm-cli/internal/model"
	"github.com/crowdy/lm-cli/internal/output"
)

// Cmd is the group command group.
var Cmd = &cobra.Command{
	Use:   "group",
	Short: "Group chat operations",
}

func init() {
	Cmd.AddCommand(infoCmd)
	Cmd.AddCommand(membersCmd)
	Cmd.AddCommand(leaveCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info <groupId>",
	Short: "Get group summary",
	Args:  cmdutil.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		groupAPI := &api.GroupAPI{Client: client}
		summary, err := groupAPI.GetSummary(args[0])
		if err != nil {
			return err
		}

		rows := []model.GroupSummaryRow{{
			GroupID:   summary.GroupID,
			GroupName: summary.GroupName,
		}}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			return output.New(format).Format(os.Stdout, summary)
		}
		return output.New(format).Format(os.Stdout, rows)
	},
}

var membersCmd = &cobra.Command{
	Use:   "members <groupId>",
	Short: "Get member user IDs in a group",
	Args:  cmdutil.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		groupAPI := &api.GroupAPI{Client: client}
		resp, err := groupAPI.GetMembers(args[0])
		if err != nil {
			return err
		}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			return output.New(format).Format(os.Stdout, resp)
		}
		for _, id := range resp.MemberIDs {
			fmt.Fprintln(os.Stdout, id)
		}
		if resp.Next != "" {
			fmt.Fprintf(os.Stderr, "(next: %s)\n", resp.Next)
		}
		return nil
	},
}

var leaveCmd = &cobra.Command{
	Use:   "leave <groupId>",
	Short: "Leave a group",
	Args:  cmdutil.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		groupAPI := &api.GroupAPI{Client: client}
		if err := groupAPI.Leave(args[0]); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Left group %s\n", args[0])
		return nil
	},
}
