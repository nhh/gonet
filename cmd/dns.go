package cmd

import (
	"github.com/spf13/cobra"
	"gonet/internal/nameserver"
)

func init() {
	rootCmd.AddCommand(dnsCmd)
}

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "spunup networks with gonet",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		nameserver.Cache(name)
	},
}
