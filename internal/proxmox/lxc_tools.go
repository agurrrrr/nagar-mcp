package proxmox

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

// ListLXCInput contains the node name
type ListLXCInput struct {
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
}

// ListLXCOutput contains the list of LXC containers
type ListLXCOutput struct {
	Containers []ContainerInfo `json:"containers"`
}

// LXCStatusInput contains node and vmid
type LXCStatusInput struct {
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
	VMID int    `json:"vmid" jsonschema:"LXC 컨테이너 ID"`
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
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
	VMID int    `json:"vmid" jsonschema:"LXC 컨테이너 ID"`
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
			return toolErrorFor[ListLXCOutput]("node 파라미터가 필요합니다"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[ListLXCOutput]("노드 조회 실패: %s", err.Error()), nil
		}

		containers, err := node.Containers(ctx)
		if err != nil {
			return toolErrorfFor[ListLXCOutput]("LXC 목록 조회 실패: %s", err.Error()), nil
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
			return toolErrorFor[LXCStatusOutput]("node 파라미터가 필요합니다"), nil
		}
		if input.VMID == 0 {
			return toolErrorFor[LXCStatusOutput]("vmid 파라미터가 필요합니다"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[LXCStatusOutput]("노드 조회 실패: %s", err.Error()), nil
		}

		ct, err := node.Container(ctx, input.VMID)
		if err != nil {
			return toolErrorfFor[LXCStatusOutput]("LXC 조회 실패: %s", err.Error()), nil
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
			return toolErrorFor[LXCConfigOutput]("node 파라미터가 필요합니다"), nil
		}
		if input.VMID == 0 {
			return toolErrorFor[LXCConfigOutput]("vmid 파라미터가 필요합니다"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[LXCConfigOutput]("노드 조회 실패: %s", err.Error()), nil
		}

		ct, err := node.Container(ctx, input.VMID)
		if err != nil {
			return toolErrorfFor[LXCConfigOutput]("LXC 조회 실패: %s", err.Error()), nil
		}

		// Config method returns Task and error, but we can access ContainerConfig directly
		config := ct.ContainerConfig
		if config == nil {
			return toolErrorFor[LXCConfigOutput]("LXC 설정을 가져올 수 없습니다"), nil
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
		Description: "특정 노드의 모든 LXC 컨테이너 목록을 조회합니다. 각 컨테이너의 상태, CPU, 메모리, 디스크 사용량 정보를 포함합니다.",
	}, createListLXCHandler(client))

	// pve_lxc_status
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_lxc_status",
		Description: "특정 LXC 컨테이너의 상세 상태를 조회합니다. 실행 상태, 리소스 사용량, 네트워크 I/O 등의 정보를 제공합니다.",
	}, createLXCStatusHandler(client))

	// pve_lxc_config
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_lxc_config",
		Description: "특정 LXC 컨테이너의 설정을 조회합니다. 호스트명, CPU, 메모리, 스토리지, 네트워크 설정 등의 정보를 제공합니다.",
	}, createLXCConfigHandler(client))

	fmt.Println("LXC tools registered: pve_list_lxc, pve_lxc_status, pve_lxc_config")
}
