package richmenu

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/cmd/cmdutil"
	"github.com/crowdy/lm-cli/internal/api"
	lmerrors "github.com/crowdy/lm-cli/internal/errors"
	"github.com/crowdy/lm-cli/internal/model"
	"github.com/crowdy/lm-cli/internal/output"
)

// Cmd is the richmenu command group.
var Cmd = &cobra.Command{
	Use:   "richmenu",
	Short: "Rich menu management",
}

var defaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Manage default rich menu",
}

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage rich menu aliases",
}

func init() {
	createCmd.Flags().String("file", "", "JSON file with rich menu definition (required)")
	_ = createCmd.MarkFlagRequired("file")

	aliasCreateCmd.Flags().String("file", "", "JSON file with alias definition (required)")
	_ = aliasCreateCmd.MarkFlagRequired("file")

	defaultCmd.AddCommand(defaultGetCmd)
	defaultCmd.AddCommand(defaultSetCmd)
	defaultCmd.AddCommand(defaultUnsetCmd)

	aliasCmd.AddCommand(aliasCreateCmd)
	aliasCmd.AddCommand(aliasListCmd)

	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(uploadCmd)
	Cmd.AddCommand(defaultCmd)
	Cmd.AddCommand(aliasCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a rich menu",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		file, _ := cmd.Flags().GetString("file")
		var menu model.RichMenu
		if err := api.ParseJSONFile(file, &menu); err != nil {
			return err
		}

		rmAPI := &api.RichMenuAPI{Client: client}
		resp, err := rmAPI.Create(menu)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Created rich menu: %s\n", resp.RichMenuID)
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get <richMenuId>",
	Short: "Get a rich menu",
	Args:  cmdutil.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		rmAPI := &api.RichMenuAPI{Client: client}
		menu, err := rmAPI.Get(args[0])
		if err != nil {
			return err
		}

		format := cmdutil.GetFormat(cmd)
		return output.New(format).Format(os.Stdout, menu)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all rich menus",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		rmAPI := &api.RichMenuAPI{Client: client}
		menus, err := rmAPI.List()
		if err != nil {
			return err
		}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			return output.New(format).Format(os.Stdout, menus)
		}

		rows := make([]model.RichMenuRow, len(menus))
		for i, m := range menus {
			rows[i] = model.RichMenuRow{
				RichMenuID:  m.RichMenuID,
				Name:        m.Name,
				ChatBarText: m.ChatBarText,
				Selected:    m.Selected,
			}
		}
		return output.New(format).Format(os.Stdout, rows)
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <richMenuId>",
	Short: "Delete a rich menu",
	Args:  cmdutil.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		rmAPI := &api.RichMenuAPI{Client: client}
		if err := rmAPI.Delete(args[0]); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Deleted rich menu %s\n", args[0])
		return nil
	},
}

var uploadCmd = &cobra.Command{
	Use:   "upload <richMenuId> <image.(jpg|png)>",
	Short: "Upload image for a rich menu",
	Args:  cmdutil.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		rmAPI := &api.RichMenuAPI{Client: client}
		if err := rmAPI.UploadImage(args[0], args[1]); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Uploaded image to rich menu %s\n", args[0])
		return nil
	},
}

var defaultGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get default rich menu ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		rmAPI := &api.RichMenuAPI{Client: client}
		id, err := rmAPI.GetDefault()
		if err != nil {
			return err
		}

		if id == "" {
			fmt.Println("(no default rich menu set)")
		} else {
			fmt.Println(id)
		}
		return nil
	},
}

var defaultSetCmd = &cobra.Command{
	Use:   "set <richMenuId>",
	Short: "Set default rich menu",
	Args:  cmdutil.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		rmAPI := &api.RichMenuAPI{Client: client}
		if err := rmAPI.SetDefault(args[0]); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Set default rich menu to %s\n", args[0])
		return nil
	},
}

var defaultUnsetCmd = &cobra.Command{
	Use:   "unset",
	Short: "Unset default rich menu",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		rmAPI := &api.RichMenuAPI{Client: client}
		if err := rmAPI.UnsetDefault(); err != nil {
			return err
		}

		fmt.Fprintln(os.Stderr, "Unset default rich menu")
		return nil
	},
}

var aliasCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a rich menu alias",
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

		if file == "" {
			return &lmerrors.ValidationError{Field: "file", Message: "required"}
		}

		rmAPI := &api.RichMenuAPI{Client: client}
		alias, err := rmAPI.CreateAlias(body)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Created alias: %s\n", alias.RichMenuAliasID)
		return nil
	},
}

var aliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all rich menu aliases",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		rmAPI := &api.RichMenuAPI{Client: client}
		aliases, err := rmAPI.ListAliases()
		if err != nil {
			return err
		}

		format := cmdutil.GetFormat(cmd)
		if format == "json" || format == "yaml" {
			return output.New(format).Format(os.Stdout, aliases)
		}

		rows := make([]model.RichMenuAliasRow, len(aliases))
		for i, a := range aliases {
			rows[i] = model.RichMenuAliasRow{
				RichMenuAliasID: a.RichMenuAliasID,
				RichMenuID:      a.RichMenuID,
			}
		}
		return output.New(format).Format(os.Stdout, rows)
	},
}
