package cmd

import (
	"github.com/spf13/cobra"
	"gonet/internal/signaling"
	"os"
)

func init() {
	rootCmd.AddCommand(mqttCmd)
}

var mqttCmd = &cobra.Command{
	Use:   "mqtt",
	Short: "spunup networks with gonet",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(-1)
		}

		command := args[0]
		session := args[1]

		// gonet join blabla123

		if session == "" {
			os.Exit(1)
		}

		if command == "join" {
			signaling.JoinRoom(session)
		} else {
			signaling.CreateRoom(session)
		}

	},
}
