package proxmox

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

// ClusterResourcesInput contains optional type filter
type ClusterResourcesInput struct {
	Type string `json:"type,omitempty" jsonschema:"리소스 타입 필터 (vm|storage|node),enum=vm,enum=storage,enum=node"`
}

// ClusterResourcesOutput contains cluster resources
type ClusterResourcesOutput struct {
	Resources []ClusterResourceInfo `json:"resources"`
}

// ClusterStatusInput - no input required
type ClusterStatusInput struct{}

// ClusterStatusOutput contains cluster status
type ClusterStatusOutput struct {
	Name    string       `json:"name"`
	Version int          `json:"version"`
	Quorate bool         `json:"quorate"`
	Nodes   []NodeStatus `json:"nodes"`
}

// NodeStatus represents a node in cluster status
type NodeStatus struct {
	Name   string `json:"name"`
	ID     int    `json:"id"`
	IP     string `json:"ip"`
	Online bool   `json:"online"`
	Local  bool   `json:"local"`
}

// ListStorageInput contains node name
type ListStorageInput struct {
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
}

// ListStorageOutput contains storage list
type ListStorageOutput struct {
	Storages []StorageInfo `json:"storages"`
}

// StorageContentInput contains node and storage name
type StorageContentInput struct {
	Node    string `json:"node" jsonschema:"Proxmox 노드 이름"`
	Storage string `json:"storage" jsonschema:"스토리지 이름"`
}

// StorageContentOutput contains storage content
type StorageContentOutput struct {
	Content []StorageContentItem `json:"content"`
}

// StorageContentItem represents a storage content item
type StorageContentItem struct {
	Volid  string `json:"volid"`
	Format string `json:"format"`
	Size   int64  `json:"size"`
	VMID   int    `json:"vmid,omitempty"`
	Notes  string `json:"notes,omitempty"`
}

// === Tool Handlers ===

