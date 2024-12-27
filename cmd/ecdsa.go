package cmd

import (
	"github.com/spf13/cobra"
	"gonet/internal/security"
	"gonet/internal/signaling"
)

func init() {
	rootCmd.AddCommand(ecdsaCmd)
}

var ecdsaCmd = &cobra.Command{
	Use:   "crypto",
	Short: "spunup networks with gonet",
	Run: func(cmd *cobra.Command, args []string) {
		pubkey := security.GenerateShortPublicKey()
		signaling.CreateRoom(pubkey)
	},
}
