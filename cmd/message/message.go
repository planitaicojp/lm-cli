package message

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/crowdy/lm-cli/cmd/cmdutil"
	"github.com/crowdy/lm-cli/internal/api"
	lmerrors "github.com/crowdy/lm-cli/internal/errors"
	"github.com/crowdy/lm-cli/internal/model"
)

// Cmd is the message command group.
var Cmd = &cobra.Command{
	Use:   "message",
	Short: "Send messages",
}

func init() {
	pushCmd.Flags().String("type", "text", "message type: text, sticker, image")
	pushCmd.Flags().String("file", "", "JSON file with message payload")

	multicastCmd.Flags().StringSlice("to", nil, "user IDs (comma-separated)")
	multicastCmd.Flags().String("to-file", "", "file with one user ID per line")
	multicastCmd.Flags().String("file", "", "JSON file with message payload")

	broadcastCmd.Flags().String("file", "", "JSON file with message payload")

	narrowcastCmd.Flags().String("filter-file", "", "JSON file with narrowcast filter")
	narrowcastCmd.Flags().String("file", "", "JSON file with message payload")

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

		if !isQuiet(cmd) {
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
		resp, err := msgAPI.Multicast(userIDs, messages)
		if err != nil {
			return err
		}

		if !isQuiet(cmd) {
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

		if !isQuiet(cmd) {
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

		if !isQuiet(cmd) {
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

		msgAPI := &api.MessageAPI{Client: client}
		resp, err := msgAPI.Reply(replyToken, messages)
		if err != nil {
			return err
		}

		if !isQuiet(cmd) {
			fmt.Fprintf(os.Stderr, "Replied with %d message(s)\n", len(resp.SentMessages))
		}
		return nil
	},
}

// buildMessages constructs a message slice from command args/flags.
func buildMessages(cmd *cobra.Command, args []string) ([]any, error) {
	// --file overrides everything
	if fileFlag, _ := cmd.Flags().GetString("file"); fileFlag != "" {
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

	msgType, _ := cmd.Flags().GetString("type")
	if msgType == "" {
		msgType = "text"
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

	default:
		return nil, &lmerrors.ValidationError{Field: "type", Message: fmt.Sprintf("unknown type %q", msgType)}
	}
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

func isQuiet(cmd *cobra.Command) bool {
	quiet, _ := cmd.Flags().GetBool("quiet")
	return quiet
}
