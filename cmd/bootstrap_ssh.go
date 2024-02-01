package cmd

import (
	"fmt"
	"os"

	"github.com/mbund/nomadic-vpn/core"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var identityFile string

var bootstrapSshCmd = &cobra.Command{
	Use:   "ssh [flags] <host>",
	Short: "Create a new VPN server on an existing host with SSH",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}

		host := args[0]

		identityFileBytes, _ := os.ReadFile(identityFile)
		signer, _ := ssh.ParsePrivateKey(identityFileBytes)

		config := &ssh.ClientConfig{
			User: "root",
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		addr := fmt.Sprintf("%s:%s", host, "22")
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
	bootstrapCmd.AddCommand(bootstrapSshCmd)

	bootstrapSshCmd.Flags().StringVarP(&identityFile, "identity-file", "i", "", "Path to ssh identity file")
	bootstrapSshCmd.MarkFlagRequired("identity-file")
}
