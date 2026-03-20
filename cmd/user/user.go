package user

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/cmd/cmdutil"
	"github.com/crowdy/lm-cli/internal/api"
	"github.com/crowdy/lm-cli/internal/model"
	"github.com/crowdy/lm-cli/internal/output"
)

// Cmd is the user command group.
var Cmd = &cobra.Command{
	Use:   "user",
	Short: "User profile operations",
}

func init() {
	followersCmd.Flags().Int("limit", 0, "maximum number of follower IDs to return")
	followersCmd.Flags().String("start", "", "continuation token for pagination")
	Cmd.AddCommand(profileCmd)
	Cmd.AddCommand(followersCmd)
}

var profileCmd = &cobra.Command{
	Use:   "profile <userId>",
	Short: "Get user profile",
	Args:  cmdutil.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		userAPI := &api.UserAPI{Client: client}
		profile, err := userAPI.GetProfile(args[0])
		if err != nil {
			return err
		}

		rows := []model.UserProfileRow{{
			UserID:        profile.UserID,
			DisplayName:   profile.DisplayName,
			Language:      profile.Language,
			StatusMessage: profile.StatusMessage,
		}}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			return output.New(format).Format(os.Stdout, profile)
		}
		return output.New(format).Format(os.Stdout, rows)
	},
}

var followersCmd = &cobra.Command{
	Use:   "followers",
	Short: "Get follower user IDs",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		limit, _ := cmd.Flags().GetInt("limit")
		start, _ := cmd.Flags().GetString("start")

		userAPI := &api.UserAPI{Client: client}
		resp, err := userAPI.GetFollowers(limit, start)
		if err != nil {
			return err
		}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			return output.New(format).Format(os.Stdout, resp)
		}
		for _, id := range resp.UserIDs {
			fmt.Fprintln(os.Stdout, id)
		}
		if resp.Next != "" {
			fmt.Fprintf(os.Stderr, "(next: %s)\n", resp.Next)
		}
		return nil
	},
}
