package core

import (
	"bytes"
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/mbund/nomadic-vpn/db"
	"github.com/spf13/viper"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func GenerateWireguardClient(allowedIPs string) db.Client {
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

func Bootstrap(host string) (clientConfig string, err error) {
	key, _ := wgtypes.GeneratePrivateKey()

	server := db.Server{
		Address:    "10.0.0.1/24",
		PrivateKey: key.String(),
		PublicKey:  key.PublicKey().String(),
	}

	client := GenerateWireguardClient("10.0.0.2/32")

	var clientConfigBuf bytes.Buffer
	writeWireguardClientConfig(&clientConfigBuf, client, fmt.Sprintf("%v:51820", host), server.PublicKey)

	var serverConfigBuf bytes.Buffer
	writeWireguardServerConfig(&serverConfigBuf, server, []db.Client{client})
	e := os.WriteFile("/etc/wireguard/wg0.conf", serverConfigBuf.Bytes(), 0644)
	if e != nil {
		return "", fmt.Errorf("failed to write wireguard server config: %v", e)
	}

	db.AddClient(client)
	db.SetServer(server)

	return clientConfigBuf.String(), nil
}

func Connect(baseUrl string, accessToken string) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // allow self signed certs
	}
	client := &http.Client{Transport: tr}

	for {
		result, err := client.Get(fmt.Sprintf("%v/healthz", baseUrl))
		fmt.Println(result)
		fmt.Println(err)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if result.StatusCode == 200 {
			break
		}
	}

	fmt.Println("Server is up, initializing DB")

	formData := url.Values{
		"vultrApiKey":   {viper.GetString("VULTR_API_KEY")},
		"duckDnsToken":  {viper.GetString("DUCKDNS_TOKEN")},
		"duckDnsDomain": {viper.GetString("DUCKDNS_DOMAIN")},
	}

	req, _ := http.NewRequest("POST", fmt.Sprintf("%v/api/initialize", baseUrl), bytes.NewBufferString(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to initialize server: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %v", res.Status)
	}

	defer res.Body.Close()
	var apiResponse db.ApiInitializeResponse
	if err := json.NewDecoder(res.Body).Decode(&apiResponse); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return apiResponse.WireguardConf, nil
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
