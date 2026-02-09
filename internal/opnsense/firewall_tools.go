package opnsense

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

type ListFirewallRulesInput struct{}

type ListFirewallRulesOutput struct {
	Rules    []map[string]any `json:"rules"`
	Total    int              `json:"total"`
	RowCount int              `json:"rowCount"`
}

type GetFirewallRuleInput struct {
	UUID string `json:"uuid" jsonschema:"방화벽 규칙 UUID"`
}

type GetFirewallRuleOutput struct {
	Rule map[string]any `json:"rule"`
}

type AddFirewallRuleInput struct {
	Enabled     string `json:"enabled,omitempty" jsonschema:"활성화 여부 (1 또는 0, 기본값: 1)"`
	Action      string `json:"action" jsonschema:"액션 (pass, block, reject)"`
	Direction   string `json:"direction,omitempty" jsonschema:"방향 (in, out, 기본값: in)"`
	IPProtocol  string `json:"ipprotocol,omitempty" jsonschema:"IP 프로토콜 (inet, inet6, 기본값: inet)"`
	Protocol    string `json:"protocol,omitempty" jsonschema:"프로토콜 (TCP, UDP, ICMP 등)"`
	Interface   string `json:"interface,omitempty" jsonschema:"인터페이스 (lan, wan, opt1 등)"`
	Source      string `json:"source_net,omitempty" jsonschema:"출발지 (any, IP/CIDR, alias)"`
	Destination string `json:"destination_net,omitempty" jsonschema:"목적지 (any, IP/CIDR, alias)"`
	DstPort     string `json:"destination_port,omitempty" jsonschema:"목적지 포트 (예: 80, 443, 8080-8090)"`
	Description string `json:"description,omitempty" jsonschema:"규칙 설명"`
}

type AddFirewallRuleOutput struct {
	Result string `json:"result"`
	UUID   string `json:"uuid"`
}

type ToggleFirewallRuleInput struct {
	UUID    string `json:"uuid" jsonschema:"방화벽 규칙 UUID"`
	Enabled string `json:"enabled,omitempty" jsonschema:"활성화 여부 (1 또는 0, 비어있으면 토글)"`
}

type ToggleFirewallRuleOutput struct {
	Result string `json:"result"`
}

type DeleteFirewallRuleInput struct {
	UUID string `json:"uuid" jsonschema:"삭제할 방화벽 규칙 UUID"`
}

type DeleteFirewallRuleOutput struct {
	Result string `json:"result"`
}

// === Tool Handlers ===

func createListFirewallRulesHandler(client *Client) mcp.ToolHandlerFor[ListFirewallRulesInput, ListFirewallRulesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListFirewallRulesInput]) (*mcp.CallToolResultFor[ListFirewallRulesOutput], error) {
		data, err := client.Get(ctx, "firewall/filter/searchRule")
		if err != nil {
			return toolErrorfFor[ListFirewallRulesOutput]("방화벽 규칙 조회 실패: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ListFirewallRulesOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(ListFirewallRulesOutput{
			Rules:    result.Rows,
			Total:    result.Total,
			RowCount: result.RowCount,
		}), nil
	}
}

