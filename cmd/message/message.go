package message

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/cmd/cmdutil"
	"github.com/crowdy/lm-cli/internal/api"
	"github.com/crowdy/lm-cli/internal/config"
	lmerrors "github.com/crowdy/lm-cli/internal/errors"
	"github.com/crowdy/lm-cli/internal/model"
	"github.com/crowdy/lm-cli/internal/prompt"
)

// Cmd is the message command group.
var Cmd = &cobra.Command{
	Use:   "message",
	Short: "Send messages",
}

func init() {
	pushCmd.Flags().String("type", "text", "message type: text, sticker, image, flex")
	pushCmd.Flags().String("file", "", "JSON file with message payload")
	pushCmd.Flags().String("alt-text", "", "alt text for flex messages (default: \"Flex Message\")")

	multicastCmd.Flags().StringSlice("to", nil, "user IDs (comma-separated)")
	multicastCmd.Flags().String("to-file", "", "file with one user ID per line")
	multicastCmd.Flags().String("file", "", "JSON file with message payload")

	broadcastCmd.Flags().String("file", "", "JSON file with message payload")
	broadcastCmd.Flags().Bool("force", false, "skip confirmation prompt")

	narrowcastCmd.Flags().String("filter-file", "", "JSON file with narrowcast filter")
	narrowcastCmd.Flags().String("file", "", "JSON file with message payload")
	narrowcastCmd.Flags().Bool("force", false, "skip confirmation prompt")

	Cmd.AddCommand(pushCmd)
	Cmd.AddCommand(multicastCmd)
	Cmd.AddCommand(broadcastCmd)
	Cmd.AddCommand(narrowcastCmd)
	Cmd.AddCommand(replyCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push <userId> <text>",
	Short: "Push a message to a user",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		userID := args[0]
		messages, err := buildMessages(cmd, args)
		if err != nil {
			return err
		}

		msgAPI := &api.MessageAPI{Client: client}
		resp, err := msgAPI.Push(userID, messages)
		if err != nil {
			return err
		}

		if !cmdutil.IsQuiet(cmd) {
			fmt.Fprintf(os.Stderr, "Pushed %d message(s) to %s\n", len(resp.SentMessages), userID)
		}
		return nil
	},
}

var multicastCmd = &cobra.Command{
	Use:   "multicast <text>",
	Short: "Multicast a message to multiple users",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		toFlag, _ := cmd.Flags().GetStringSlice("to")
		toFile, _ := cmd.Flags().GetString("to-file")

		var userIDs []string
		userIDs = append(userIDs, toFlag...)

		if toFile != "" {
			ids, err := readLines(toFile)
			if err != nil {
				return err
			}
			userIDs = append(userIDs, ids...)
		}

		if len(userIDs) == 0 {
			return &lmerrors.ValidationError{Field: "to", Message: "at least one user ID required (use --to or --to-file)"}
		}

		messages, err := buildMessages(cmd, args)
		if err != nil {
			return err
		}

		msgAPI := &api.MessageAPI{Client: client}
		resp, err := msgAPI.MulticastBatch(userIDs, messages, func(batch, total int) {
			if !cmdutil.IsQuiet(cmd) {
				fmt.Fprintf(os.Stderr, "Sending batch %d/%d (%d users)...\n", batch, total, min(500, len(userIDs)-(batch-1)*500))
			}
		})
		if err != nil {
			return err
		}

		if !cmdutil.IsQuiet(cmd) {
			fmt.Fprintf(os.Stderr, "Multicasted %d message(s) to %d users\n", len(resp.SentMessages), len(userIDs))
		}
		return nil
	},
}

var broadcastCmd = &cobra.Command{
	Use:   "broadcast <text>",
	Short: "Broadcast a message to all followers",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		forceFlag, _ := cmd.Flags().GetBool("force")
		if err := confirmDangerous(cmd, forceFlag, "Broadcast to ALL followers. This cannot be undone.\nProceed?"); err != nil {
			return err
		}

		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		messages, err := buildMessages(cmd, args)
		if err != nil {
			return err
		}

		msgAPI := &api.MessageAPI{Client: client}
		resp, err := msgAPI.Broadcast(messages)
		if err != nil {
			return err
		}

		if !cmdutil.IsQuiet(cmd) {
			fmt.Fprintf(os.Stderr, "Broadcasted %d message(s)\n", len(resp.SentMessages))
		}
		return nil
	},
}

