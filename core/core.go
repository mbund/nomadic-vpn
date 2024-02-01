package core

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"text/template"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/mbund/nomadic-vpn/db"
	"golang.org/x/crypto/ssh"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

//go:embed nomadic-vpn.service
var systemdNomadicVpnService []byte

//go:embed wireguard-config.path
var systemdWireguardConfigPath []byte

//go:embed wireguard-config.service
var systemdWireguardConfigService []byte

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

func scpCopyFile(sshClient *ssh.Client, filepath string, remotePath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

	scpClient, err := scp.NewClientBySSH(sshClient)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer scpClient.Close()

	err = scpClient.CopyFromFile(context.Background(), *file, remotePath, "0755")
	if err != nil {
		return fmt.Errorf("failed to copy file %v to server: %v", file, err)
	}

	return nil
}

// func wireguard() {
// 	key, _ := wgtypes.GeneratePrivateKey()
// 	presharedKey, _ := wgtypes.GenerateKey()
// }

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

	// copy self to server
	// fmt.Println("Copying self to server")
	// self, err := os.Executable()
	// if err != nil {
	// 	return fmt.Errorf("failed to get path to self executable: %v", err)
	// }
	// scpWriteFile(sshClient, self, "/usr/local/sbin/nomadic-vpn")

	scpWriteData(sshClient, systemdNomadicVpnService, "/etc/systemd/system/nomadic-vpn.service")
	scpWriteData(sshClient, systemdWireguardConfigPath, "/etc/systemd/system/wireguard-config.path")
	scpWriteData(sshClient, systemdWireguardConfigService, "/etc/systemd/system/wireguard-config.service")

	// copy db to server
	scpCopyFile(sshClient, "nomadic-vpn.db", "nomadic-vpn.db")
	os.Remove("nomadic-vpn.db")

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
