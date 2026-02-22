package opnsense

import (
	"context"
	"encoding/json"
	"log/slog"

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
	UUID string `json:"uuid" jsonschema:"NAT rule UUID"`
}

type GetNATRuleOutput struct {
	Rule map[string]any `json:"rule"`
}

type AddNATRuleInput struct {
	Enabled     string `json:"enabled,omitempty" jsonschema:"Enabled (1 or 0, default: 1)"`
	Interface   string `json:"interface,omitempty" jsonschema:"Interface (wan, lan, etc., default: wan)"`
	IPProtocol  string `json:"ipprotocol,omitempty" jsonschema:"IP protocol (inet, inet6, default: inet)"`
	Protocol    string `json:"protocol" jsonschema:"Protocol (TCP, UDP, TCP/UDP, etc.)"`
	Source      string `json:"source,omitempty" jsonschema:"Source (any, IP/CIDR)"`
	SourcePort  string `json:"source_port,omitempty" jsonschema:"Source port"`
	Destination string `json:"destination,omitempty" jsonschema:"Destination (any, WAN address, etc.)"`
	DstPort     string `json:"destination_port" jsonschema:"Destination port (external port, e.g. 8080, 30000-30100)"`
	TargetIP    string `json:"target" jsonschema:"Redirect target IP (internal host)"`
	TargetPort  string `json:"target_port,omitempty" jsonschema:"Redirect target port (same as destination port if empty)"`
	Description string `json:"description,omitempty" jsonschema:"Rule description"`
}

type AddNATRuleOutput struct {
	Result string `json:"result"`
	UUID   string `json:"uuid"`
}

type UpdateNATRuleInput struct {
	UUID        string `json:"uuid" jsonschema:"NAT rule UUID to update"`
	Enabled     string `json:"enabled,omitempty" jsonschema:"Enabled (1 or 0)"`
	Interface   string `json:"interface,omitempty" jsonschema:"Interface"`
	IPProtocol  string `json:"ipprotocol,omitempty" jsonschema:"IP protocol"`
	Protocol    string `json:"protocol,omitempty" jsonschema:"Protocol"`
	Source      string `json:"source,omitempty" jsonschema:"Source"`
	SourcePort  string `json:"source_port,omitempty" jsonschema:"Source port"`
	Destination string `json:"destination,omitempty" jsonschema:"Destination"`
	DstPort     string `json:"destination_port,omitempty" jsonschema:"Destination port"`
	TargetIP    string `json:"target,omitempty" jsonschema:"Redirect target IP"`
	TargetPort  string `json:"target_port,omitempty" jsonschema:"Redirect target port"`
	Description string `json:"description,omitempty" jsonschema:"Rule description"`
}

type UpdateNATRuleOutput struct {
	Result string `json:"result"`
}

type DeleteNATRuleInput struct {
	UUID string `json:"uuid" jsonschema:"NAT rule UUID to delete"`
}

type DeleteNATRuleOutput struct {
	Result string `json:"result"`
}

type ToggleNATRuleInput struct {
	UUID     string `json:"uuid" jsonschema:"NAT rule UUID"`
	Disabled string `json:"disabled,omitempty" jsonschema:"Disabled (1 or 0, toggles if empty)"`
}

type ToggleNATRuleOutput struct {
	Result string `json:"result"`
}

// === Tool Handlers ===

