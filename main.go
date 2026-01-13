package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/luthermonson/go-proxmox"
)

func main() {
	insecureHTTPClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	tokenID := "root@pam!mytoken"
	secret := "somegeneratedapitokenguidefromtheproxmoxui"

	client := proxmox.NewClient("https://localhost:8006/api2/json",
		proxmox.WithHTTPClient(&insecureHTTPClient),
		proxmox.WithAPIToken(tokenID, secret),
	)

	version, err := client.Version(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(version.Release) // 6.3
}
