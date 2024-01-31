/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/mbund/nomadic-vpn/core"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// manualCmd represents the test command
var manualCmd = &cobra.Command{
	Use:   "manual [ssh_key] [ip]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ssh_key_path := args[0]
		ip := args[1]

		bs, _ := os.ReadFile(ssh_key_path)
		signer, _ := ssh.ParsePrivateKey(bs)

		config := &ssh.ClientConfig{
			User: "root",
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		addr := fmt.Sprintf("%s:%s", ip, "22")
		sshClient, err := ssh.Dial("tcp", addr, config)
		if err != nil {
			panic(err)
		}
		defer sshClient.Close()

		err = core.Bootstrap(sshClient)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(manualCmd)
}
