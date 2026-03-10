package main_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luthermonson/go-proxmox"
	"github.com/anoland/outrigger/pkg/rebalance"
)

var _ = Describe("Main", func() {
	Context("with mock data", func() {
		nodes := []*proxmox.Node{
			{Name: "pve-01", Memory: proxmox.Memory{Total: 120 * 1024 * 1024 * 1024}},
			{Name: "pve-02", Memory: proxmox.Memory{Total: 60 * 1024 * 1024 * 1024}},
			{Name: "pve-03", Memory: proxmox.Memory{Total: 55 * 1024 * 1024 * 1024}},
			{Name: "pve-04", Memory: proxmox.Memory{Total: 30 * 1024 * 1024 * 1024}},
		}

		vms := []*proxmox.VirtualMachine{
			{VMID: 101, Name: "Database", Mem: 45 * 1024 * 1024 * 1024, Node: "pve-01"},
			{VMID: 102, Name: "Web-Srv", Mem: 20 * 1024 * 1024 * 1024, Node: "pve-01"},
			{VMID: 103, Name: "Utility", Mem: 10 * 1024 * 1024 * 1024, Node: "pve-02"},
			{VMID: 104, Name: "Docker-Host", Mem: 30 * 1024 * 1024 * 1024, Node: "pve-03"},
		}

		It("should calculate SD correctly", func() {
			mean, sd := rebalance.CalculateSD(nodes)
			Expect(mean).To(BeNumerically("~", 66.25))
			Expect(sd).To(BeNumerically("~", 33.05, 0.01))
		})

		It("should select the best move", func() {
			mean, sd := rebalance.CalculateSD(nodes)
			plan, found := rebalance.SelectBestMove(nodes, vms, mean, sd)

			Expect(found).To(BeTrue())
			Expect(plan.VM.VMID).To(Equal(proxmox.StringOrUint64(101)))
			Expect(plan.Source).To(Equal("pve-01"))
			Expect(plan.Destination).To(Equal("pve-04"))
			Expect(plan.Improvement).To(BeNumerically("~", 0.729, 0.001))
		})
	})
})
