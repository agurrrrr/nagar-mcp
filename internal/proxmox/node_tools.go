package proxmox

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

// ListNodesInput - no input required
type ListNodesInput struct{}

// ListNodesOutput contains the list of nodes
type ListNodesOutput struct {
	Nodes []NodeInfo `json:"nodes"`
}

// NodeStatusInput contains the node name
type NodeStatusInput struct {
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
}

// NodeStatusOutput contains detailed node status
type NodeStatusOutput struct {
	Node          string  `json:"node"`
	CPU           float64 `json:"cpu"`
	Uptime        int64   `json:"uptime"`
	KernelVersion string  `json:"kernel_version,omitempty"`
	PVEVersion    string  `json:"pve_version,omitempty"`
}

// === Tool Handlers ===

func createListNodesHandler(client *Client) mcp.ToolHandlerFor[ListNodesInput, ListNodesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListNodesInput]) (*mcp.CallToolResultFor[ListNodesOutput], error) {
		nodes, err := client.Nodes(ctx)
		if err != nil {
			return toolErrorfFor[ListNodesOutput]("Proxmox API 오류: %s", err.Error()), nil
		}

		output := ListNodesOutput{
			Nodes: make([]NodeInfo, 0, len(nodes)),
		}

		for _, n := range nodes {
			output.Nodes = append(output.Nodes, NodeInfo{
				Node:    n.Node,
				Status:  n.Status,
				CPU:     n.CPU,
				MaxCPU:  int(n.MaxCPU),
				Mem:     int64(n.Mem),
				MaxMem:  int64(n.MaxMem),
				Disk:    int64(n.Disk),
				MaxDisk: int64(n.MaxDisk),
				Uptime:  int64(n.Uptime),
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createNodeStatusHandler(client *Client) mcp.ToolHandlerFor[NodeStatusInput, NodeStatusOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[NodeStatusInput]) (*mcp.CallToolResultFor[NodeStatusOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return toolErrorFor[NodeStatusOutput]("node 파라미터가 필요합니다"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[NodeStatusOutput]("노드 조회 실패: %s", err.Error()), nil
		}

		output := NodeStatusOutput{
			Node:          node.Name,
			CPU:           node.CPU,
			Uptime:        int64(node.Uptime),
			KernelVersion: node.Kversion,
			PVEVersion:    node.PVEVersion,
		}

		return toolSuccessFor(output), nil
	}
}

// === Registration ===

func registerNodeTools(server *mcp.Server, client *Client) {
	// pve_list_nodes
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_list_nodes",
		Description: "Proxmox 클러스터의 모든 노드 목록과 상태를 조회합니다. CPU, 메모리, 디스크 사용량, 가동시간 등의 정보를 포함합니다.",
	}, createListNodesHandler(client))

	// pve_node_status
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_node_status",
		Description: "특정 Proxmox 노드의 상세 상태를 조회합니다. CPU, 메모리, 디스크 사용량 및 가동시간 정보를 제공합니다.",
	}, createNodeStatusHandler(client))

	fmt.Println("Node tools registered: pve_list_nodes, pve_node_status")
}
