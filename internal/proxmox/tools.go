package proxmox

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterTools registers all Proxmox MCP tools
func RegisterTools(server *mcp.Server, client *Client) {
	// Node tools
	registerNodeTools(server, client)

	// VM tools (read)
	registerVMTools(server, client)

	// VM action tools (write)
	registerVMActionTools(server, client)

	// LXC tools (read)
	registerLXCTools(server, client)

	// LXC action tools (write)
	registerLXCActionTools(server, client)

	// Cluster/Storage tools
	registerClusterTools(server, client)

	// Task tools
	registerTaskTools(server, client)

	// Ceph tools
	registerCephTools(server, client)

	// Monitoring tools (RRD)
	registerMonitoringTools(server, client)
}
