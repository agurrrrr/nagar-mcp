package opnsense

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

type ListServicesInput struct{}

type ListServicesOutput struct {
	Services []map[string]any `json:"services"`
	Total    int              `json:"total"`
	RowCount int              `json:"rowCount"`
}

type ServiceActionInput struct {
	Service string `json:"service" jsonschema:"서비스 이름 (예: unbound, wireguard, haproxy)"`
	Action  string `json:"action" jsonschema:"액션 (start, stop, restart, status)"`
}

type ServiceActionOutput struct {
	Result map[string]any `json:"result"`
}

type ApplyChangesInput struct {
	Module string `json:"module" jsonschema:"모듈 이름 (예: firewall, unbound, wireguard, haproxy)"`
}

type ApplyChangesOutput struct {
	Result map[string]any `json:"result"`
}

// === Tool Handlers ===

func createListServicesHandler(client *Client) mcp.ToolHandlerFor[ListServicesInput, ListServicesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListServicesInput]) (*mcp.CallToolResultFor[ListServicesOutput], error) {
		data, err := client.Get(ctx, "core/service/search")
		if err != nil {
			return toolErrorfFor[ListServicesOutput]("서비스 목록 조회 실패: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ListServicesOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(ListServicesOutput{
			Services: result.Rows,
			Total:    result.Total,
			RowCount: result.RowCount,
		}), nil
	}
}

func createServiceActionHandler(client *Client) mcp.ToolHandlerFor[ServiceActionInput, ServiceActionOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ServiceActionInput]) (*mcp.CallToolResultFor[ServiceActionOutput], error) {
		input := params.Arguments
		if input.Service == "" {
			return toolErrorFor[ServiceActionOutput]("service 파라미터가 필요합니다"), nil
		}
		if input.Action == "" {
			return toolErrorFor[ServiceActionOutput]("action 파라미터가 필요합니다 (start, stop, restart, status)"), nil
		}

		// Map service names to API module paths
		moduleMap := map[string]string{
			"unbound":   "unbound",
			"dns":       "unbound",
			"wireguard": "wireguard",
			"haproxy":   "haproxy",
			"openvpn":   "openvpn",
		}

		module, ok := moduleMap[input.Service]
		if !ok {
			module = input.Service
		}

		var (
			data json.RawMessage
			err  error
		)

		if input.Action == "status" {
			data, err = client.Get(ctx, module+"/service/status")
		} else {
			data, err = client.Post(ctx, module+"/service/"+input.Action, nil)
		}
		if err != nil {
			return toolErrorfFor[ServiceActionOutput]("서비스 %s %s 실패: %s", input.Service, input.Action, err.Error()), nil
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ServiceActionOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(ServiceActionOutput{Result: result}), nil
	}
}

func createApplyChangesHandler(client *Client) mcp.ToolHandlerFor[ApplyChangesInput, ApplyChangesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ApplyChangesInput]) (*mcp.CallToolResultFor[ApplyChangesOutput], error) {
		input := params.Arguments
		if input.Module == "" {
			return toolErrorFor[ApplyChangesOutput]("module 파라미터가 필요합니다"), nil
		}

		// Map module names to reconfigure paths
		pathMap := map[string]string{
			"firewall":  "firewall/filter/apply",
			"unbound":   "unbound/service/reconfigure",
			"dns":       "unbound/service/reconfigure",
			"wireguard": "wireguard/service/reconfigure",
			"haproxy":   "haproxy/service/reconfigure",
		}

		path, ok := pathMap[input.Module]
		if !ok {
			path = input.Module + "/service/reconfigure"
		}

		data, err := client.Post(ctx, path, nil)
		if err != nil {
			return toolErrorfFor[ApplyChangesOutput]("변경사항 적용 실패: %s", err.Error()), nil
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ApplyChangesOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(ApplyChangesOutput{Result: result}), nil
	}
}

// === Registration ===

func registerServiceTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_list_services",
		Description: "OPNsense에서 실행 중인 서비스 목록과 상태를 조회합니다.",
	}, createListServicesHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_service_action",
		Description: "OPNsense 서비스를 제어합니다 (start, stop, restart, status).",
	}, createServiceActionHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_apply_changes",
		Description: "설정 변경사항을 적용합니다. 방화벽 규칙, DNS, WireGuard, HAProxy 등의 변경 후 반드시 호출해야 합니다.",
	}, createApplyChangesHandler(client))

	fmt.Println("OPNsense service tools registered: ops_list_services, ops_service_action, ops_apply_changes")
}
