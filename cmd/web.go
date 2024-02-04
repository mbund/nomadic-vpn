package cmd

import (
	"github.com/mbund/nomadic-vpn/web"
	"github.com/spf13/cobra"
)

var port uint16
var accessToken string
var domain string

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Launch the web configuration interface",
	Long:  `A`,
	Run: func(cmd *cobra.Command, args []string) {
		web.Run(port, domain, accessToken)
	},
}

func init() {
	rootCmd.AddCommand(webCmd)

	webCmd.Flags().Uint16VarP(&port, "port", "p", 443, "Port to listen on")

	webCmd.Flags().StringVarP(&accessToken, "token", "t", "", "Access token for the web interface and api")
	webCmd.MarkFlagRequired("token")

	webCmd.Flags().StringVarP(&domain, "domain", "d", "", "Domain to listen on")
}
