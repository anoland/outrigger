package main

import (
	"math"
)

// VM represents a virtual machine in the Proxmox cluster.
type VM struct {
	ID        int
	Name      string
	MemUsedGB float64
}

// Node represents a node in the Proxmox cluster.
type Node struct {
	Name      string
	MemUsedGB float64
	VMs       []VM
}

// RebalancePlan represents a plan to move a VM from a source to a destination node.
type RebalancePlan struct {
	VM          VM
	Source      string
	Destination string
	Improvement float64
}

// calculateSD returns the mean and standard deviation of memory usage across the nodes.
func calculateSD(nodes []Node) (float64, float64) {
	var sum float64
	for _, n := range nodes {
		sum += n.MemUsedGB
	}
	mean := sum / float64(len(nodes))

	var sqDiffSum float64
	for _, n := range nodes {
		sqDiffSum += math.Pow(n.MemUsedGB-mean, 2)
	}
	sd := math.Sqrt(sqDiffSum / float64(len(nodes)))
	return mean, sd
}

// selectBestMove finds the best VM to move to improve the cluster's memory balance.
func selectBestMove(nodes []Node, mean float64, currentSD float64) (RebalancePlan, bool) {
	var bestPlan RebalancePlan
	var found bool
	maxImprovement := 0.0

	// Identify the heavy and light nodes
	var heavyNode, lightNode *Node
	maxDist, minDist := -1.0, 1.0

	for i := range nodes {
		dist := nodes[i].MemUsedGB - mean
		if dist > maxDist {
			maxDist = dist
			heavyNode = &nodes[i]
		}
		if dist < minDist {
			minDist = dist
			lightNode = &nodes[i]
		}
	}

	if heavyNode == nil || lightNode == nil || heavyNode == lightNode {
		return bestPlan, false
	}

	// Evaluate VMs on the heavy node
	overage := heavyNode.MemUsedGB - mean
	for _, vm := range heavyNode.VMs {
		// Don't move a VM that is larger than the overage + 10% (prevents over-correction)
		if vm.MemUsedGB > overage*1.1 {
			continue
		}

		// Simulate the move
		projectedLoads := []float64{}
		for _, n := range nodes {
			load := n.MemUsedGB
			if n.Name == heavyNode.Name {
				load -= vm.MemUsedGB
			} else if n.Name == lightNode.Name {
				load += vm.MemUsedGB
			}
			projectedLoads = append(projectedLoads, load)
		}

		// Calculate the projected improvement
		_, projSD := calculateSDFromLoads(projectedLoads)
		improvement := (currentSD - projSD) / currentSD

		if improvement > maxImprovement {
			maxImprovement = improvement
			bestPlan = RebalancePlan{
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
