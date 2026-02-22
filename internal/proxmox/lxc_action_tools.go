package proxmox

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types for LXC Actions ===

// LXCStartInput contains parameters for starting an LXC container
type LXCStartInput struct {
	Node string `json:"node" jsonschema:"Proxmox node name"`
	VMID int    `json:"vmid" jsonschema:"LXC container ID"`
}

// LXCStartOutput contains the result of starting an LXC container
type LXCStartOutput struct {
	UPID    string `json:"upid"`
	Message string `json:"message"`
}

// LXCStopInput contains parameters for stopping an LXC container
type LXCStopInput struct {
	Node    string `json:"node" jsonschema:"Proxmox node name"`
	VMID    int    `json:"vmid" jsonschema:"LXC container ID"`
	Confirm bool   `json:"confirm" jsonschema:"Must be set to true to execute stop"`
}

// LXCStopOutput contains the result of stopping an LXC container
type LXCStopOutput struct {
	UPID    string `json:"upid,omitempty"`
	Message string `json:"message"`
}

// LXCShutdownInput contains parameters for gracefully shutting down an LXC container
type LXCShutdownInput struct {
	Node    string `json:"node" jsonschema:"Proxmox node name"`
	VMID    int    `json:"vmid" jsonschema:"LXC container ID"`
	Timeout int    `json:"timeout,omitempty" jsonschema:"Shutdown timeout in seconds"`
}

// LXCShutdownOutput contains the result of shutting down an LXC container
type LXCShutdownOutput struct {
	UPID    string `json:"upid"`
	Message string `json:"message"`
}

// LXCRebootInput contains parameters for rebooting an LXC container
type LXCRebootInput struct {
	Node string `json:"node" jsonschema:"Proxmox node name"`
	VMID int    `json:"vmid" jsonschema:"LXC container ID"`
}

// LXCRebootOutput contains the result of rebooting an LXC container
type LXCRebootOutput struct {
	UPID    string `json:"upid"`
	Message string `json:"message"`
}

// === Tool Handlers ===

func createLXCStartHandler(client *Client) mcp.ToolHandlerFor[LXCStartInput, LXCStartOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[LXCStartInput]) (*mcp.CallToolResultFor[LXCStartOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[LXCStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node parameter is required"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[LXCStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid parameter is required"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[LXCStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query node: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		ct, err := node.Container(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[LXCStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query LXC container: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := ct.Start(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[LXCStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to start LXC container: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[LXCStartOutput]{
			StructuredContent: LXCStartOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("LXC %d start requested (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

func createLXCStopHandler(client *Client) mcp.ToolHandlerFor[LXCStopInput, LXCStopOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[LXCStopInput]) (*mcp.CallToolResultFor[LXCStopOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[LXCStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node parameter is required"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[LXCStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid parameter is required"}},
				IsError: true,
			}, nil
		}
		if !input.Confirm {
			return &mcp.CallToolResultFor[LXCStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "WARNING: Force stopping an LXC container may cause data loss. Please set confirm: true. For a graceful shutdown, use pve_lxc_shutdown."}},
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[LXCStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query node: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		ct, err := node.Container(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[LXCStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query LXC container: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := ct.Stop(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[LXCStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to stop LXC container: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[LXCStopOutput]{
			StructuredContent: LXCStopOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("LXC %d stop requested (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

func createLXCShutdownHandler(client *Client) mcp.ToolHandlerFor[LXCShutdownInput, LXCShutdownOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[LXCShutdownInput]) (*mcp.CallToolResultFor[LXCShutdownOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[LXCShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node parameter is required"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[LXCShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid parameter is required"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[LXCShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query node: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		ct, err := node.Container(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[LXCShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query LXC container: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		// Shutdown with optional timeout
		timeout := input.Timeout
		if timeout == 0 {
			timeout = 30
		}
		task, err := ct.Shutdown(ctx, false, timeout)
		if err != nil {
			return &mcp.CallToolResultFor[LXCShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to shut down LXC container: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[LXCShutdownOutput]{
			StructuredContent: LXCShutdownOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("LXC %d graceful shutdown requested (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

func createLXCRebootHandler(client *Client) mcp.ToolHandlerFor[LXCRebootInput, LXCRebootOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[LXCRebootInput]) (*mcp.CallToolResultFor[LXCRebootOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[LXCRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node parameter is required"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[LXCRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid parameter is required"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[LXCRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query node: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		ct, err := node.Container(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[LXCRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to query LXC container: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := ct.Reboot(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[LXCRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to reboot LXC container: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[LXCRebootOutput]{
			StructuredContent: LXCRebootOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("LXC %d reboot requested (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

// === Registration ===

func registerLXCActionTools(server *mcp.Server, client *Client) {
	// pve_lxc_start
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_lxc_start",
		Description: "Starts an LXC container. Returns a UPID for the asynchronous task.",
	}, createLXCStartHandler(client))

	// pve_lxc_stop
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_lxc_stop",
		Description: "Force stops an LXC container. Requires confirm: true due to risk of data loss.",
	}, createLXCStopHandler(client))

	// pve_lxc_shutdown
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_lxc_shutdown",
		Description: "Gracefully shuts down an LXC container.",
	}, createLXCShutdownHandler(client))

	// pve_lxc_reboot
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_lxc_reboot",
		Description: "Reboots an LXC container.",
	}, createLXCRebootHandler(client))

	slog.Info("LXC Action tools registered: pve_lxc_start, pve_lxc_stop, pve_lxc_shutdown, pve_lxc_reboot")
}
