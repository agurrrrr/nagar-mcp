package opnsense

import (
	"context"
	"encoding/json"
	"log/slog"

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
	UUID string `json:"uuid" jsonschema:"Firewall rule UUID"`
}

type GetFirewallRuleOutput struct {
	Rule map[string]any `json:"rule"`
}

type AddFirewallRuleInput struct {
	Enabled     string `json:"enabled,omitempty" jsonschema:"Enabled (1 or 0, default: 1)"`
	Action      string `json:"action" jsonschema:"Action (pass, block, reject)"`
	Direction   string `json:"direction,omitempty" jsonschema:"Direction (in, out, default: in)"`
	IPProtocol  string `json:"ipprotocol,omitempty" jsonschema:"IP protocol (inet, inet6, default: inet)"`
	Protocol    string `json:"protocol,omitempty" jsonschema:"Protocol (TCP, UDP, ICMP, etc.)"`
	Interface   string `json:"interface,omitempty" jsonschema:"Interface (lan, wan, opt1, etc.)"`
	Source      string `json:"source_net,omitempty" jsonschema:"Source (any, IP/CIDR, alias)"`
	Destination string `json:"destination_net,omitempty" jsonschema:"Destination (any, IP/CIDR, alias)"`
	DstPort     string `json:"destination_port,omitempty" jsonschema:"Destination port (e.g. 80, 443, 8080-8090)"`
	Description string `json:"description,omitempty" jsonschema:"Rule description"`
}

type AddFirewallRuleOutput struct {
	Result string `json:"result"`
	UUID   string `json:"uuid"`
}

type ToggleFirewallRuleInput struct {
	UUID    string `json:"uuid" jsonschema:"Firewall rule UUID"`
	Enabled string `json:"enabled,omitempty" jsonschema:"Enabled (1 or 0, toggles if empty)"`
}

type ToggleFirewallRuleOutput struct {
	Result string `json:"result"`
}

type DeleteFirewallRuleInput struct {
	UUID string `json:"uuid" jsonschema:"Firewall rule UUID to delete"`
}

type DeleteFirewallRuleOutput struct {
	Result string `json:"result"`
}

// === Tool Handlers ===

func createListFirewallRulesHandler(client *Client) mcp.ToolHandlerFor[ListFirewallRulesInput, ListFirewallRulesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListFirewallRulesInput]) (*mcp.CallToolResultFor[ListFirewallRulesOutput], error) {
		data, err := client.Get(ctx, "firewall/filter/searchRule")
		if err != nil {
			return toolErrorfFor[ListFirewallRulesOutput]("failed to query firewall rules: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ListFirewallRulesOutput]("failed to parse response: %s", err.Error()), nil
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
			return toolErrorFor[GetFirewallRuleOutput]("uuid parameter is required"), nil
		}

		data, err := client.Get(ctx, "firewall/filter/getRule/"+input.UUID)
		if err != nil {
			return toolErrorfFor[GetFirewallRuleOutput]("failed to query rule: %s", err.Error()), nil
		}

		var rule map[string]any
		if err := json.Unmarshal(data, &rule); err != nil {
			return toolErrorfFor[GetFirewallRuleOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(GetFirewallRuleOutput{Rule: rule}), nil
	}
}

func createAddFirewallRuleHandler(client *Client) mcp.ToolHandlerFor[AddFirewallRuleInput, AddFirewallRuleOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[AddFirewallRuleInput]) (*mcp.CallToolResultFor[AddFirewallRuleOutput], error) {
		input := params.Arguments
		if input.Action == "" {
			return toolErrorFor[AddFirewallRuleOutput]("action parameter is required (pass, block, reject)"), nil
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
			return toolErrorfFor[AddFirewallRuleOutput]("failed to add rule: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[AddFirewallRuleOutput]("failed to parse response: %s", err.Error()), nil
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
			return toolErrorFor[ToggleFirewallRuleOutput]("uuid parameter is required"), nil
		}

		path := "firewall/filter/toggleRule/" + input.UUID
		if input.Enabled != "" {
			path += "/" + input.Enabled
		}

		data, err := client.Post(ctx, path, nil)
		if err != nil {
			return toolErrorfFor[ToggleFirewallRuleOutput]("failed to toggle rule: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ToggleFirewallRuleOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(ToggleFirewallRuleOutput{Result: result.Result}), nil
	}
}

func createDeleteFirewallRuleHandler(client *Client) mcp.ToolHandlerFor[DeleteFirewallRuleInput, DeleteFirewallRuleOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[DeleteFirewallRuleInput]) (*mcp.CallToolResultFor[DeleteFirewallRuleOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[DeleteFirewallRuleOutput]("uuid parameter is required"), nil
		}

		data, err := client.Post(ctx, "firewall/filter/delRule/"+input.UUID, nil)
		if err != nil {
			return toolErrorfFor[DeleteFirewallRuleOutput]("failed to delete rule: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[DeleteFirewallRuleOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(DeleteFirewallRuleOutput{Result: result.Result}), nil
	}
}

// === Registration ===

func registerFirewallTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_list_firewall_rules",
		Description: "Lists OPNsense firewall filter rules.",
	}, createListFirewallRulesHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_get_firewall_rule",
		Description: "Retrieves the detailed configuration of a specific firewall rule.",
	}, createGetFirewallRuleHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_add_firewall_rule",
		Description: "Adds a new firewall filter rule. After adding, you must call ops_apply_changes(module=firewall) to apply.",
	}, createAddFirewallRuleHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_toggle_firewall_rule",
		Description: "Enables/disables a firewall rule. After changing, you must call ops_apply_changes(module=firewall) to apply.",
	}, createToggleFirewallRuleHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_delete_firewall_rule",
		Description: "Deletes a firewall rule. After deleting, you must call ops_apply_changes(module=firewall) to apply.",
	}, createDeleteFirewallRuleHandler(client))

	slog.Info("OPNsense firewall tools registered: ops_list_firewall_rules, ops_get_firewall_rule, ops_add_firewall_rule, ops_toggle_firewall_rule, ops_delete_firewall_rule")
}
