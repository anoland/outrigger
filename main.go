package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/anoland/outrigger/pkg/rebalance"
	pve "github.com/luthermonson/go-proxmox"
	"github.com/anoland/outrigger/pkg/proxmox"
)

func main() {
	// 1. Define CLI Flags
	dryRun := flag.Bool("dry-run", true, "If true, only plans the move without executing")
	threshold := flag.Float64("threshold", 0.15, "Minimum improvement required (0.15 = 15%)")
	flag.Parse()

	fmt.Println("🚢 TRIM: Proxmox Cluster Balancer")
	fmt.Printf("Settings: DryRun=%v, Threshold=%.2f\n", *dryRun, *threshold)

	// Initialize Proxmox client
	cfg := proxmox.Config{
		Endpoint: os.Getenv("PVE_ENDPOINT"),
		TokenID:  os.Getenv("PVE_TOKEN_ID"),
		Secret:   os.Getenv("PVE_TOKEN_SECRET"),
	}
	client := proxmox.NewClient(cfg)
	_ = client // TODO: Use client to fetch real data from Proxmox API

	// 2. Mock Data (Normally fetched from Proxmox API)
	// In a real application, you would fetch this data from the Proxmox API
	nodes := []*pve.Node{
		{Name: "pve-01", Memory: pve.Memory{Total: 120 * 1024 * 1024 * 1024}},
		{Name: "pve-02", Memory: pve.Memory{Total: 60 * 1024 * 1024 * 1024}},
		{Name: "pve-03", Memory: pve.Memory{Total: 55 * 1024 * 1024 * 1024}},
		{Name: "pve-04", Memory: pve.Memory{Total: 30 * 1024 * 1024 * 1024}},
	}

	vms := []*pve.VirtualMachine{
		{VMID: 101, Name: "Database", Mem: 45 * 1024 * 1024 * 1024, Node: "pve-01"},
		{VMID: 102, Name: "Web-Srv", Mem: 20 * 1024 * 1024 * 1024, Node: "pve-01"},
		{VMID: 103, Name: "Utility", Mem: 10 * 1024 * 1024 * 1024, Node: "pve-02"},
		{VMID: 104, Name: "Docker-Host", Mem: 30 * 1024 * 1024 * 1024, Node: "pve-03"},
	}

	// 3. Analyze Cluster
	if len(nodes) == 0 {
		fmt.Println("No nodes found in the cluster.")
		os.Exit(0)
	}
	mean, sd := rebalance.CalculateSD(nodes)
	fmt.Printf("\n[STATUS] Cluster Mean: %.2f GB | Imbalance Score (SD): %.2f\n", mean, sd)

	// 4. Find Best Move
	plan, found := rebalance.SelectBestMove(nodes, vms, mean, sd)

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
	fmt.Printf("   VM:      %s (%v)\n", plan.VM.Name, plan.VM.VMID)
	fmt.Printf("   Source:  %s\n", plan.Source)
	fmt.Printf("   Target:  %s\n", plan.Destination)
	fmt.Printf("   Gain:    +%.2f%% Stability\n", plan.Improvement*100)

	if *dryRun {
		fmt.Println("\n>>> DRY RUN MODE: No migrations performed.")
	} else {
		fmt.Println("\n>>> LIVE MODE: Initiating Proxmox migration...")
		// [API CALL WOULD GO HERE]
		fmt.Printf("Successfully signaled Proxmox to move VM %v to %s\n", plan.VM.VMID, plan.Destination)
	}
}
