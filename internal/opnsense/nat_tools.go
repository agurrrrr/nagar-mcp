package opnsense

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

type ListNATRulesInput struct{}

type ListNATRulesOutput struct {
	Rules    []map[string]any `json:"rules"`
	Total    int              `json:"total"`
	RowCount int              `json:"rowCount"`
}

type GetNATRuleInput struct {
	UUID string `json:"uuid" jsonschema:"NAT 규칙 UUID"`
}

type GetNATRuleOutput struct {
	Rule map[string]any `json:"rule"`
}

type AddNATRuleInput struct {
	Enabled     string `json:"enabled,omitempty" jsonschema:"활성화 여부 (1 또는 0, 기본값: 1)"`
	Interface   string `json:"interface,omitempty" jsonschema:"인터페이스 (wan, lan 등, 기본값: wan)"`
	IPProtocol  string `json:"ipprotocol,omitempty" jsonschema:"IP 프로토콜 (inet, inet6, 기본값: inet)"`
	Protocol    string `json:"protocol" jsonschema:"프로토콜 (TCP, UDP, TCP/UDP 등)"`
	Source      string `json:"source,omitempty" jsonschema:"출발지 (any, IP/CIDR)"`
	SourcePort  string `json:"source_port,omitempty" jsonschema:"출발지 포트"`
	Destination string `json:"destination,omitempty" jsonschema:"목적지 (any, WAN address 등)"`
	DstPort     string `json:"destination_port" jsonschema:"목적지 포트 (외부 포트, 예: 8080, 30000-30100)"`
	TargetIP    string `json:"target" jsonschema:"리다이렉트 대상 IP (내부 호스트)"`
	TargetPort  string `json:"target_port,omitempty" jsonschema:"리다이렉트 대상 포트 (비어있으면 목적지 포트와 동일)"`
	Description string `json:"description,omitempty" jsonschema:"규칙 설명"`
}

type AddNATRuleOutput struct {
	Result string `json:"result"`
	UUID   string `json:"uuid"`
}

type UpdateNATRuleInput struct {
	UUID        string `json:"uuid" jsonschema:"수정할 NAT 규칙 UUID"`
	Enabled     string `json:"enabled,omitempty" jsonschema:"활성화 여부 (1 또는 0)"`
	Interface   string `json:"interface,omitempty" jsonschema:"인터페이스"`
	IPProtocol  string `json:"ipprotocol,omitempty" jsonschema:"IP 프로토콜"`
	Protocol    string `json:"protocol,omitempty" jsonschema:"프로토콜"`
	Source      string `json:"source,omitempty" jsonschema:"출발지"`
	SourcePort  string `json:"source_port,omitempty" jsonschema:"출발지 포트"`
	Destination string `json:"destination,omitempty" jsonschema:"목적지"`
	DstPort     string `json:"destination_port,omitempty" jsonschema:"목적지 포트"`
	TargetIP    string `json:"target,omitempty" jsonschema:"리다이렉트 대상 IP"`
	TargetPort  string `json:"target_port,omitempty" jsonschema:"리다이렉트 대상 포트"`
	Description string `json:"description,omitempty" jsonschema:"규칙 설명"`
}

type UpdateNATRuleOutput struct {
	Result string `json:"result"`
}

type DeleteNATRuleInput struct {
	UUID string `json:"uuid" jsonschema:"삭제할 NAT 규칙 UUID"`
}

type DeleteNATRuleOutput struct {
	Result string `json:"result"`
}

type ToggleNATRuleInput struct {
	UUID     string `json:"uuid" jsonschema:"NAT 규칙 UUID"`
	Disabled string `json:"disabled,omitempty" jsonschema:"비활성화 여부 (1 또는 0, 비어있으면 토글)"`
}

type ToggleNATRuleOutput struct {
	Result string `json:"result"`
}

// === Tool Handlers ===

