package cmd

import (
	"fmt"

	"github.com/mbund/nomadic-vpn/provider"
	"github.com/spf13/cobra"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap [instance]",
	Short: "Create a new VPN server on a cloud provider",
	Long:  `A`,
	Run: func(cmd *cobra.Command, args []string) {
		provider.InitializeProviders()

		instanceId := args[0]

		if len(instanceId) <= 4 {
			fmt.Println("Invalid instance ID")
			return
		}

		locationId := instanceId[:3]
		providerId := instanceId[4:]

		location, locationExists := provider.Locations[locationId]
		if !locationExists {
			fmt.Println("Invalid location ID")
			return
		}

		plan, planExists := location.Plans[providerId]
		if !planExists {
			fmt.Println("Invalid provider ID")
			return
		}

		plan.Provider.Bootstrap()
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)
}
