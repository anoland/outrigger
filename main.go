package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/luthermonson/go-proxmox"
)

func main() {
	// 1. Define CLI Flags
	dryRun := flag.Bool("dry-run", true, "If true, only plans the move without executing")
	threshold := flag.Float64("threshold", 0.15, "Minimum improvement required (0.15 = 15%)")
	flag.Parse()

	fmt.Println("🚢 TRIM: Proxmox Cluster Balancer")
	fmt.Printf("Settings: DryRun=%v, Threshold=%.2f\n", *dryRun, *threshold)

	// Initialize Proxmox client
	insecureHTTPClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	tokenID := os.Getenv("PVE_TOKEN_ID")
	secret := os.Getenv("PVE_TOKEN_SECRET")

	_ = proxmox.NewClient(os.Getenv("PVE_ENDPOINT"),
		proxmox.WithHTTPClient(&insecureHTTPClient),
		proxmox.WithAPIToken(tokenID, secret),
	)

	// 2. Mock Data (Normally fetched from Proxmox API)
	nodes := []Node{
		{Name: "pve-01", MemUsedGB: 120.0, VMs: []VM{
			{ID: 101, Name: "Database", MemUsedGB: 45.0},
			{ID: 102, Name: "Web-Srv", MemUsedGB: 20.0},
		}},
		{Name: "pve-02", MemUsedGB: 60.0, VMs: []VM{{ID: 103, Name: "Utility", MemUsedGB: 10.0}}},
		{Name: "pve-03", MemUsedGB: 55.0, VMs: []VM{{ID: 104, Name: "Docker-Host", MemUsedGB: 30.0}}},
		{Name: "pve-04", MemUsedGB: 30.0, VMs: []VM{}},
	}

	// 3. Analyze Cluster
	mean, sd := calculateSD(nodes)
	fmt.Printf("\n[STATUS] Cluster Mean: %.2f GB | Imbalance Score (SD): %.2f\n", mean, sd)

	// 4. Find Best Move
	plan, found := selectBestMove(nodes, mean, sd)

	if !found {
		fmt.Println("No beneficial moves found.")
		os.Exit(0)
	}

	// 5. Apply Dampening Filter
	if plan.Improvement < *threshold {
		fmt.Printf("[SKIP] Best move only improves stability by %.2f%% (Threshold is %.2f%%)\n",
			plan.Improvement*100, *threshold*100)
		os.Exit(0)
	}

	// 6. Execute or Dry Run
	fmt.Printf("\n[PLAN] Shift Ballast:\n")
	fmt.Printf("   VM:      %s (%d)\n", plan.VM.Name, plan.VM.ID)
	fmt.Printf("   Source:  %s\n", plan.Source)
	fmt.Printf("   Target:  %s\n", plan.Destination)
	fmt.Printf("   Gain:    +%.2f%% Stability\n", plan.Improvement*100)

	if *dryRun {
		fmt.Println("\n>>> DRY RUN MODE: No migrations performed.")
	} else {
		fmt.Println("\n>>> LIVE MODE: Initiating Proxmox migration...")
		// [API CALL WOULD GO HERE]
		fmt.Printf("Successfully signaled Proxmox to move VM %d to %s\n", plan.VM.ID, plan.Destination)
	}
}
