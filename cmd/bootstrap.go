package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Create a new VPN server",
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)

	bootstrapCmd.PersistentFlags().String("duckdns-token", "", "DuckDNS token")
	viper.BindPFlag("DUCKDNS_TOKEN", bootstrapCmd.PersistentFlags().Lookup("duckdns-token"))

	bootstrapCmd.PersistentFlags().String("duckdns-domain", "", "DuckDNS domain")
	viper.BindPFlag("DUCKDNS_DOMAIN", bootstrapCmd.PersistentFlags().Lookup("duckdns-domain"))
}
