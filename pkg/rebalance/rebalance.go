package rebalance

import (
	"math"

	"github.com/luthermonson/go-proxmox"
)

// Plan represents a plan to move a VM from a source to a destination node.
type Plan struct {
	VM          *proxmox.VirtualMachine
	Source      string
	Destination string
	Improvement float64
}

// CalculateSD returns the mean and standard deviation of memory usage across the nodes.
func CalculateSD(nodes []*proxmox.Node) (float64, float64) {
	var sum float64
	for _, n := range nodes {
		sum += float64(n.Memory.Total) / 1024 / 1024 / 1024
	}
	mean := sum / float64(len(nodes))

	var sqDiffSum float64
	for _, n := range nodes {
		memUsedGB := float64(n.Memory.Total) / 1024 / 1024 / 1024
		sqDiffSum += math.Pow(memUsedGB-mean, 2)
	}
	sd := math.Sqrt(sqDiffSum / float64(len(nodes)))
	return mean, sd
}

// SelectBestMove finds the best VM to move to improve the cluster's memory balance.
func SelectBestMove(nodes []*proxmox.Node, vms []*proxmox.VirtualMachine, mean float64, currentSD float64) (Plan, bool) {
	var bestPlan Plan
	var found bool
	maxImprovement := 0.0

	// Identify the heavy and light nodes
	var heavyNode, lightNode *proxmox.Node
	maxDist, minDist := -1.0, 1.0

	for _, n := range nodes {
		memUsedGB := float64(n.Memory.Total) / 1024 / 1024 / 1024
		dist := memUsedGB - mean
		if dist > maxDist {
			maxDist = dist
			heavyNode = n
		}
		if dist < minDist {
			minDist = dist
			lightNode = n
		}
	}

	if heavyNode == nil || lightNode == nil || heavyNode.Name == lightNode.Name {
		return bestPlan, false
	}

	// Evaluate VMs on the heavy node
	heavyNodeMemUsedGB := float64(heavyNode.Memory.Total) / 1024 / 1024 / 1024
	overage := heavyNodeMemUsedGB - mean
	for _, vm := range vms {
		if vm.Node != heavyNode.Name {
			continue
		}

		vmMemUsedGB := float64(vm.Mem) / 1024 / 1024 / 1024
		// Don't move a VM that is larger than the overage + 10% (prevents over-correction)
		if vmMemUsedGB > overage*1.1 {
			continue
		}

		// Simulate the move
		projectedLoads := []float64{}
		for _, n := range nodes {
			load := float64(n.Memory.Total) / 1024 / 1024 / 1024
			if n.Name == heavyNode.Name {
				load -= vmMemUsedGB
			} else if n.Name == lightNode.Name {
				load += vmMemUsedGB
			}
			projectedLoads = append(projectedLoads, load)
		}

		// Calculate the projected improvement
		_, projSD := calculateSDFromLoads(projectedLoads)
		improvement := (currentSD - projSD) / currentSD

		if improvement > maxImprovement {
			maxImprovement = improvement
			bestPlan = Plan{
				VM:          vm,
				Source:      heavyNode.Name,
				Destination: lightNode.Name,
				Improvement: improvement,
			}
			found = true
		}
	}

	return bestPlan, found
}

// calculateSDFromLoads is a helper function to calculate the standard deviation from a slice of memory loads.
func calculateSDFromLoads(loads []float64) (float64, float64) {
	var sum float64
	for _, l := range loads {
		sum += l
	}
	mean := sum / float64(len(loads))
	var sqDiffSum float64
	for _, l := range loads {
		sqDiffSum += math.Pow(l-mean, 2)
	}
	return mean, math.Sqrt(sqDiffSum / float64(len(loads)))
}