func createClusterResourcesHandler(client *Client) mcp.ToolHandlerFor[ClusterResourcesInput, ClusterResourcesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ClusterResourcesInput]) (*mcp.CallToolResultFor[ClusterResourcesOutput], error) {
		input := params.Arguments

		cluster, err := client.Cluster(ctx)
		if err != nil {
			return toolErrorfFor[ClusterResourcesOutput]("클러스터 조회 실패: %s", err.Error()), nil
		}

		resources, err := cluster.Resources(ctx, input.Type)
		if err != nil {
			return toolErrorfFor[ClusterResourcesOutput]("리소스 목록 조회 실패: %s", err.Error()), nil
		}

		output := ClusterResourcesOutput{
			Resources: make([]ClusterResourceInfo, 0, len(resources)),
		}

		for _, r := range resources {
			output.Resources = append(output.Resources, ClusterResourceInfo{
				ID:      r.ID,
				Type:    r.Type,
				Node:    r.Node,
				Status:  r.Status,
				Name:    r.Name,
				VMID:    int(r.VMID),
				CPU:     r.CPU,
				MaxCPU:  int(r.MaxCPU),
				Mem:     int64(r.Mem),
				MaxMem:  int64(r.MaxMem),
				Disk:    int64(r.Disk),
				MaxDisk: int64(r.MaxDisk),
				Uptime:  int64(r.Uptime),
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createClusterStatusHandler(client *Client) mcp.ToolHandlerFor[ClusterStatusInput, ClusterStatusOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ClusterStatusInput]) (*mcp.CallToolResultFor[ClusterStatusOutput], error) {
		cluster, err := client.Cluster(ctx)
		if err != nil {
			return toolErrorfFor[ClusterStatusOutput]("클러스터 조회 실패: %s", err.Error()), nil
		}

		// Status() populates the cluster's fields including Nodes
		err = cluster.Status(ctx)
		if err != nil {
			return toolErrorfFor[ClusterStatusOutput]("클러스터 상태 조회 실패: %s", err.Error()), nil
		}

		output := ClusterStatusOutput{
			Name:    cluster.Name,
			Version: cluster.Version,
			Quorate: cluster.Quorate == 1,
			Nodes:   make([]NodeStatus, 0),
		}

		for _, item := range cluster.Nodes {
			output.Nodes = append(output.Nodes, NodeStatus{
				Name:   item.Name,
				ID:     item.NodeID,
				IP:     item.IP,
				Online: item.Online == 1,
				Local:  item.Local == 1,
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createListStorageHandler(client *Client) mcp.ToolHandlerFor[ListStorageInput, ListStorageOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListStorageInput]) (*mcp.CallToolResultFor[ListStorageOutput], error) {
		input := params.Arguments

		if input.Node == "" {
			return toolErrorFor[ListStorageOutput]("node 파라미터가 필요합니다"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[ListStorageOutput]("노드 조회 실패: %s", err.Error()), nil
		}

		storages, err := node.Storages(ctx)
		if err != nil {
			return toolErrorfFor[ListStorageOutput]("스토리지 목록 조회 실패: %s", err.Error()), nil
		}

		output := ListStorageOutput{
			Storages: make([]StorageInfo, 0, len(storages)),
		}

		for _, s := range storages {
			usedFraction := float64(0)
			if s.Total > 0 {
				usedFraction = float64(s.Used) / float64(s.Total)
			}

			output.Storages = append(output.Storages, StorageInfo{
				Storage:      s.Storage,
				Type:         s.Type,
				Content:      s.Content,
				Active:       int(s.Active),
				Enabled:      int(s.Enabled),
				Shared:       int(s.Shared),
				Used:         int64(s.Used),
				Avail:        int64(s.Avail),
				Total:        int64(s.Total),
				UsedFraction: usedFraction,
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createStorageContentHandler(client *Client) mcp.ToolHandlerFor[StorageContentInput, StorageContentOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[StorageContentInput]) (*mcp.CallToolResultFor[StorageContentOutput], error) {
		input := params.Arguments

		if input.Node == "" {
			return toolErrorFor[StorageContentOutput]("node 파라미터가 필요합니다"), nil
		}
		if input.Storage == "" {
			return toolErrorFor[StorageContentOutput]("storage 파라미터가 필요합니다"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[StorageContentOutput]("노드 조회 실패: %s", err.Error()), nil
		}

		storage, err := node.Storage(ctx, input.Storage)
		if err != nil {
			return toolErrorfFor[StorageContentOutput]("스토리지 조회 실패: %s", err.Error()), nil
		}

		content, err := storage.GetContent(ctx)
		if err != nil {
			return toolErrorfFor[StorageContentOutput]("스토리지 콘텐츠 조회 실패: %s", err.Error()), nil
		}

		output := StorageContentOutput{
			Content: make([]StorageContentItem, 0, len(content)),
		}

		for _, c := range content {
			output.Content = append(output.Content, StorageContentItem{
				Volid:  c.Volid,
				Format: c.Format,
				Size:   int64(c.Size),
				VMID:   int(c.VMID),
				Notes:  c.Notes,
			})
		}

		return toolSuccessFor(output), nil
	}
}

// === Registration ===

func registerClusterTools(server *mcp.Server, client *Client) {
	// pve_cluster_resources
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_cluster_resources",
		Description: "Proxmox 클러스터의 전체 리소스(VM, LXC, 스토리지, 노드)를 통합 조회합니다. type 파라미터로 필터링할 수 있습니다.",
	}, createClusterResourcesHandler(client))

	// pve_cluster_status
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_cluster_status",
		Description: "Proxmox 클러스터의 상태를 조회합니다. 클러스터 이름, 쿼럼 상태, 노드 목록 및 각 노드의 온라인 상태를 제공합니다.",
	}, createClusterStatusHandler(client))

	// pve_list_storage
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_list_storage",
		Description: "특정 노드의 스토리지 목록을 조회합니다. 각 스토리지의 타입, 용량, 사용량 정보를 포함합니다.",
	}, createListStorageHandler(client))

	// pve_storage_content
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_storage_content",
		Description: "특정 스토리지의 콘텐츠(ISO, 템플릿, 디스크 이미지 등)를 조회합니다.",
	}, createStorageContentHandler(client))

	fmt.Println("Cluster/Storage tools registered: pve_cluster_resources, pve_cluster_status, pve_list_storage, pve_storage_content")
}