var narrowcastCmd = &cobra.Command{
	Use:   "narrowcast <text>",
	Short: "Narrowcast a message",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		forceFlag, _ := cmd.Flags().GetBool("force")
		if err := confirmDangerous(cmd, forceFlag, "Narrowcast a message to targeted followers. This cannot be undone.\nProceed?"); err != nil {
			return err
		}

		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		messages, err := buildMessages(cmd, args)
		if err != nil {
			return err
		}

		req := model.NarrowcastRequest{Messages: messages}

		filterFile, _ := cmd.Flags().GetString("filter-file")
		if filterFile != "" {
			var filter any
			if err := api.ParseJSONFile(filterFile, &filter); err != nil {
				return err
			}
			req.Filter = filter
		}

		msgAPI := &api.MessageAPI{Client: client}
		resp, err := msgAPI.Narrowcast(req)
		if err != nil {
			return err
		}

		if !cmdutil.IsQuiet(cmd) {
			fmt.Fprintf(os.Stderr, "Narrowcasted %d message(s)\n", len(resp.SentMessages))
		}
		return nil
	},
}

var replyCmd = &cobra.Command{
	Use:   "reply <replyToken> <text>",
	Short: "Reply to a webhook event",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cmdutil.NewClient(cmd)
		if err != nil {
			return err
		}

		replyToken := args[0]
		replyArgs := args[1:]
		messages, err := buildMessages(cmd, replyArgs)
		if err != nil {
			return err
		}

		if !cmdutil.IsQuiet(cmd) {
			fmt.Fprintln(os.Stderr, "Note: reply token is valid for 30 seconds from the webhook event.")
		}

		msgAPI := &api.MessageAPI{Client: client}
		resp, err := msgAPI.Reply(replyToken, messages)
		if err != nil {
			return err
		}

		if !cmdutil.IsQuiet(cmd) {
			fmt.Fprintf(os.Stderr, "Replied with %d message(s)\n", len(resp.SentMessages))
		}
		return nil
	},
}

// buildMessages constructs a message slice from command args/flags.
func buildMessages(cmd *cobra.Command, args []string) ([]any, error) {
	msgType, _ := cmd.Flags().GetString("type")
	if msgType == "" {
		msgType = "text"
	}

	// --file overrides everything (except for flex which uses --file differently)
	if fileFlag, _ := cmd.Flags().GetString("file"); fileFlag != "" && msgType != "flex" {
		var msgs []any
		if err := api.ParseJSONFile(fileFlag, &msgs); err != nil {
			// Try as a single message
			var msg any
			if err2 := api.ParseJSONFile(fileFlag, &msg); err2 != nil {
				return nil, err
			}
			return []any{msg}, nil
		}
		return msgs, nil
	}

	switch msgType {
	case "text":
		if len(args) == 0 {
			return nil, &lmerrors.ValidationError{Field: "text", Message: "text argument required"}
		}
		return []any{model.NewTextMessage(args[0])}, nil

	case "sticker":
		// Expect: <packageId> <stickerId>
		if len(args) < 2 {
			return nil, &lmerrors.ValidationError{
				Field:   "args",
				Message: "sticker requires <packageId> <stickerId>",
			}
		}
		return []any{model.StickerMessage{Type: "sticker", PackageID: args[0], StickerID: args[1]}}, nil

	case "image":
		if len(args) < 2 {
			return nil, &lmerrors.ValidationError{
				Field:   "args",
				Message: "image requires <originalContentUrl> <previewImageUrl>",
			}
		}
		return []any{model.ImageMessage{Type: "image", OriginalContentURL: args[0], PreviewImageURL: args[1]}}, nil

	case "flex":
		fileFlag, _ := cmd.Flags().GetString("file")
		if fileFlag == "" {
			return nil, &lmerrors.ValidationError{Field: "file", Message: "--file is required for --type flex"}
		}
		var contents any
		if err := api.ParseJSONFile(fileFlag, &contents); err != nil {
			return nil, err
		}
		altText, _ := cmd.Flags().GetString("alt-text")
		if altText == "" {
			altText = "Flex Message"
		}
		return []any{map[string]any{
			"type":     "flex",
			"altText":  altText,
			"contents": contents,
		}}, nil

	default:
		return nil, &lmerrors.ValidationError{Field: "type", Message: fmt.Sprintf("unknown type %q", msgType)}
	}
}

func confirmDangerous(cmd *cobra.Command, forceFlag bool, message string) error {
	if config.IsNoInput() || forceFlag {
		return nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return &lmerrors.ValidationError{
			Message: "confirmation required; use --force or LM_NO_INPUT=1 to bypass",
		}
	}
	confirmed, err := prompt.Confirm(message)
	if err != nil || !confirmed {
		return &lmerrors.CancelledError{}
	}
	return nil
}

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}
