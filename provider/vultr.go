package provider

import (
	"context"
	_ "embed"
	b64 "encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/mbund/nomadic-vpn/db"
	"github.com/vultr/govultr/v3"
	"golang.org/x/oauth2"
)

func initializeVultr() {
	apiKey, err := db.GetVultrAPIKey()
	if err != nil {
		return
	}

	config := &oauth2.Config{}
	ctx := context.Background()
	ts := config.TokenSource(ctx, &oauth2.Token{AccessToken: apiKey})
	vultrClient := govultr.NewClient(oauth2.NewClient(ctx, ts))

	plans, _, _, err := vultrClient.Plan.List(ctx, "", nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var cheapest_plan = plans[0]

	for _, plan := range plans {
		if plan.MonthlyCost < cheapest_plan.MonthlyCost {
			cheapest_plan = plan
		}
	}

	for _, locationCode := range cheapest_plan.Locations {
		l, exists := Locations[locationCode]
		if exists {
			l.Plans["vultr"] = Plan{
				Price: cheapest_plan.MonthlyCost,
				Provider: VultrProvider{
					PlanId:       cheapest_plan.ID,
					LocationCode: locationCode,
				},
			}
		}
	}
}

type VultrProvider struct {
	PlanId       string
	LocationCode string
}

func (v VultrProvider) CreateInstance(cloudConfig string) (string, error) {
	apiKey, err := db.GetVultrAPIKey()
	if err != nil {
		return "", fmt.Errorf("failed to get Vultr API key: %v", err)
	}

	config := &oauth2.Config{}
	ctx := context.Background()
	ts := config.TokenSource(ctx, &oauth2.Token{AccessToken: apiKey})
	vultrClient := govultr.NewClient(oauth2.NewClient(ctx, ts))

	instance, _, err := vultrClient.Instance.Create(context.Background(), &govultr.InstanceCreateReq{
		Label:    "nomadic-vpn",
		Backups:  "disabled",
		OsID:     2136, // Debian 12 x64 (bookworm)
		Plan:     v.PlanId,
		Region:   v.LocationCode,
		UserData: b64.StdEncoding.EncodeToString([]byte(cloudConfig)),
		Tags:     []string{"nomadic-vpn"},
	})

	if err != nil {
		return "", fmt.Errorf("failed to create instance: %v", err)
	}

	instanceId := instance.ID
	fmt.Println("Creating instance")

	for instance == nil || instance.Status != "active" {
		time.Sleep(30 * time.Second)
		instance, _, _ = vultrClient.Instance.Get(context.Background(), instanceId)
	}

	return instance.MainIP, nil
}

func (v VultrProvider) DestroyInstance(ip string) error {
	apiKey, err := db.GetVultrAPIKey()
	if err != nil {
		return nil
	}

	config := &oauth2.Config{}
	ctx := context.Background()
	ts := config.TokenSource(ctx, &oauth2.Token{AccessToken: apiKey})
	vultrClient := govultr.NewClient(oauth2.NewClient(ctx, ts))

	instances, _, _, err := vultrClient.Instance.List(context.Background(), &govultr.ListOptions{
		MainIP: ip,
	})
	if err != nil {
		return fmt.Errorf("failed to list instances: %v", err)
	}

	if len(instances) > 0 {
		instance := instances[0]
		err := vultrClient.Instance.Delete(context.Background(), instance.ID)
		if err != nil {
			return fmt.Errorf("failed to delete instance: %v", err)
		}
	}

	return nil
}
