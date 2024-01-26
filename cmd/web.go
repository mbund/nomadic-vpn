package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var port uint16

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Launch the web configuration interface",
	Long:  `A`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("web called")
	},
}

func init() {
	rootCmd.AddCommand(webCmd)

	webCmd.Flags().Uint16VarP(&port, "port", "p", 8080, "Port to listen on")
}
