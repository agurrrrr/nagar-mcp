package proxmox

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

// ListLXCInput contains the node name
type ListLXCInput struct {
	Node string `json:"node" jsonschema:"Proxmox node name"`
}

// ListLXCOutput contains the list of LXC containers
type ListLXCOutput struct {
	Containers []ContainerInfo `json:"containers"`
}

// LXCStatusInput contains node and vmid
type LXCStatusInput struct {
	Node string `json:"node" jsonschema:"Proxmox node name"`
	VMID int    `json:"vmid" jsonschema:"LXC container ID"`
}

// LXCStatusOutput contains detailed LXC status
type LXCStatusOutput struct {
	VMID    int    `json:"vmid"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	CPUs    int    `json:"cpus"`
	MaxMem  int64  `json:"maxmem"`
	MaxDisk int64  `json:"maxdisk"`
	MaxSwap int64  `json:"maxswap"`
	Uptime  int64  `json:"uptime"`
}

// LXCConfigInput contains node and vmid
type LXCConfigInput struct {
	Node string `json:"node" jsonschema:"Proxmox node name"`
	VMID int    `json:"vmid" jsonschema:"LXC container ID"`
}

// LXCConfigOutput contains LXC configuration
type LXCConfigOutput struct {
	VMID         int               `json:"vmid"`
	Hostname     string            `json:"hostname"`
	Memory       int               `json:"memory"`
	Swap         int               `json:"swap"`
	Cores        int               `json:"cores"`
	OSType       string            `json:"ostype,omitempty"`
	Arch         string            `json:"arch,omitempty"`
	RootFS       string            `json:"rootfs,omitempty"`
	Net          map[string]string `json:"net,omitempty"`
	MP           map[string]string `json:"mp,omitempty"`
	Description  string            `json:"description,omitempty"`
	Unprivileged bool              `json:"unprivileged"`
}

// === Tool Handlers ===

func createListLXCHandler(client *Client) mcp.ToolHandlerFor[ListLXCInput, ListLXCOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListLXCInput]) (*mcp.CallToolResultFor[ListLXCOutput], error) {
		input := params.Arguments

		if input.Node == "" {
			return toolErrorFor[ListLXCOutput]("node parameter is required"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[ListLXCOutput]("failed to query node: %s", err.Error()), nil
		}

		containers, err := node.Containers(ctx)
		if err != nil {
			return toolErrorfFor[ListLXCOutput]("failed to list LXC containers: %s", err.Error()), nil
		}

		output := ListLXCOutput{
			Containers: make([]ContainerInfo, 0, len(containers)),
		}

		for _, ct := range containers {
			output.Containers = append(output.Containers, ContainerInfo{
				VMID:    int(ct.VMID),
				Name:    ct.Name,
				Status:  ct.Status,
				CPUs:    ct.CPUs,
				MaxMem:  int64(ct.MaxMem),
				MaxDisk: int64(ct.MaxDisk),
				MaxSwap: int64(ct.MaxSwap),
				Uptime:  int64(ct.Uptime),
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createLXCStatusHandler(client *Client) mcp.ToolHandlerFor[LXCStatusInput, LXCStatusOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[LXCStatusInput]) (*mcp.CallToolResultFor[LXCStatusOutput], error) {
		input := params.Arguments

		if input.Node == "" {
			return toolErrorFor[LXCStatusOutput]("node parameter is required"), nil
		}
		if input.VMID == 0 {
			return toolErrorFor[LXCStatusOutput]("vmid parameter is required"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[LXCStatusOutput]("failed to query node: %s", err.Error()), nil
		}

		ct, err := node.Container(ctx, input.VMID)
		if err != nil {
			return toolErrorfFor[LXCStatusOutput]("failed to query LXC container: %s", err.Error()), nil
		}

		output := LXCStatusOutput{
			VMID:    int(ct.VMID),
			Name:    ct.Name,
			Status:  ct.Status,
			CPUs:    ct.CPUs,
			MaxMem:  int64(ct.MaxMem),
			MaxDisk: int64(ct.MaxDisk),
			MaxSwap: int64(ct.MaxSwap),
			Uptime:  int64(ct.Uptime),
		}

		return toolSuccessFor(output), nil
	}
}

func createLXCConfigHandler(client *Client) mcp.ToolHandlerFor[LXCConfigInput, LXCConfigOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[LXCConfigInput]) (*mcp.CallToolResultFor[LXCConfigOutput], error) {
		input := params.Arguments

		if input.Node == "" {
			return toolErrorFor[LXCConfigOutput]("node parameter is required"), nil
		}
		if input.VMID == 0 {
			return toolErrorFor[LXCConfigOutput]("vmid parameter is required"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[LXCConfigOutput]("failed to query node: %s", err.Error()), nil
		}

		ct, err := node.Container(ctx, input.VMID)
		if err != nil {
			return toolErrorfFor[LXCConfigOutput]("failed to query LXC container: %s", err.Error()), nil
		}

		// Config method returns Task and error, but we can access ContainerConfig directly
		config := ct.ContainerConfig
		if config == nil {
			return toolErrorFor[LXCConfigOutput]("unable to retrieve LXC configuration"), nil
		}

		output := LXCConfigOutput{
			VMID:        int(ct.VMID),
			Hostname:    config.Hostname,
			Memory:      int(config.Memory),
			Swap:        int(config.Swap),
			Cores:       int(config.Cores),
			Arch:        config.Arch,
			Description: config.Description,
		}

		return toolSuccessFor(output), nil
	}
}

// === Registration ===

func registerLXCTools(server *mcp.Server, client *Client) {
	// pve_list_lxc
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_list_lxc",
		Description: "Lists all LXC containers on a specific node. Includes status, CPU, memory, and disk usage for each container.",
	}, createListLXCHandler(client))

	// pve_lxc_status
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_lxc_status",
		Description: "Retrieves detailed status of a specific LXC container. Provides running state, resource usage, network I/O, and more.",
	}, createLXCStatusHandler(client))

	// pve_lxc_config
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_lxc_config",
		Description: "Retrieves configuration of a specific LXC container. Provides hostname, CPU, memory, storage, and network settings.",
	}, createLXCConfigHandler(client))

	slog.Info("LXC tools registered: pve_list_lxc, pve_lxc_status, pve_lxc_config")
}
