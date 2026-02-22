package proxmox

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

// ClusterResourcesInput contains optional type filter
type ClusterResourcesInput struct {
	Type string `json:"type,omitempty" jsonschema:"Resource type filter (vm|storage|node),enum=vm,enum=storage,enum=node"`
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
	Node string `json:"node" jsonschema:"Proxmox node name"`
}

// ListStorageOutput contains storage list
type ListStorageOutput struct {
	Storages []StorageInfo `json:"storages"`
}

// StorageContentInput contains node and storage name
type StorageContentInput struct {
	Node    string `json:"node" jsonschema:"Proxmox node name"`
	Storage string `json:"storage" jsonschema:"Storage name"`
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
			return toolErrorfFor[ClusterResourcesOutput]("failed to query cluster: %s", err.Error()), nil
		}

		resources, err := cluster.Resources(ctx, input.Type)
		if err != nil {
			return toolErrorfFor[ClusterResourcesOutput]("failed to list resources: %s", err.Error()), nil
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
			return toolErrorfFor[ClusterStatusOutput]("failed to query cluster: %s", err.Error()), nil
		}

		// Status() populates the cluster's fields including Nodes
		err = cluster.Status(ctx)
		if err != nil {
			return toolErrorfFor[ClusterStatusOutput]("failed to query cluster status: %s", err.Error()), nil
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
			return toolErrorFor[ListStorageOutput]("node parameter is required"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[ListStorageOutput]("failed to query node: %s", err.Error()), nil
		}

		storages, err := node.Storages(ctx)
		if err != nil {
			return toolErrorfFor[ListStorageOutput]("failed to list storage: %s", err.Error()), nil
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
			return toolErrorFor[StorageContentOutput]("node parameter is required"), nil
		}
		if input.Storage == "" {
			return toolErrorFor[StorageContentOutput]("storage parameter is required"), nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[StorageContentOutput]("failed to query node: %s", err.Error()), nil
		}

		storage, err := node.Storage(ctx, input.Storage)
		if err != nil {
			return toolErrorfFor[StorageContentOutput]("failed to query storage: %s", err.Error()), nil
		}

		content, err := storage.GetContent(ctx)
		if err != nil {
			return toolErrorfFor[StorageContentOutput]("failed to query storage content: %s", err.Error()), nil
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
		Description: "Lists all resources (VMs, LXC, storage, nodes) in the Proxmox cluster. Can be filtered by the type parameter.",
	}, createClusterResourcesHandler(client))

	// pve_cluster_status
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_cluster_status",
		Description: "Retrieves Proxmox cluster status. Provides cluster name, quorum status, node list, and online status of each node.",
	}, createClusterStatusHandler(client))

	// pve_list_storage
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_list_storage",
		Description: "Lists storage on a specific node. Includes type, capacity, and usage information for each storage.",
	}, createListStorageHandler(client))

	// pve_storage_content
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_storage_content",
		Description: "Lists content of a specific storage (ISOs, templates, disk images, etc.).",
	}, createStorageContentHandler(client))

	slog.Info("Cluster/Storage tools registered: pve_cluster_resources, pve_cluster_status, pve_list_storage, pve_storage_content")
}
