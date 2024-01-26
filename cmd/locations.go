package cmd

import (
	"fmt"

	"github.com/mbund/nomadic-vpn/provider"
	"github.com/spf13/cobra"
)

var locationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "List available locations",
	Long:  `A`,
	Run: func(cmd *cobra.Command, args []string) {
		provider.InitializeProviders()
		for locationCode, location := range provider.Locations {
			fmt.Print(location.City, ", ", location.Country, " ")
			for providerId, plan := range location.Plans {
				fmt.Print(locationCode, "-", providerId, "/$", plan.Price, " ")
			}
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(locationsCmd)
}
