package insight

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/cmd/cmdutil"
	"github.com/crowdy/lm-cli/internal/api"
	lmerrors "github.com/crowdy/lm-cli/internal/errors"
	"github.com/crowdy/lm-cli/internal/model"
	"github.com/crowdy/lm-cli/internal/output"
)

// Cmd is the insight command group.
var Cmd = &cobra.Command{
	Use:   "insight",
	Short: "Statistics and insights",
}

func init() {
	followersCmd.Flags().String("date", "", "date in YYYYMMDD format (default: yesterday)")
	deliveryCmd.Flags().String("date", "", "date in YYYYMMDD format (default: yesterday)")
	deliveryCmd.Flags().String("type", "broadcast", "delivery type: broadcast, push, multicast, narrowcast")

	Cmd.AddCommand(followersCmd)
	Cmd.AddCommand(deliveryCmd)
}

var followersCmd = &cobra.Command{
	Use:   "followers",
	Short: "Get follower statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		date, _ := cmd.Flags().GetString("date")
		if err := validateDate(date); err != nil {
			return err
		}

		insightAPI := &api.InsightAPI{Client: client}
		stats, err := insightAPI.GetFollowers(date)
		if err != nil {
			return err
		}

		rows := []model.FollowerStatsRow{{
			Status:    stats.Status,
			Followers: stats.Followers,
			Targeted:  stats.Targeted,
			Blocks:    stats.Blocks,
		}}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			return output.New(format).Format(os.Stdout, stats)
		}
		return output.New(format).Format(os.Stdout, rows)
	},
}

var deliveryCmd = &cobra.Command{
	Use:   "delivery",
	Short: "Get message delivery statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		date, _ := cmd.Flags().GetString("date")
		if err := validateDate(date); err != nil {
			return err
		}
		msgType, _ := cmd.Flags().GetString("type")

		insightAPI := &api.InsightAPI{Client: client}
		stats, err := insightAPI.GetDelivery(msgType, date)
		if err != nil {
			return err
		}

		rows := []model.DeliveryStatsRow{{
			Status:    stats.Status,
			Broadcast: stats.Broadcast,
			Targeting: stats.Targeting,
		}}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			return output.New(format).Format(os.Stdout, stats)
		}
		return output.New(format).Format(os.Stdout, rows)
	},
}

func validateDate(date string) error {
	if date == "" {
		return nil
	}
	if len(date) != 8 {
		return &lmerrors.ValidationError{
			Field:   "date",
			Message: fmt.Sprintf("must be YYYYMMDD format, got %q", date),
		}
	}
	if _, err := time.Parse("20060102", date); err != nil {
		return &lmerrors.ValidationError{
			Field:   "date",
			Message: fmt.Sprintf("invalid date %q (expected YYYYMMDD)", date),
		}
	}
	return nil
}
