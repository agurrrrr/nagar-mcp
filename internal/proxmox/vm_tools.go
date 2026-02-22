package proxmox

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

// ListVMsInput contains the node name
type ListVMsInput struct {
	Node string `json:"node" jsonschema:"Proxmox node name"`
}

// ListVMsOutput contains the list of VMs
type ListVMsOutput struct {
	VMs []VMInfo `json:"vms"`
}

// VMStatusInput contains node and vmid
type VMStatusInput struct {
	Node string `json:"node" jsonschema:"Proxmox node name"`
	VMID int    `json:"vmid" jsonschema:"VM ID"`
}

// VMStatusOutput contains detailed VM status
type VMStatusOutput struct {
	VMID      int     `json:"vmid"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	CPU       float64 `json:"cpu"`
	CPUs      int     `json:"cpus"`
	Mem       int64   `json:"mem"`
	MaxMem    int64   `json:"maxmem"`
	Disk      int64   `json:"disk"`
	MaxDisk   int64   `json:"maxdisk"`
	Uptime    int64   `json:"uptime"`
	NetIn     int64   `json:"netin"`
	NetOut    int64   `json:"netout"`
	DiskRead  int64   `json:"diskread"`
	DiskWrite int64   `json:"diskwrite"`
	PID       int     `json:"pid,omitempty"`
	QMPStatus string  `json:"qmp_status,omitempty"`
}

// VMConfigInput contains node and vmid
type VMConfigInput struct {
	Node string `json:"node" jsonschema:"Proxmox node name"`
	VMID int    `json:"vmid" jsonschema:"VM ID"`
}

// VMConfigOutput contains VM configuration
type VMConfigOutput struct {
	VMID        int    `json:"vmid"`
	Name        string `json:"name"`
	Memory      int    `json:"memory"`
	Cores       int    `json:"cores"`
	Sockets     int    `json:"sockets"`
	OSType      string `json:"ostype,omitempty"`
	Boot        string `json:"boot,omitempty"`
	Description string `json:"description,omitempty"`
}

// === Helper functions for typed results ===

func toolErrorFor[Out any](msg string) *mcp.CallToolResultFor[Out] {
	return &mcp.CallToolResultFor[Out]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
		IsError: true,
	}
}

func toolErrorfFor[Out any](format string, args ...any) *mcp.CallToolResultFor[Out] {
	return toolErrorFor[Out](fmt.Sprintf(format, args...))
}

func toolSuccessFor[Out any](output Out) *mcp.CallToolResultFor[Out] {
	return &mcp.CallToolResultFor[Out]{
		Content:           []mcp.Content{},
		StructuredContent: output,
	}
}

// === Tool Handlers ===

func createListVMsHandler(client *Client) mcp.ToolHandlerFor[ListVMsInput, ListVMsOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListVMsInput]) (*mcp.CallToolResultFor[ListVMsOutput], error) {
		input := params.Arguments

		if input.Node == "" {
			return toolErrorFor[ListVMsOutput]("node parameter is required"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[ListVMsOutput]("failed to query node: %s", err.Error()), nil
		}

		vms, err := node.VirtualMachines(ctx)
		if err != nil {
			return toolErrorfFor[ListVMsOutput]("failed to list VMs: %s", err.Error()), nil
		}

		output := ListVMsOutput{
			VMs: make([]VMInfo, 0, len(vms)),
		}

		for _, vm := range vms {
			output.VMs = append(output.VMs, VMInfo{
				VMID:      int(vm.VMID),
				Name:      vm.Name,
				Status:    vm.Status,
				CPU:       vm.CPU,
				MaxCPU:    vm.CPUs,
				Mem:       int64(vm.Mem),
				MaxMem:    int64(vm.MaxMem),
				Disk:      int64(vm.Disk),
				MaxDisk:   int64(vm.MaxDisk),
				Uptime:    int64(vm.Uptime),
				NetIn:     int64(vm.NetIn),
				NetOut:    int64(vm.Netout),
				DiskRead:  int64(vm.DiskRead),
				DiskWrite: int64(vm.DiskWrite),
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createVMStatusHandler(client *Client) mcp.ToolHandlerFor[VMStatusInput, VMStatusOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[VMStatusInput]) (*mcp.CallToolResultFor[VMStatusOutput], error) {
		input := params.Arguments

		if input.Node == "" {
			return toolErrorFor[VMStatusOutput]("node parameter is required"), nil
		}
		if input.VMID == 0 {
			return toolErrorFor[VMStatusOutput]("vmid parameter is required"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[VMStatusOutput]("failed to query node: %s", err.Error()), nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return toolErrorfFor[VMStatusOutput]("failed to query VM: %s", err.Error()), nil
		}

		output := VMStatusOutput{
			VMID:      int(vm.VMID),
			Name:      vm.Name,
			Status:    vm.Status,
			CPU:       vm.CPU,
			CPUs:      vm.CPUs,
			Mem:       int64(vm.Mem),
			MaxMem:    int64(vm.MaxMem),
			Disk:      int64(vm.Disk),
			MaxDisk:   int64(vm.MaxDisk),
			Uptime:    int64(vm.Uptime),
			NetIn:     int64(vm.NetIn),
			NetOut:    int64(vm.Netout),
			DiskRead:  int64(vm.DiskRead),
			DiskWrite: int64(vm.DiskWrite),
			PID:       int(vm.PID),
			QMPStatus: vm.QMPStatus,
		}

		return toolSuccessFor(output), nil
	}
}

func createVMConfigHandler(client *Client) mcp.ToolHandlerFor[VMConfigInput, VMConfigOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[VMConfigInput]) (*mcp.CallToolResultFor[VMConfigOutput], error) {
		input := params.Arguments

		if input.Node == "" {
			return toolErrorFor[VMConfigOutput]("node parameter is required"), nil
		}
		if input.VMID == 0 {
			return toolErrorFor[VMConfigOutput]("vmid parameter is required"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[VMConfigOutput]("failed to query node: %s", err.Error()), nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return toolErrorfFor[VMConfigOutput]("failed to query VM: %s", err.Error()), nil
		}

		// VirtualMachineConfig is embedded in VirtualMachine
		config := vm.VirtualMachineConfig
		if config == nil {
			return toolErrorFor[VMConfigOutput]("unable to retrieve VM configuration"), nil
		}

		output := VMConfigOutput{
			VMID:        int(vm.VMID),
			Name:        config.Name,
			Memory:      int(config.Memory),
			Cores:       config.Cores,
			Sockets:     config.Sockets,
			OSType:      config.OSType,
			Boot:        config.Boot,
			Description: config.Description,
		}

		return toolSuccessFor(output), nil
	}
}

// === Registration ===

func registerVMTools(server *mcp.Server, client *Client) {
	// pve_list_vms
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_list_vms",
		Description: "Lists all QEMU VMs on a specific node. Includes status, CPU, memory, and disk usage for each VM.",
	}, createListVMsHandler(client))

	// pve_vm_status
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_status",
		Description: "Retrieves detailed status of a specific VM. Provides runtime status, resource usage, network I/O, and more.",
	}, createVMStatusHandler(client))

	// pve_vm_config
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_config",
		Description: "Retrieves the configuration of a specific VM. Provides CPU, memory, disk, and network configuration details.",
	}, createVMConfigHandler(client))

	slog.Info("VM tools registered: pve_list_vms, pve_vm_status, pve_vm_config")
}