func createListNATRulesHandler(client *Client) mcp.ToolHandlerFor[ListNATRulesInput, ListNATRulesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListNATRulesInput]) (*mcp.CallToolResultFor[ListNATRulesOutput], error) {
		data, err := client.Get(ctx, "firewall/d_nat/search_rule")
		if err != nil {
			return toolErrorfFor[ListNATRulesOutput]("NAT 규칙 조회 실패: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ListNATRulesOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(ListNATRulesOutput{
			Rules:    result.Rows,
			Total:    result.Total,
			RowCount: result.RowCount,
		}), nil
	}
}

func createGetNATRuleHandler(client *Client) mcp.ToolHandlerFor[GetNATRuleInput, GetNATRuleOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[GetNATRuleInput]) (*mcp.CallToolResultFor[GetNATRuleOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[GetNATRuleOutput]("uuid 파라미터가 필요합니다"), nil
		}

		data, err := client.Get(ctx, "firewall/d_nat/get_rule/"+input.UUID)
		if err != nil {
			return toolErrorfFor[GetNATRuleOutput]("NAT 규칙 조회 실패: %s", err.Error()), nil
		}

		var rule map[string]any
		if err := json.Unmarshal(data, &rule); err != nil {
			return toolErrorfFor[GetNATRuleOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(GetNATRuleOutput{Rule: rule}), nil
	}
}

func createAddNATRuleHandler(client *Client) mcp.ToolHandlerFor[AddNATRuleInput, AddNATRuleOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[AddNATRuleInput]) (*mcp.CallToolResultFor[AddNATRuleOutput], error) {
		input := params.Arguments
		if input.Protocol == "" {
			return toolErrorFor[AddNATRuleOutput]("protocol 파라미터가 필요합니다 (TCP, UDP, TCP/UDP)"), nil
		}
		if input.DstPort == "" {
			return toolErrorFor[AddNATRuleOutput]("destination_port (외부 포트) 파라미터가 필요합니다"), nil
		}
		if input.TargetIP == "" {
			return toolErrorFor[AddNATRuleOutput]("target (리다이렉트 대상 IP) 파라미터가 필요합니다"), nil
		}

		if input.Enabled == "" {
			input.Enabled = "1"
		}
		if input.Interface == "" {
			input.Interface = "wan"
		}
		if input.IPProtocol == "" {
			input.IPProtocol = "inet"
		}

		rule := map[string]any{
			"disabled":         boolInvert(input.Enabled),
			"interface":        input.Interface,
			"ipprotocol":       input.IPProtocol,
			"protocol":         input.Protocol,
			"source":           input.Source,
			"source_port":      input.SourcePort,
			"destination":      input.Destination,
			"destination_port": input.DstPort,
			"target":           input.TargetIP,
			"target_port":      input.TargetPort,
			"description":      input.Description,
		}

		body := map[string]any{"rule": rule}

		data, err := client.Post(ctx, "firewall/d_nat/add_rule", body)
		if err != nil {
			return toolErrorfFor[AddNATRuleOutput]("NAT 규칙 추가 실패: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[AddNATRuleOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(AddNATRuleOutput{
			Result: result.Result,
			UUID:   result.UUID,
		}), nil
	}
}

