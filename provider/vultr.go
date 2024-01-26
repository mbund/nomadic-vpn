package provider

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vultr/govultr/v3"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
)

func initializeVultr() {
	apiKey := os.Getenv("VULTR_API_KEY")

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

func (v VultrProvider) Bootstrap() error {
	apiKey := os.Getenv("VULTR_API_KEY")

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

	instance, _, err := vultrClient.Instance.Create(context.Background(), &govultr.InstanceCreateReq{
		Label:   "nomadic-vpn",
		Backups: "disabled",
		OsID:    2136, // Debian 12 x64 (bookworm)
		Plan:    v.PlanId,
		Region:  v.LocationCode,
		SSHKeys: []string{vultrSshKey.ID},
	})

	if err != nil {
		fmt.Println(err)
		panic("Failed to create instance")
	}

	for instance.Status != "active" {
		instance, _, err = vultrClient.Instance.Get(context.Background(), instance.ID)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(5 * time.Second)
	}

	ip := instance.MainIP

	p, _ := ssh.MarshalPrivateKey(privateKey, "nomadic-vpn")
	privateKeyPem := pem.EncodeToMemory(p)
	// os.WriteFile("ssh_key", privateKeyPem, 0600)

	signer, _ := ssh.ParsePrivateKey(privateKeyPem)
	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", ip+":22", sshConfig)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}

	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput("ls /")
	if err != nil {
		log.Fatalf("Failed to run command: %v", err)
	}

	fmt.Println(string(output))

	return nil
}
