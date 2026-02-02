package proxmox

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

// ListVMsInput contains the node name
type ListVMsInput struct {
	Node string `json:"node" jsonschema:"description=Proxmox 노드 이름"`
}

// ListVMsOutput contains the list of VMs
type ListVMsOutput struct {
	VMs []VMInfo `json:"vms"`
}

// VMStatusInput contains node and vmid
type VMStatusInput struct {
	Node string `json:"node" jsonschema:"description=Proxmox 노드 이름"`
	VMID int    `json:"vmid" jsonschema:"description=VM ID"`
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
	Node string `json:"node" jsonschema:"description=Proxmox 노드 이름"`
	VMID int    `json:"vmid" jsonschema:"description=VM ID"`
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
		StructuredContent: output,
	}
}

// === Tool Handlers ===

func createListVMsHandler(client *Client) mcp.ToolHandlerFor[ListVMsInput, ListVMsOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListVMsInput]) (*mcp.CallToolResultFor[ListVMsOutput], error) {
		input := params.Arguments

		if input.Node == "" {
			return toolErrorFor[ListVMsOutput]("node 파라미터가 필요합니다"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[ListVMsOutput]("노드 조회 실패: %s", err.Error()), nil
		}

		vms, err := node.VirtualMachines(ctx)
		if err != nil {
			return toolErrorfFor[ListVMsOutput]("VM 목록 조회 실패: %s", err.Error()), nil
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
			return toolErrorFor[VMStatusOutput]("node 파라미터가 필요합니다"), nil
		}
		if input.VMID == 0 {
			return toolErrorFor[VMStatusOutput]("vmid 파라미터가 필요합니다"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[VMStatusOutput]("노드 조회 실패: %s", err.Error()), nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return toolErrorfFor[VMStatusOutput]("VM 조회 실패: %s", err.Error()), nil
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
			return toolErrorFor[VMConfigOutput]("node 파라미터가 필요합니다"), nil
		}
		if input.VMID == 0 {
			return toolErrorFor[VMConfigOutput]("vmid 파라미터가 필요합니다"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[VMConfigOutput]("노드 조회 실패: %s", err.Error()), nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return toolErrorfFor[VMConfigOutput]("VM 조회 실패: %s", err.Error()), nil
		}

		// VirtualMachineConfig is embedded in VirtualMachine
		config := vm.VirtualMachineConfig
		if config == nil {
			return toolErrorFor[VMConfigOutput]("VM 설정을 가져올 수 없습니다"), nil
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
		Description: "특정 노드의 모든 QEMU VM 목록을 조회합니다. 각 VM의 상태, CPU, 메모리, 디스크 사용량 정보를 포함합니다.",
	}, createListVMsHandler(client))

	// pve_vm_status
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_status",
		Description: "특정 VM의 상세 상태를 조회합니다. 실행 상태, 리소스 사용량, 네트워크 I/O 등의 정보를 제공합니다.",
	}, createVMStatusHandler(client))

	// pve_vm_config
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_config",
		Description: "특정 VM의 설정을 조회합니다. CPU, 메모리, 디스크, 네트워크 설정 등의 정보를 제공합니다.",
	}, createVMConfigHandler(client))

	fmt.Println("VM tools registered: pve_list_vms, pve_vm_status, pve_vm_config")
}
