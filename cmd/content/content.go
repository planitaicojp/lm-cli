package content

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/cmd/cmdutil"
	"github.com/crowdy/lm-cli/internal/api"
)

// Cmd is the content command group.
var Cmd = &cobra.Command{
	Use:   "content",
	Short: "Download message content",
}

func init() {
	getCmd.Flags().String("output", "", "output file path (default: stdout)")
	Cmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get <messageId>",
	Short: "Download binary content for a message",
	Args:  cmdutil.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		contentAPI := &api.ContentAPI{Client: client}
		body, err := contentAPI.Get(args[0])
		if err != nil {
			return err
		}
		defer body.Close()

		outputPath, _ := cmd.Flags().GetString("output")

		var dest io.Writer
		if outputPath != "" {
			f, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("creating output file: %w", err)
			}
			defer f.Close()
			dest = f
			fmt.Fprintf(os.Stderr, "Saving to %s...\n", outputPath)
		} else {
			dest = os.Stdout
		}

		n, err := io.Copy(dest, body)
		if err != nil {
			return fmt.Errorf("writing content: %w", err)
		}

		if outputPath != "" {
			fmt.Fprintf(os.Stderr, "Saved %d bytes to %s\n", n, outputPath)
		}
		return nil
	},
}
