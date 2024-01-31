package core

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/mbund/nomadic-vpn/db"
	"golang.org/x/crypto/ssh"
)

//go:embed nomadic-vpn.service
var service []byte

func Bootstrap(sshClient *ssh.Client) error {
	_, _ = db.DB.Exec("UPDATE keys SET vultr = ? WHERE id=0", "thanos")

	// copy self to server
	fmt.Println("Copying self to server")
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get path to self executable: %v", err)
	}

	selfFile, err := os.Open(self)
	if err != nil {
		return fmt.Errorf("failed to open self executable file: %v", err)
	}
	defer selfFile.Close()

	scpClient, err := scp.NewClientBySSH(sshClient)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	err = scpClient.CopyFromFile(context.Background(), *selfFile, "/usr/local/sbin/nomadic-vpn", "0755")
	if err != nil {
		return fmt.Errorf("failed to copy file %v to server: %v", selfFile, err)
	}
	scpClient.Close()

	// copy service file to server
	fmt.Println("Copying systemd service file to server")
	serviceReader := bytes.NewReader(service)
	scpClient, err = scp.NewClientBySSH(sshClient)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	err = scpClient.CopyFile(context.Background(), serviceReader, "/etc/systemd/system/nomadic-vpn.service", "0644")
	if err != nil {
		return fmt.Errorf("failed to copy service file to server: %v", err)
	}
	scpClient.Close()

	// copy db to server
	fmt.Println("Copying db to server")
	dbFile, err := os.Open("nomadic-vpn.db")
	if err != nil {
		return fmt.Errorf("failed to open db file: %v", err)
	}

	scpClient, err = scp.NewClientBySSH(sshClient)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	err = scpClient.CopyFromFile(context.Background(), *dbFile, "nomadic-vpn.db", "0755")
	if err != nil {
		return fmt.Errorf("failed to copy file %v to server: %v", dbFile, err)
	}
	scpClient.Close()

	os.Remove("nomadic-vpn.db")

	// run commands

	commands := []string{
		"systemctl enable nomadic-vpn",
		"systemctl start nomadic-vpn",
	}

	for _, command := range commands {
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
	}

	return nil
}
