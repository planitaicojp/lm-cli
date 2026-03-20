package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("lm version %s\n", version)
		fmt.Println("LINE Messaging API CLI by crowdy@gmail.com")
		fmt.Println("This is an unofficial tool and is not affiliated with or endorsed by LINE Corporation.")
	},
}
