package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

var dB *sql.DB

func InitDb() {
	database, err := createDb("nomadic-vpn.db")
	if err != nil {
		panic(err)
	}
	dB = database
}

func createDb(dbPath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent directories for the database: %s", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %s", err)
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS config (
			id INTEGER PRIMARY KEY CHECK (id = 0),
			vultrapikey VARCHAR(64),
			duckdnstoken VARCHAR(64),
			duckdnsdomain VARCHAR(64)
		);`,
		`INSERT OR IGNORE INTO config (id, vultrapikey) VALUES (0, NULL);`,
		`CREATE TABLE IF NOT EXISTS client (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			privatekey VARCHAR(64),
			publickey VARCHAR(64),
			presharedkey VARCHAR(64),
			allowedips VARCHAR(256)
		);`,
		`CREATE TABLE IF NOT EXISTS server (
			id INTEGER PRIMARY KEY CHECK (id = 0),
			privatekey VARCHAR(64),
			publickey VARCHAR(64),
			address VARCHAR(64),
			listenport INTEGER
		);`,
	}

	for _, stmt := range stmts {
		_, err := db.Exec(stmt)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func SetVultrApiKey(vultrApiKey string) {
	if len(vultrApiKey) > 0 {
		_, _ = dB.Exec("UPDATE config SET vultrapikey = ? WHERE id = 0", vultrApiKey)
	}
}

func SetDuckDnsToken(duckDnsToken string) {
	if len(duckDnsToken) > 0 {
		_, _ = dB.Exec("UPDATE config SET duckdnstoken = ? WHERE id = 0", duckDnsToken)
	}
}

func SetDuckDnsDomain(duckDnsDomain string) {
	if len(duckDnsDomain) > 0 {
		_, _ = dB.Exec("UPDATE config SET duckdnsdomain = ? WHERE id = 0", duckDnsDomain)
	}
}

func GetVultrAPIKey() (string, error) {
	if dB == nil {
		return viper.GetString("VULTR_API_KEY"), nil
	}

	var vultrApiKey string
	dB.QueryRow("SELECT vultrapikey FROM config WHERE id = 0").Scan(&vultrApiKey)
	if len(vultrApiKey) > 0 {
		return vultrApiKey, nil
	}

	return "", fmt.Errorf("vultr api key not set")
}

type Server struct {
	PrivateKey string
	PublicKey  string
	Address    string
	ListenPort int
}

type Client struct {
	PrivateKey   string
	PublicKey    string
	PresharedKey string
	AllowedIPs   string
}

func GetClients() []Client {
	rows, err := dB.Query("SELECT privatekey, publickey, presharedkey, allowedips FROM client")
	if err != nil {
		return nil
	}
	defer rows.Close()

	clients := []Client{}
	for rows.Next() {
		var client Client
		err := rows.Scan(&client.PrivateKey, &client.PublicKey, &client.PresharedKey, &client.AllowedIPs)
		if err != nil {
			return nil
		}
		clients = append(clients, client)
	}

	return clients
}

func GetServer() (Server, error) {
	var server Server
	err := dB.QueryRow("SELECT privatekey, publickey, address, listenport FROM server WHERE id = 0").Scan(&server.PrivateKey, &server.PublicKey, &server.Address, &server.ListenPort)
	if err != nil {
		return Server{}, err
	}

	return server, nil
}

func SetServer(server Server) error {
	_, err := dB.Exec("UPDATE server SET privatekey = ?, publickey = ?, address = ?, listenport = ? WHERE id = 0", server.PrivateKey, server.PublicKey, server.Address, server.ListenPort)
	return err
}

func AddClient(client Client) error {
	_, err := dB.Exec("INSERT INTO client (privatekey, publickey, presharedkey, allowedips) VALUES (?, ?, ?, ?)", client.PrivateKey, client.PublicKey, client.PresharedKey, client.AllowedIPs)
	return err
}

type ApiInitializeResponse struct {
	WireguardConf string `json:"conf"`
}
