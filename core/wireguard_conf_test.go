package core

import (
	"fmt"
	"os"
	"testing"

	"github.com/mbund/nomadic-vpn/db"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestWriteConfig(t *testing.T) {
	key, _ := wgtypes.GeneratePrivateKey()

	server := db.Server{
		Address:    "10.0.0.1/24",
		PrivateKey: key.String(),
		PublicKey:  key.PublicKey().String(),
	}

	clients := []db.Client{
		generateClient("10.0.0.2/32"),
		generateClient("10.0.0.3/32"),
	}

	writeWireguardServerConfig(os.Stdout, server, clients)
	fmt.Println("-------------------")
	writeWireguardClientConfig(os.Stdout, clients[0], "remote.example.com:51820", server.PublicKey)
	fmt.Println("-------------------")
	writeWireguardClientConfig(os.Stdout, clients[1], "remote.example.com:51820", server.PublicKey)
}
