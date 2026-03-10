package proxmox

import (
	"crypto/tls"
	"net/http"

	"github.com/luthermonson/go-proxmox"
)

// Config holds the configuration for the Proxmox client.
type Config struct {
	Endpoint string
	TokenID  string
	Secret   string
}

// NewClient creates and configures a new Proxmox API client.
func NewClient(cfg Config) *proxmox.Client {
	insecureHTTPClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	client := proxmox.NewClient(cfg.Endpoint,
		proxmox.WithHTTPClient(&insecureHTTPClient),
		proxmox.WithAPIToken(cfg.TokenID, cfg.Secret),
	)

	return client
}