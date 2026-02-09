package opnsense

import (
	"context"
	"encoding/json"
	"fmt"

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
			return toolErrorfFor[ListDHCPLeasesOutput]("DHCP 리스 조회 실패: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ListDHCPLeasesOutput]("응답 파싱 실패: %s", err.Error()), nil
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
		Description: "OPNsense DHCP 서버의 현재 임대 목록을 조회합니다. IP, MAC, 호스트명 등의 정보를 제공합니다.",
	}, createListDHCPLeasesHandler(client))

	fmt.Println("OPNsense DHCP tools registered: ops_list_dhcp_leases")
}