func createListNATRulesHandler(client *Client) mcp.ToolHandlerFor[ListNATRulesInput, ListNATRulesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListNATRulesInput]) (*mcp.CallToolResultFor[ListNATRulesOutput], error) {
		data, err := client.Get(ctx, "firewall/d_nat/search_rule")
		if err != nil {
			return toolErrorfFor[ListNATRulesOutput]("failed to query NAT rules: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ListNATRulesOutput]("failed to parse response: %s", err.Error()), nil
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
			return toolErrorFor[GetNATRuleOutput]("uuid parameter is required"), nil
		}

		data, err := client.Get(ctx, "firewall/d_nat/get_rule/"+input.UUID)
		if err != nil {
			return toolErrorfFor[GetNATRuleOutput]("failed to query NAT rule: %s", err.Error()), nil
		}

		var rule map[string]any
		if err := json.Unmarshal(data, &rule); err != nil {
			return toolErrorfFor[GetNATRuleOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(GetNATRuleOutput{Rule: rule}), nil
	}
}

func createAddNATRuleHandler(client *Client) mcp.ToolHandlerFor[AddNATRuleInput, AddNATRuleOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[AddNATRuleInput]) (*mcp.CallToolResultFor[AddNATRuleOutput], error) {
		input := params.Arguments
		if input.Protocol == "" {
			return toolErrorFor[AddNATRuleOutput]("protocol parameter is required (TCP, UDP, TCP/UDP)"), nil
		}
		if input.DstPort == "" {
			return toolErrorFor[AddNATRuleOutput]("destination_port (external port) parameter is required"), nil
		}
		if input.TargetIP == "" {
			return toolErrorFor[AddNATRuleOutput]("target (redirect target IP) parameter is required"), nil
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
			return toolErrorfFor[AddNATRuleOutput]("failed to add NAT rule: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[AddNATRuleOutput]("failed to parse response: %s", err.Error()), nil
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
			return toolErrorFor[UpdateNATRuleOutput]("uuid parameter is required"), nil
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
			return toolErrorfFor[UpdateNATRuleOutput]("failed to update NAT rule: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[UpdateNATRuleOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(UpdateNATRuleOutput{Result: result.Result}), nil
	}
}

func createDeleteNATRuleHandler(client *Client) mcp.ToolHandlerFor[DeleteNATRuleInput, DeleteNATRuleOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[DeleteNATRuleInput]) (*mcp.CallToolResultFor[DeleteNATRuleOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[DeleteNATRuleOutput]("uuid parameter is required"), nil
		}

		data, err := client.Post(ctx, "firewall/d_nat/del_rule/"+input.UUID, nil)
		if err != nil {
			return toolErrorfFor[DeleteNATRuleOutput]("failed to delete NAT rule: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[DeleteNATRuleOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(DeleteNATRuleOutput{Result: result.Result}), nil
	}
}

func createToggleNATRuleHandler(client *Client) mcp.ToolHandlerFor[ToggleNATRuleInput, ToggleNATRuleOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ToggleNATRuleInput]) (*mcp.CallToolResultFor[ToggleNATRuleOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[ToggleNATRuleOutput]("uuid parameter is required"), nil
		}

		path := "firewall/d_nat/toggle_rule/" + input.UUID
		if input.Disabled != "" {
			path += "/" + input.Disabled
		}

		data, err := client.Post(ctx, path, nil)
		if err != nil {
			return toolErrorfFor[ToggleNATRuleOutput]("failed to toggle NAT rule: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ToggleNATRuleOutput]("failed to parse response: %s", err.Error()), nil
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
		Description: "Lists OPNsense NAT port forwarding (DNAT) rules.",
	}, createListNATRulesHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_get_nat_rule",
		Description: "Retrieves detailed configuration of a specific NAT port forwarding rule.",
	}, createGetNATRuleHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_add_nat_rule",
		Description: "Adds a NAT port forwarding rule. Redirects external ports to internal IP:port. You must call ops_apply_changes(module=firewall) after adding to apply.",
	}, createAddNATRuleHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_update_nat_rule",
		Description: "Updates an existing NAT port forwarding rule. Used for changing target IP/port, etc. You must call ops_apply_changes(module=firewall) after updating to apply.",
	}, createUpdateNATRuleHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_delete_nat_rule",
		Description: "Deletes a NAT port forwarding rule. You must call ops_apply_changes(module=firewall) after deleting to apply.",
	}, createDeleteNATRuleHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_toggle_nat_rule",
		Description: "Enables or disables a NAT port forwarding rule. You must call ops_apply_changes(module=firewall) after changing to apply.",
	}, createToggleNATRuleHandler(client))

	slog.Info("OPNsense NAT tools registered: ops_list_nat_rules, ops_get_nat_rule, ops_add_nat_rule, ops_update_nat_rule, ops_delete_nat_rule, ops_toggle_nat_rule")
}