func createGetFirewallRuleHandler(client *Client) mcp.ToolHandlerFor[GetFirewallRuleInput, GetFirewallRuleOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[GetFirewallRuleInput]) (*mcp.CallToolResultFor[GetFirewallRuleOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[GetFirewallRuleOutput]("uuid 파라미터가 필요합니다"), nil
		}

		data, err := client.Get(ctx, "firewall/filter/getRule/"+input.UUID)
		if err != nil {
			return toolErrorfFor[GetFirewallRuleOutput]("규칙 조회 실패: %s", err.Error()), nil
		}

		var rule map[string]any
		if err := json.Unmarshal(data, &rule); err != nil {
			return toolErrorfFor[GetFirewallRuleOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(GetFirewallRuleOutput{Rule: rule}), nil
	}
}

func createAddFirewallRuleHandler(client *Client) mcp.ToolHandlerFor[AddFirewallRuleInput, AddFirewallRuleOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[AddFirewallRuleInput]) (*mcp.CallToolResultFor[AddFirewallRuleOutput], error) {
		input := params.Arguments
		if input.Action == "" {
			return toolErrorFor[AddFirewallRuleOutput]("action 파라미터가 필요합니다 (pass, block, reject)"), nil
		}

		// Set defaults
		if input.Enabled == "" {
			input.Enabled = "1"
		}
		if input.Direction == "" {
			input.Direction = "in"
		}
		if input.IPProtocol == "" {
			input.IPProtocol = "inet"
		}

		body := map[string]any{
			"rule": map[string]any{
				"enabled":         input.Enabled,
				"action":          input.Action,
				"direction":       input.Direction,
				"ipprotocol":      input.IPProtocol,
				"protocol":        input.Protocol,
				"interface":       input.Interface,
				"source_net":      input.Source,
				"destination_net": input.Destination,
				"destination_port": input.DstPort,
				"description":     input.Description,
			},
		}

		data, err := client.Post(ctx, "firewall/filter/addRule", body)
		if err != nil {
			return toolErrorfFor[AddFirewallRuleOutput]("규칙 추가 실패: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[AddFirewallRuleOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(AddFirewallRuleOutput{
			Result: result.Result,
			UUID:   result.UUID,
		}), nil
	}
}

func createToggleFirewallRuleHandler(client *Client) mcp.ToolHandlerFor[ToggleFirewallRuleInput, ToggleFirewallRuleOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ToggleFirewallRuleInput]) (*mcp.CallToolResultFor[ToggleFirewallRuleOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[ToggleFirewallRuleOutput]("uuid 파라미터가 필요합니다"), nil
		}

		path := "firewall/filter/toggleRule/" + input.UUID
		if input.Enabled != "" {
			path += "/" + input.Enabled
		}

		data, err := client.Post(ctx, path, nil)
		if err != nil {
			return toolErrorfFor[ToggleFirewallRuleOutput]("규칙 토글 실패: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ToggleFirewallRuleOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(ToggleFirewallRuleOutput{Result: result.Result}), nil
	}
}

func createDeleteFirewallRuleHandler(client *Client) mcp.ToolHandlerFor[DeleteFirewallRuleInput, DeleteFirewallRuleOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[DeleteFirewallRuleInput]) (*mcp.CallToolResultFor[DeleteFirewallRuleOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[DeleteFirewallRuleOutput]("uuid 파라미터가 필요합니다"), nil
		}

		data, err := client.Post(ctx, "firewall/filter/delRule/"+input.UUID, nil)
		if err != nil {
			return toolErrorfFor[DeleteFirewallRuleOutput]("규칙 삭제 실패: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[DeleteFirewallRuleOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(DeleteFirewallRuleOutput{Result: result.Result}), nil
	}
}

// === Registration ===

func registerFirewallTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_list_firewall_rules",
		Description: "OPNsense 방화벽 필터 규칙 목록을 조회합니다.",
	}, createListFirewallRulesHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_get_firewall_rule",
		Description: "특정 방화벽 규칙의 상세 설정을 조회합니다.",
	}, createGetFirewallRuleHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_add_firewall_rule",
		Description: "새 방화벽 필터 규칙을 추가합니다. 추가 후 ops_apply_changes(module=firewall)를 호출해야 적용됩니다.",
	}, createAddFirewallRuleHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_toggle_firewall_rule",
		Description: "방화벽 규칙을 활성화/비활성화합니다. 변경 후 ops_apply_changes(module=firewall)를 호출해야 적용됩니다.",
	}, createToggleFirewallRuleHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_delete_firewall_rule",
		Description: "방화벽 규칙을 삭제합니다. 삭제 후 ops_apply_changes(module=firewall)를 호출해야 적용됩니다.",
	}, createDeleteFirewallRuleHandler(client))

	fmt.Println("OPNsense firewall tools registered: ops_list_firewall_rules, ops_get_firewall_rule, ops_add_firewall_rule, ops_toggle_firewall_rule, ops_delete_firewall_rule")
}
