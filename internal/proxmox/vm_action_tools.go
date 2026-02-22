package proxmox

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types for VM Actions ===

// VMStartInput contains parameters for starting a VM
type VMStartInput struct {
	Node string `json:"node" jsonschema:"Proxmox node name"`
	VMID int    `json:"vmid" jsonschema:"VM ID"`
}

// VMStartOutput contains the result of starting a VM
type VMStartOutput struct {
	UPID    string `json:"upid"`
	Message string `json:"message"`
}

// VMStopInput contains parameters for stopping a VM
type VMStopInput struct {
	Node    string `json:"node" jsonschema:"Proxmox node name"`
	VMID    int    `json:"vmid" jsonschema:"VM ID"`
	Confirm bool   `json:"confirm" jsonschema:"Must be set to true to execute the stop (risk of data loss)"`
}

// VMStopOutput contains the result of stopping a VM
type VMStopOutput struct {
	UPID    string `json:"upid,omitempty"`
	Message string `json:"message"`
}

// VMShutdownInput contains parameters for gracefully shutting down a VM
type VMShutdownInput struct {
	Node    string `json:"node" jsonschema:"Proxmox node name"`
	VMID    int    `json:"vmid" jsonschema:"VM ID"`
	Timeout int    `json:"timeout,omitempty" jsonschema:"Shutdown timeout in seconds"`
}

// VMShutdownOutput contains the result of shutting down a VM
type VMShutdownOutput struct {
	UPID    string `json:"upid"`
	Message string `json:"message"`
}

// VMRebootInput contains parameters for rebooting a VM
type VMRebootInput struct {
	Node string `json:"node" jsonschema:"Proxmox node name"`
	VMID int    `json:"vmid" jsonschema:"VM ID"`
}

// VMRebootOutput contains the result of rebooting a VM
type VMRebootOutput struct {
	UPID    string `json:"upid"`
	Message string `json:"message"`
}

// VMSnapshotListInput contains parameters for listing VM snapshots
type VMSnapshotListInput struct {
	Node string `json:"node" jsonschema:"Proxmox node name"`
	VMID int    `json:"vmid" jsonschema:"VM ID"`
}

// VMSnapshotListOutput contains the list of VM snapshots
type VMSnapshotListOutput struct {
	Snapshots []SnapshotInfo `json:"snapshots"`
}

// VMSnapshotCreateInput contains parameters for creating a VM snapshot
type VMSnapshotCreateInput struct {
	Node        string `json:"node" jsonschema:"Proxmox node name"`
	VMID        int    `json:"vmid" jsonschema:"VM ID"`
	Name        string `json:"name" jsonschema:"Snapshot name"`
	Description string `json:"description,omitempty" jsonschema:"Snapshot description"`
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
				Content: []mcp.Content{&mcp.TextContent{Text: "node parameter is required"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[VMStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid parameter is required"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[VMStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query node: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[VMStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query VM: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := vm.Start(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[VMStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to start VM: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[VMStartOutput]{
			StructuredContent: VMStartOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("VM %d start requested (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

func createVMStopHandler(client *Client) mcp.ToolHandlerFor[VMStopInput, VMStopOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[VMStopInput]) (*mcp.CallToolResultFor[VMStopOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[VMStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node parameter is required"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[VMStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid parameter is required"}},
				IsError: true,
			}, nil
		}
		if !input.Confirm {
			return &mcp.CallToolResultFor[VMStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "WARNING: Force stopping a VM may cause data loss. Please set confirm: true. For a graceful shutdown, use pve_vm_shutdown instead."}},
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[VMStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query node: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[VMStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query VM: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := vm.Stop(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[VMStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to stop VM: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[VMStopOutput]{
			StructuredContent: VMStopOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("VM %d stop requested (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

func createVMShutdownHandler(client *Client) mcp.ToolHandlerFor[VMShutdownInput, VMShutdownOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[VMShutdownInput]) (*mcp.CallToolResultFor[VMShutdownOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[VMShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node parameter is required"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[VMShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid parameter is required"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[VMShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query node: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[VMShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query VM: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := vm.Shutdown(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[VMShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to shutdown VM: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[VMShutdownOutput]{
			StructuredContent: VMShutdownOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("VM %d graceful shutdown requested (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

func createVMRebootHandler(client *Client) mcp.ToolHandlerFor[VMRebootInput, VMRebootOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[VMRebootInput]) (*mcp.CallToolResultFor[VMRebootOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[VMRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node parameter is required"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[VMRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid parameter is required"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[VMRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query node: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[VMRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query VM: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := vm.Reboot(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[VMRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to reboot VM: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[VMRebootOutput]{
			StructuredContent: VMRebootOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("VM %d reboot requested (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

func createVMSnapshotListHandler(client *Client) mcp.ToolHandlerFor[VMSnapshotListInput, VMSnapshotListOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[VMSnapshotListInput]) (*mcp.CallToolResultFor[VMSnapshotListOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[VMSnapshotListOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node parameter is required"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[VMSnapshotListOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid parameter is required"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[VMSnapshotListOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query node: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[VMSnapshotListOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query VM: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		snapshots, err := vm.Snapshots(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[VMSnapshotListOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to list snapshots: %s", err.Error())}},
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
				Content: []mcp.Content{&mcp.TextContent{Text: "node parameter is required"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[VMSnapshotCreateOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid parameter is required"}},
				IsError: true,
			}, nil
		}
		if input.Name == "" {
			return &mcp.CallToolResultFor[VMSnapshotCreateOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "name parameter is required"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[VMSnapshotCreateOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query node: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[VMSnapshotCreateOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query VM: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := vm.NewSnapshot(ctx, input.Name)
		if err != nil {
			return &mcp.CallToolResultFor[VMSnapshotCreateOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to create snapshot: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[VMSnapshotCreateOutput]{
			StructuredContent: VMSnapshotCreateOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("VM %d snapshot '%s' creation requested (UPID: %s)", input.VMID, input.Name, string(task.UPID)),
			},
		}, nil
	}
}

// === Registration ===

func registerVMActionTools(server *mcp.Server, client *Client) {
	// pve_vm_start
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_start",
		Description: "Starts a VM. Returns a UPID for the asynchronous task.",
	}, createVMStartHandler(client))

	// pve_vm_stop
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_stop",
		Description: "Force stops a VM. Requires confirm: true due to risk of data loss. For graceful shutdown, use pve_vm_shutdown.",
	}, createVMStopHandler(client))

	// pve_vm_shutdown
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_shutdown",
		Description: "Gracefully shuts down a VM (ACPI shutdown). More reliable when Guest Agent is installed.",
	}, createVMShutdownHandler(client))

	// pve_vm_reboot
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_reboot",
		Description: "Reboots a VM.",
	}, createVMRebootHandler(client))

	// pve_vm_snapshot_list
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_snapshot_list",
		Description: "Lists snapshots of a VM.",
	}, createVMSnapshotListHandler(client))

	// pve_vm_snapshot_create
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_snapshot_create",
		Description: "Creates a snapshot of a VM.",
	}, createVMSnapshotCreateHandler(client))

	slog.Info("VM Action tools registered: pve_vm_start, pve_vm_stop, pve_vm_shutdown, pve_vm_reboot, pve_vm_snapshot_list, pve_vm_snapshot_create")
}
