package cmd

import (
	"github.com/spf13/cobra"
	"gonet/internal/wg"
	"os"
)

func init() {
	rootCmd.AddCommand(wgCmd)
}

var wgCmd = &cobra.Command{
	Use:   "wg",
	Short: "spunup networks with gonet",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(-1)
		}

		command := args[0]

		// gonet join blabla123

		if command == "join" {
			wg.Join()
		} else {
			wg.Run()
		}
	},
}
