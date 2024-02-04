package core

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"
	"time"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/mbund/nomadic-vpn/db"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func scpWriteData(sshClient *ssh.Client, data []byte, remotePath string) error {
	reader := bytes.NewReader(data)
	scpClient, err := scp.NewClientBySSH(sshClient)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer scpClient.Close()

	err = scpClient.CopyFile(context.Background(), reader, remotePath, "0644")
	if err != nil {
		return fmt.Errorf("failed to copy service file to server: %v", err)
	}

	return nil
}

func generateClient(allowedIPs string) db.Client {
	key, _ := wgtypes.GenerateKey()
	presharedKey, _ := wgtypes.GenerateKey()

	return db.Client{
		PrivateKey:   key.String(),
		PublicKey:    key.PublicKey().String(),
		PresharedKey: presharedKey.String(),
		AllowedIPs:   allowedIPs,
	}
}

//go:embed wg.conf
var wireguardServerConfigTemplate string

func writeWireguardServerConfig(writer io.Writer, server db.Server, clients []db.Client) {
	t, _ := template.New("wg.conf").Parse(wireguardServerConfigTemplate)

	t.Execute(writer, map[string]interface{}{
		"server":  server,
		"clients": clients,
	})
}

//go:embed client.conf
var wireguardClientConfigTemplate string

func writeWireguardClientConfig(writer io.Writer, client db.Client, endpoint string, serverPublicKey string) {
	t, _ := template.New("client.conf").Parse(wireguardClientConfigTemplate)

	t.Execute(writer, map[string]interface{}{
		"client":          client,
		"endpoint":        endpoint,
		"serverPublicKey": serverPublicKey,
	})
}

func Bootstrap(sshClient *ssh.Client) error {
	key, _ := wgtypes.GeneratePrivateKey()

	server := db.Server{
		Address:    "10.0.0.1/24",
		PrivateKey: key.String(),
		PublicKey:  key.PublicKey().String(),
	}

	client := generateClient("10.0.0.2/32")

	db.AddClient(client)
	db.SetServer(server)

	writeWireguardClientConfig(os.Stdout, client, "45.77.14.43:51820", server.PublicKey)

	err := runCommand(sshClient, "mkdir -p /etc/wireguard")
	if err != nil {
		return fmt.Errorf("failed to create wireguard directory: %v", err)
	}

	var configBuf bytes.Buffer
	writeWireguardServerConfig(&configBuf, server, []db.Client{client})
	err = scpWriteData(sshClient, configBuf.Bytes(), "/etc/wireguard/wg0.conf")
	if err != nil {
		return fmt.Errorf("failed to write wireguard server config: %v", err)
	}

	// run commands
	commands := []string{
		"wget https://github.com/mbund/nomadic-vpn/releases/latest/download/nomadic-vpn-linux-amd64 -O /usr/local/sbin/nomadic-vpn",
		"apt install -y wireguard",
		"systemctl enable nomadic-vpn.service wireguard-config.path wireguard-config.service",
		"systemctl start nomadic-vpn.service wireguard-config.path wireguard-config.service",
		"ufw disable",
	}

	for _, command := range commands {
		err := runCommand(sshClient, command)
		if err != nil {
			return err
		}
	}

	return nil
}

func runCommand(sshClient *ssh.Client, command string) error {
	session, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	fmt.Printf("Running command `%v`\n", command)
	_, err = session.CombinedOutput(command)
	if err != nil {
		return fmt.Errorf("failed to run command %v: %v", command, err)
	}

	return nil
}

func Connect(host string, accessToken string) error {
	for {
		result, err := http.Get(fmt.Sprintf("https://%v/healthz", host))
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if result.StatusCode == 200 {
			break
		}
	}

	fmt.Println("Connected to server")

	return nil
}

func UpdateDuckDNS(ip string) (string, error) {
	domain := viper.GetString("DUCKDNS_DOMAIN")
	token := viper.GetString("DUCKDNS_TOKEN")
	_, err := http.Get(fmt.Sprintf("https://www.duckdns.org/update?domains=%v&token=%v&ip=%v", domain, token, ip))
	if err != nil {
		return "", fmt.Errorf("failed to update duckdns: %v", err)
	}
	return fmt.Sprintf("%v.duckdns.org", domain), nil
}

func GenerateAccessToken() string {
	// this key does not have to be a wireguard style key, but should be cryptographically secure
	key, _ := wgtypes.GenerateKey()
	accessToken := key.String()
	return accessToken
}

//go:embed cloud-config.yaml
var cloudConfigTemplate string

func GenerateCloudConfig(accessToken string) string {
	domain := viper.GetString("DUCKDNS_DOMAIN")
	t, _ := template.New("cloud-config").Parse(cloudConfigTemplate)
	var cloudConfig bytes.Buffer
	t.Execute(&cloudConfig, map[string]interface{}{
		"Token":  accessToken,
		"Domain": fmt.Sprintf("%v.duckdns.org", domain),
	})
	return cloudConfig.String()
}
