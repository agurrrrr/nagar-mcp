package opnsense

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

type ListDNSOverridesInput struct{}

type ListDNSOverridesOutput struct {
	Overrides []map[string]any `json:"overrides"`
	Total     int              `json:"total"`
	RowCount  int              `json:"rowCount"`
}

type AddDNSOverrideInput struct {
	Hostname string `json:"hostname" jsonschema:"호스트명 (예: myserver)"`
	Domain   string `json:"domain,omitempty" jsonschema:"도메인 (예: local)"`
	Server   string `json:"server" jsonschema:"IP 주소 (예: 192.168.1.100)"`
	Descr    string `json:"description,omitempty" jsonschema:"설명"`
}

type AddDNSOverrideOutput struct {
	Result string `json:"result"`
	UUID   string `json:"uuid"`
}

type DeleteDNSOverrideInput struct {
	UUID string `json:"uuid" jsonschema:"삭제할 DNS 오버라이드 UUID"`
}

type DeleteDNSOverrideOutput struct {
	Result string `json:"result"`
}

type DNSStatsInput struct{}

type DNSStatsOutput struct {
	Stats map[string]any `json:"stats"`
}

// === Tool Handlers ===

func createListDNSOverridesHandler(client *Client) mcp.ToolHandlerFor[ListDNSOverridesInput, ListDNSOverridesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListDNSOverridesInput]) (*mcp.CallToolResultFor[ListDNSOverridesOutput], error) {
		data, err := client.Get(ctx, "unbound/settings/searchHostOverride")
		if err != nil {
			return toolErrorfFor[ListDNSOverridesOutput]("DNS 오버라이드 조회 실패: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ListDNSOverridesOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(ListDNSOverridesOutput{
			Overrides: result.Rows,
			Total:     result.Total,
			RowCount:  result.RowCount,
		}), nil
	}
}

func createAddDNSOverrideHandler(client *Client) mcp.ToolHandlerFor[AddDNSOverrideInput, AddDNSOverrideOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[AddDNSOverrideInput]) (*mcp.CallToolResultFor[AddDNSOverrideOutput], error) {
		input := params.Arguments
		if input.Hostname == "" {
			return toolErrorFor[AddDNSOverrideOutput]("hostname 파라미터가 필요합니다"), nil
		}
		if input.Server == "" {
			return toolErrorFor[AddDNSOverrideOutput]("server (IP 주소) 파라미터가 필요합니다"), nil
		}

		body := map[string]any{
			"host": map[string]any{
				"enabled":     "1",
				"hostname":    input.Hostname,
				"domain":      input.Domain,
				"server":      input.Server,
				"description": input.Descr,
			},
		}

		data, err := client.Post(ctx, "unbound/settings/addHostOverride", body)
		if err != nil {
			return toolErrorfFor[AddDNSOverrideOutput]("DNS 오버라이드 추가 실패: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[AddDNSOverrideOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(AddDNSOverrideOutput{
			Result: result.Result,
			UUID:   result.UUID,
		}), nil
	}
}

func createDeleteDNSOverrideHandler(client *Client) mcp.ToolHandlerFor[DeleteDNSOverrideInput, DeleteDNSOverrideOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[DeleteDNSOverrideInput]) (*mcp.CallToolResultFor[DeleteDNSOverrideOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[DeleteDNSOverrideOutput]("uuid 파라미터가 필요합니다"), nil
		}

		data, err := client.Post(ctx, "unbound/settings/delHostOverride/"+input.UUID, nil)
		if err != nil {
			return toolErrorfFor[DeleteDNSOverrideOutput]("DNS 오버라이드 삭제 실패: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[DeleteDNSOverrideOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(DeleteDNSOverrideOutput{Result: result.Result}), nil
	}
}

func createDNSStatsHandler(client *Client) mcp.ToolHandlerFor[DNSStatsInput, DNSStatsOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[DNSStatsInput]) (*mcp.CallToolResultFor[DNSStatsOutput], error) {
		data, err := client.Get(ctx, "unbound/diagnostics/stats")
		if err != nil {
			return toolErrorfFor[DNSStatsOutput]("DNS 통계 조회 실패: %s", err.Error()), nil
		}

		var stats map[string]any
		if err := json.Unmarshal(data, &stats); err != nil {
			return toolErrorfFor[DNSStatsOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(DNSStatsOutput{Stats: stats}), nil
	}
}

// === Registration ===

func registerDNSTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_list_dns_overrides",
		Description: "OPNsense Unbound DNS 호스트 오버라이드 목록을 조회합니다.",
	}, createListDNSOverridesHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_add_dns_override",
		Description: "DNS 호스트 오버라이드를 추가합니다. 추가 후 ops_apply_changes(module=unbound)를 호출해야 적용됩니다.",
	}, createAddDNSOverrideHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_delete_dns_override",
		Description: "DNS 호스트 오버라이드를 삭제합니다. 삭제 후 ops_apply_changes(module=unbound)를 호출해야 적용됩니다.",
	}, createDeleteDNSOverrideHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_dns_stats",
		Description: "Unbound DNS 서버의 통계를 조회합니다. 쿼리 수, 캐시 적중률 등의 정보를 제공합니다.",
	}, createDNSStatsHandler(client))

	fmt.Println("OPNsense DNS tools registered: ops_list_dns_overrides, ops_add_dns_override, ops_delete_dns_override, ops_dns_stats")
}
