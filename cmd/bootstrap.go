package cmd

import (
	"fmt"

	"github.com/mbund/nomadic-vpn/core"
	"github.com/mbund/nomadic-vpn/provider"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Create a new VPN server on a public cloud provider",
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

		accessToken := core.GenerateAccessToken()
		cloudConfig := core.GenerateCloudConfig(accessToken)

		ip, err := plan.Provider.CreateInstance(cloudConfig)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Created instance with IP %v\n", ip)

		host, err := core.UpdateDuckDNS(ip)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Waiting for server to come online")
		err = core.Connect(host, accessToken)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)

	bootstrapCmd.PersistentFlags().String("duckdns-token", "", "DuckDNS token")
	viper.BindPFlag("DUCKDNS_TOKEN", bootstrapCmd.PersistentFlags().Lookup("duckdns-token"))

	bootstrapCmd.PersistentFlags().String("duckdns-domain", "", "DuckDNS domain")
	viper.BindPFlag("DUCKDNS_DOMAIN", bootstrapCmd.PersistentFlags().Lookup("duckdns-domain"))

	bootstrapCmd.Flags().String("vultr-api-key", "", "Vultr API key")
	viper.BindPFlag("VULTR_API_KEY", bootstrapCmd.Flags().Lookup("vultr-api-key"))
}
