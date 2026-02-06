package proxmox

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types for VM Actions ===

// VMStartInput contains parameters for starting a VM
type VMStartInput struct {
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
	VMID int    `json:"vmid" jsonschema:"VM ID"`
}

// VMStartOutput contains the result of starting a VM
type VMStartOutput struct {
	UPID    string `json:"upid"`
	Message string `json:"message"`
}

// VMStopInput contains parameters for stopping a VM
type VMStopInput struct {
	Node    string `json:"node" jsonschema:"Proxmox 노드 이름"`
	VMID    int    `json:"vmid" jsonschema:"VM ID"`
	Confirm bool   `json:"confirm" jsonschema:"true로 설정해야 실제 정지 실행 (데이터 손실 위험)"`
}

// VMStopOutput contains the result of stopping a VM
type VMStopOutput struct {
	UPID    string `json:"upid,omitempty"`
	Message string `json:"message"`
}

// VMShutdownInput contains parameters for gracefully shutting down a VM
type VMShutdownInput struct {
	Node    string `json:"node" jsonschema:"Proxmox 노드 이름"`
	VMID    int    `json:"vmid" jsonschema:"VM ID"`
	Timeout int    `json:"timeout,omitempty" jsonschema:"종료 대기 시간 (초)"`
}

// VMShutdownOutput contains the result of shutting down a VM
type VMShutdownOutput struct {
	UPID    string `json:"upid"`
	Message string `json:"message"`
}

// VMRebootInput contains parameters for rebooting a VM
type VMRebootInput struct {
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
	VMID int    `json:"vmid" jsonschema:"VM ID"`
}

// VMRebootOutput contains the result of rebooting a VM
type VMRebootOutput struct {
	UPID    string `json:"upid"`
	Message string `json:"message"`
}

// VMSnapshotListInput contains parameters for listing VM snapshots
type VMSnapshotListInput struct {
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
	VMID int    `json:"vmid" jsonschema:"VM ID"`
}

// VMSnapshotListOutput contains the list of VM snapshots
type VMSnapshotListOutput struct {
	Snapshots []SnapshotInfo `json:"snapshots"`
}

// VMSnapshotCreateInput contains parameters for creating a VM snapshot
type VMSnapshotCreateInput struct {
	Node        string `json:"node" jsonschema:"Proxmox 노드 이름"`
	VMID        int    `json:"vmid" jsonschema:"VM ID"`
	Name        string `json:"name" jsonschema:"스냅샷 이름"`
	Description string `json:"description,omitempty" jsonschema:"스냅샷 설명"`
}

// VMSnapshotCreateOutput contains the result of creating a VM snapshot
type VMSnapshotCreateOutput struct {
	UPID    string `json:"upid"`
	Message string `json:"message"`
}

// === Tool Handlers ===

