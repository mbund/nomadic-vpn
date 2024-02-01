package cmd

import (
	"fmt"

	"github.com/mbund/nomadic-vpn/provider"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var bootstrapCloudCmd = &cobra.Command{
	Use:   "cloud <instance>",
	Short: "Create a new VPN server on a cloud provider",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}

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
	bootstrapCmd.AddCommand(bootstrapCloudCmd)

	bootstrapCloudCmd.Flags().String("vultr-api-key", "", "Vultr API key")
	viper.BindPFlag("VULTR_API_KEY", bootstrapCloudCmd.Flags().Lookup("vultr-api-key"))
}
