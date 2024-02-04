package provider

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	_ "embed"
	b64 "encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/mbund/nomadic-vpn/core"
	"github.com/mbund/nomadic-vpn/db"
	"github.com/vultr/govultr/v3"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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

//go:embed cloud-config.yaml
var cloudConfigTemplate string

type VultrProvider struct {
	PlanId       string
	LocationCode string
}

func (v VultrProvider) Bootstrap() error {
	apiKey, err := db.GetVultrAPIKey()
	if err != nil {
		return nil
	}

	config := &oauth2.Config{}
	ctx := context.Background()
	ts := config.TokenSource(ctx, &oauth2.Token{AccessToken: apiKey})
	vultrClient := govultr.NewClient(oauth2.NewClient(ctx, ts))

	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	sshPublicKey, _ := ssh.NewPublicKey(publicKey)
	sshPublicKeyString := string(ssh.MarshalAuthorizedKey(sshPublicKey))

	vultrSshKey, _, err := vultrClient.SSHKey.Create(context.Background(), &govultr.SSHKeyReq{
		Name:   "nomadic-vpn",
		SSHKey: sshPublicKeyString,
	})

	if err != nil {
		fmt.Println(err)
		panic("Failed to create SSH key")
	}

	// this key does not have to be a wireguard style key, but should be cryptographically secure
	key, _ := wgtypes.GenerateKey()
	accessToken := key.String()

	t, _ := template.New("cloud-config").Parse(cloudConfigTemplate)
	var cloudConfig bytes.Buffer
	t.Execute(&cloudConfig, map[string]interface{}{
		"Token": accessToken,
	})
	userData := b64.StdEncoding.EncodeToString(cloudConfig.Bytes())

	instance, _, err := vultrClient.Instance.Create(context.Background(), &govultr.InstanceCreateReq{
		Label:    "nomadic-vpn",
		Backups:  "disabled",
		OsID:     2136, // Debian 12 x64 (bookworm)
		Plan:     v.PlanId,
		Region:   v.LocationCode,
		UserData: userData,
		SSHKeys:  []string{vultrSshKey.ID},
		Tags:     []string{"nomadic-vpn"},
	})

	if err != nil {
		fmt.Println(err)
		panic("Failed to create instance")
	}

	instanceId := instance.ID
	fmt.Println("Creating instance")

	for instance == nil || instance.Status != "active" {
		time.Sleep(30 * time.Second)
		instance, _, _ = vultrClient.Instance.Get(context.Background(), instanceId)
		if err != nil {
			fmt.Println(err)
		}
	}

	ip := instance.MainIP
	fmt.Printf("Instance created with IP: %v\n", ip)

	p, _ := ssh.MarshalPrivateKey(privateKey, "nomadic-vpn")
	privateKeyPem := pem.EncodeToMemory(p)
	os.WriteFile("ssh_key", privateKeyPem, 0600)

	core.Connect(ip, accessToken)

	return nil
}