func createVMStartHandler(client *Client) mcp.ToolHandlerFor[VMStartInput, VMStartOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[VMStartInput]) (*mcp.CallToolResultFor[VMStartOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[VMStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[VMStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[VMStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("노드 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[VMStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("VM 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := vm.Start(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[VMStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("VM 시작 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[VMStartOutput]{
			StructuredContent: VMStartOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("VM %d 시작 요청됨 (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

func createVMStopHandler(client *Client) mcp.ToolHandlerFor[VMStopInput, VMStopOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[VMStopInput]) (*mcp.CallToolResultFor[VMStopOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[VMStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[VMStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}
		if !input.Confirm {
			return &mcp.CallToolResultFor[VMStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "⚠️ VM 강제 정지는 데이터 손실을 유발할 수 있습니다. confirm: true를 설정해주세요. 정상 종료를 원하면 pve_vm_shutdown을 사용하세요."}},
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[VMStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("노드 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[VMStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("VM 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := vm.Stop(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[VMStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("VM 정지 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[VMStopOutput]{
			StructuredContent: VMStopOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("VM %d 정지 요청됨 (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

func createVMShutdownHandler(client *Client) mcp.ToolHandlerFor[VMShutdownInput, VMShutdownOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[VMShutdownInput]) (*mcp.CallToolResultFor[VMShutdownOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[VMShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[VMShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[VMShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("노드 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[VMShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("VM 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := vm.Shutdown(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[VMShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("VM 종료 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[VMShutdownOutput]{
			StructuredContent: VMShutdownOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("VM %d 정상 종료 요청됨 (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

func createVMRebootHandler(client *Client) mcp.ToolHandlerFor[VMRebootInput, VMRebootOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[VMRebootInput]) (*mcp.CallToolResultFor[VMRebootOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[VMRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[VMRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[VMRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("노드 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[VMRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("VM 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := vm.Reboot(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[VMRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("VM 재시작 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[VMRebootOutput]{
			StructuredContent: VMRebootOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("VM %d 재시작 요청됨 (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

func createVMSnapshotListHandler(client *Client) mcp.ToolHandlerFor[VMSnapshotListInput, VMSnapshotListOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[VMSnapshotListInput]) (*mcp.CallToolResultFor[VMSnapshotListOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[VMSnapshotListOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[VMSnapshotListOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[VMSnapshotListOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("노드 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[VMSnapshotListOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("VM 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		snapshots, err := vm.Snapshots(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[VMSnapshotListOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("스냅샷 목록 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		output := VMSnapshotListOutput{
			Snapshots: make([]SnapshotInfo, 0, len(snapshots)),
		}

		for _, snap := range snapshots {
			output.Snapshots = append(output.Snapshots, SnapshotInfo{
				Name:        snap.Name,
				Description: snap.Description,
				SnapTime:    snap.Snaptime,
				Parent:      snap.Parent,
			})
		}

		return &mcp.CallToolResultFor[VMSnapshotListOutput]{
			StructuredContent: output,
		}, nil
	}
}

func createVMSnapshotCreateHandler(client *Client) mcp.ToolHandlerFor[VMSnapshotCreateInput, VMSnapshotCreateOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[VMSnapshotCreateInput]) (*mcp.CallToolResultFor[VMSnapshotCreateOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[VMSnapshotCreateOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[VMSnapshotCreateOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}
		if input.Name == "" {
			return &mcp.CallToolResultFor[VMSnapshotCreateOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "name 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[VMSnapshotCreateOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("노드 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[VMSnapshotCreateOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("VM 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := vm.NewSnapshot(ctx, input.Name)
		if err != nil {
			return &mcp.CallToolResultFor[VMSnapshotCreateOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("스냅샷 생성 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[VMSnapshotCreateOutput]{
			StructuredContent: VMSnapshotCreateOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("VM %d 스냅샷 '%s' 생성 요청됨 (UPID: %s)", input.VMID, input.Name, string(task.UPID)),
			},
		}, nil
	}
}

// === Registration ===

func registerVMActionTools(server *mcp.Server, client *Client) {
	// pve_vm_start
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_start",
		Description: "VM을 시작합니다. 비동기 작업으로 UPID가 반환됩니다.",
	}, createVMStartHandler(client))

	// pve_vm_stop
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_stop",
		Description: "VM을 강제 정지합니다. 데이터 손실 위험이 있으므로 confirm: true가 필요합니다. 정상 종료는 pve_vm_shutdown을 사용하세요.",
	}, createVMStopHandler(client))

	// pve_vm_shutdown
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_shutdown",
		Description: "VM을 정상 종료(ACPI shutdown)합니다. Guest Agent가 설치되어 있으면 더 안정적입니다.",
	}, createVMShutdownHandler(client))

	// pve_vm_reboot
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_reboot",
		Description: "VM을 재시작합니다.",
	}, createVMRebootHandler(client))

	// pve_vm_snapshot_list
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_snapshot_list",
		Description: "VM의 스냅샷 목록을 조회합니다.",
	}, createVMSnapshotListHandler(client))

	// pve_vm_snapshot_create
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_snapshot_create",
		Description: "VM의 스냅샷을 생성합니다.",
	}, createVMSnapshotCreateHandler(client))

	fmt.Println("VM Action tools registered: pve_vm_start, pve_vm_stop, pve_vm_shutdown, pve_vm_reboot, pve_vm_snapshot_list, pve_vm_snapshot_create")
}
