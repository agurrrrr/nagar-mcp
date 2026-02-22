package opnsense

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

type ListDHCPLeasesInput struct{}

type ListDHCPLeasesOutput struct {
	Leases   []map[string]any `json:"leases"`
	Total    int              `json:"total"`
	RowCount int              `json:"rowCount"`
}

// === Tool Handlers ===

func createListDHCPLeasesHandler(client *Client) mcp.ToolHandlerFor[ListDHCPLeasesInput, ListDHCPLeasesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListDHCPLeasesInput]) (*mcp.CallToolResultFor[ListDHCPLeasesOutput], error) {
		data, err := client.Get(ctx, "dhcpv4/leases/searchLease")
		if err != nil {
			return toolErrorfFor[ListDHCPLeasesOutput]("failed to query DHCP leases: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ListDHCPLeasesOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(ListDHCPLeasesOutput{
			Leases:   result.Rows,
			Total:    result.Total,
			RowCount: result.RowCount,
		}), nil
	}
}

// === Registration ===

func registerDHCPTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_list_dhcp_leases",
		Description: "Lists current DHCP leases from the OPNsense DHCP server. Provides IP, MAC, hostname, and other information.",
	}, createListDHCPLeasesHandler(client))

	slog.Info("OPNsense DHCP tools registered: ops_list_dhcp_leases")
}