func createUpdateNATRuleHandler(client *Client) mcp.ToolHandlerFor[UpdateNATRuleInput, UpdateNATRuleOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[UpdateNATRuleInput]) (*mcp.CallToolResultFor[UpdateNATRuleOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[UpdateNATRuleOutput]("uuid 파라미터가 필요합니다"), nil
		}

		rule := make(map[string]any)
		if input.Enabled != "" {
			rule["disabled"] = boolInvert(input.Enabled)
		}
		if input.Interface != "" {
			rule["interface"] = input.Interface
		}
		if input.IPProtocol != "" {
			rule["ipprotocol"] = input.IPProtocol
		}
		if input.Protocol != "" {
			rule["protocol"] = input.Protocol
		}
		if input.Source != "" {
			rule["source"] = input.Source
		}
		if input.SourcePort != "" {
			rule["source_port"] = input.SourcePort
		}
		if input.Destination != "" {
			rule["destination"] = input.Destination
		}
		if input.DstPort != "" {
			rule["destination_port"] = input.DstPort
		}
		if input.TargetIP != "" {
			rule["target"] = input.TargetIP
		}
		if input.TargetPort != "" {
			rule["target_port"] = input.TargetPort
		}
		if input.Description != "" {
			rule["description"] = input.Description
		}

		body := map[string]any{"rule": rule}

		data, err := client.Post(ctx, "firewall/d_nat/set_rule/"+input.UUID, body)
		if err != nil {
			return toolErrorfFor[UpdateNATRuleOutput]("NAT 규칙 수정 실패: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[UpdateNATRuleOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(UpdateNATRuleOutput{Result: result.Result}), nil
	}
}

func createDeleteNATRuleHandler(client *Client) mcp.ToolHandlerFor[DeleteNATRuleInput, DeleteNATRuleOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[DeleteNATRuleInput]) (*mcp.CallToolResultFor[DeleteNATRuleOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[DeleteNATRuleOutput]("uuid 파라미터가 필요합니다"), nil
		}

		data, err := client.Post(ctx, "firewall/d_nat/del_rule/"+input.UUID, nil)
		if err != nil {
			return toolErrorfFor[DeleteNATRuleOutput]("NAT 규칙 삭제 실패: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[DeleteNATRuleOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(DeleteNATRuleOutput{Result: result.Result}), nil
	}
}

func createToggleNATRuleHandler(client *Client) mcp.ToolHandlerFor[ToggleNATRuleInput, ToggleNATRuleOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ToggleNATRuleInput]) (*mcp.CallToolResultFor[ToggleNATRuleOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[ToggleNATRuleOutput]("uuid 파라미터가 필요합니다"), nil
		}

		path := "firewall/d_nat/toggle_rule/" + input.UUID
		if input.Disabled != "" {
			path += "/" + input.Disabled
		}

		data, err := client.Post(ctx, path, nil)
		if err != nil {
			return toolErrorfFor[ToggleNATRuleOutput]("NAT 규칙 토글 실패: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ToggleNATRuleOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(ToggleNATRuleOutput{Result: result.Result}), nil
	}
}

// boolInvert converts enabled "1"/"0" to disabled "0"/"1"
func boolInvert(enabled string) string {
	if enabled == "1" {
		return "0"
	}
	return "1"
}

// === Registration ===

func registerNATTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_list_nat_rules",
		Description: "OPNsense NAT 포트포워딩(DNAT) 규칙 목록을 조회합니다.",
	}, createListNATRulesHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_get_nat_rule",
		Description: "특정 NAT 포트포워딩 규칙의 상세 설정을 조회합니다.",
	}, createGetNATRuleHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_add_nat_rule",
		Description: "NAT 포트포워딩 규칙을 추가합니다. 외부 포트를 내부 IP:포트로 리다이렉트합니다. 추가 후 ops_apply_changes(module=firewall)를 호출해야 적용됩니다.",
	}, createAddNATRuleHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_update_nat_rule",
		Description: "기존 NAT 포트포워딩 규칙을 수정합니다. target IP/포트 변경 등에 사용합니다. 수정 후 ops_apply_changes(module=firewall)를 호출해야 적용됩니다.",
	}, createUpdateNATRuleHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_delete_nat_rule",
		Description: "NAT 포트포워딩 규칙을 삭제합니다. 삭제 후 ops_apply_changes(module=firewall)를 호출해야 적용됩니다.",
	}, createDeleteNATRuleHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_toggle_nat_rule",
		Description: "NAT 포트포워딩 규칙을 활성화/비활성화합니다. 변경 후 ops_apply_changes(module=firewall)를 호출해야 적용됩니다.",
	}, createToggleNATRuleHandler(client))

	fmt.Println("OPNsense NAT tools registered: ops_list_nat_rules, ops_get_nat_rule, ops_add_nat_rule, ops_update_nat_rule, ops_delete_nat_rule, ops_toggle_nat_rule")
}
