package audience

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/cmd/cmdutil"
	"github.com/crowdy/lm-cli/internal/api"
	lmerrors "github.com/crowdy/lm-cli/internal/errors"
	"github.com/crowdy/lm-cli/internal/model"
	"github.com/crowdy/lm-cli/internal/output"
)

// Cmd is the audience command group.
var Cmd = &cobra.Command{
	Use:   "audience",
	Short: "Audience group management",
}

func init() {
	createCmd.Flags().String("file", "", "JSON file with audience group definition (required)")
	_ = createCmd.MarkFlagRequired("file")

	listCmd.Flags().Int("page", 0, "page number to retrieve")
	listCmd.Flags().Bool("all", false, "fetch all pages with auto-pagination")

	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(deleteCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an audience group",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		file, _ := cmd.Flags().GetString("file")
		var body any
		if err := api.ParseJSONFile(file, &body); err != nil {
			return err
		}

		audienceAPI := &api.AudienceAPI{Client: client}
		resp, err := audienceAPI.Create(body)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Created audience group: %d\n", resp.AudienceGroupID)
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get <audienceGroupId>",
	Short: "Get an audience group",
	Args:  cmdutil.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid audience group ID: %s", args[0])
		}

		audienceAPI := &api.AudienceAPI{Client: client}
		group, err := audienceAPI.Get(id)
		if err != nil {
			return err
		}

		rows := []model.AudienceGroupRow{{
			AudienceGroupID: group.AudienceGroupID,
			Type:            group.Type,
			Description:     group.Description,
			Status:          group.Status,
			AudienceCount:   group.AudienceCount,
		}}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			return output.New(format).Format(os.Stdout, group)
		}
		return output.New(format).Format(os.Stdout, rows)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all audience groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		allFlag, _ := cmd.Flags().GetBool("all")
		page, _ := cmd.Flags().GetInt("page")

		if allFlag && page > 0 {
			return &lmerrors.ValidationError{Field: "flags", Message: "--all and --page cannot be used together"}
		}

		audienceAPI := &api.AudienceAPI{Client: client}

		var allGroups []model.AudienceGroup

		if allFlag {
			const maxPages = 10000
			for p := 1; p <= maxPages; p++ {
				resp, err := audienceAPI.List(p)
				if err != nil {
					return err
				}
				allGroups = append(allGroups, resp.AudienceGroups...)
				if !resp.HasNextPage {
					break
				}
				if !cmdutil.IsQuiet(cmd) {
					fmt.Fprintf(os.Stderr, "Fetched %d audience groups...\r", len(allGroups))
				}
			}
			if !cmdutil.IsQuiet(cmd) {
				fmt.Fprintln(os.Stderr)
			}
		} else {
			resp, err := audienceAPI.List(page)
			if err != nil {
				return err
			}
			allGroups = resp.AudienceGroups
		}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			return output.New(format).Format(os.Stdout, allGroups)
		}

		rows := make([]model.AudienceGroupRow, len(allGroups))
		for i, g := range allGroups {
			rows[i] = model.AudienceGroupRow{
				AudienceGroupID: g.AudienceGroupID,
				Type:            g.Type,
				Description:     g.Description,
				Status:          g.Status,
				AudienceCount:   g.AudienceCount,
			}
		}
		return output.New(format).Format(os.Stdout, rows)
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <audienceGroupId>",
	Short: "Delete an audience group",
	Args:  cmdutil.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid audience group ID: %s", args[0])
		}

		audienceAPI := &api.AudienceAPI{Client: client}
		if err := audienceAPI.Delete(id); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Deleted audience group %d\n", id)
		return nil
	},
}
