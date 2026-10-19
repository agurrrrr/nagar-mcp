package opnsense

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output types for OPNsense firewall aliases ===

// ListFirewallAliasesInput has no required arguments. searchItem returns all rows.
type ListFirewallAliasesInput struct{}

// ListFirewallAliasesOutput carries the alias rows from searchItem.
type ListFirewallAliasesOutput struct {
	Aliases []map[string]any `json:"aliases"`
	Total   int              `json:"total"`
}

// GetFirewallAliasInput selects one alias by uuid.
type GetFirewallAliasInput struct {
	UUID string `json:"uuid" jsonschema:"UUID of the alias to retrieve"`
}

// GetFirewallAliasOutput carries the full alias object from getItem.
type GetFirewallAliasOutput struct {
	Alias map[string]any `json:"alias"`
}

// AddFirewallAliasInput creates a new firewall alias.
type AddFirewallAliasInput struct {
	Name        string `json:"name" jsonschema:"Alias name (unique)"`
	Type        string `json:"type" jsonschema:"Alias type: host, network, port, url, urltable, geoip, mac, external, internal"`
	Content     string `json:"content" jsonschema:"Alias content. For port: space-separated ports or ranges (e.g. 80 443 8080-8090). For host/network: space-separated IPs/CIDRs"`
	Description string `json:"description,omitempty" jsonschema:"Optional description"`
	Enabled     string `json:"enabled,omitempty" jsonschema:"Enabled (1 or 0, default: 1)"`
}

// AddFirewallAliasOutput carries the result and the new alias uuid.
type AddFirewallAliasOutput struct {
	Result string `json:"result"`
	UUID   string `json:"uuid,omitempty"`
}

// UpdateFirewallAliasInput edits an existing alias by uuid. Only provided
// fields are sent; omitted fields keep their current value.
type UpdateFirewallAliasInput struct {
	UUID        string `json:"uuid" jsonschema:"UUID of the alias to update"`
	Name        string `json:"name,omitempty" jsonschema:"Alias name (unique)"`
	Type        string `json:"type,omitempty" jsonschema:"Alias type: host, network, port, url, urltable, geoip, mac, external, internal"`
	Content     string `json:"content,omitempty" jsonschema:"Alias content (see ops_add_firewall_alias)"`
	Description string `json:"description,omitempty" jsonschema:"Optional description"`
	Enabled     string `json:"enabled,omitempty" jsonschema:"Enabled (1 or 0)"`
}

// UpdateFirewallAliasOutput carries the setItem result.
type UpdateFirewallAliasOutput struct {
	Result string `json:"result"`
}

// DeleteFirewallAliasInput selects one alias by uuid to remove.
type DeleteFirewallAliasInput struct {
	UUID string `json:"uuid" jsonschema:"UUID of the alias to delete"`
}

// DeleteFirewallAliasOutput carries the delItem result.
type DeleteFirewallAliasOutput struct {
	Result string `json:"result"`
}

// ToggleFirewallAliasInput selects one alias by uuid and sets enabled state.
type ToggleFirewallAliasInput struct {
	UUID    string `json:"uuid" jsonschema:"UUID of the alias to toggle"`
	Enabled string `json:"enabled,omitempty" jsonschema:"Enabled (1 or 0, toggles if empty)"`
}

// ToggleFirewallAliasOutput carries the toggleItem result.
type ToggleFirewallAliasOutput struct {
	Result string `json:"result"`
}

// === Handlers ===

func createListFirewallAliasesHandler(client *Client) mcp.ToolHandlerFor[ListFirewallAliasesInput, ListFirewallAliasesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListFirewallAliasesInput]) (*mcp.CallToolResultFor[ListFirewallAliasesOutput], error) {
		data, err := client.Get(ctx, "firewall/alias/searchItem")
		if err != nil {
			return toolErrorfFor[ListFirewallAliasesOutput]("failed to query aliases: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ListFirewallAliasesOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(ListFirewallAliasesOutput{
			Aliases: result.Rows,
			Total:   result.Total,
		}), nil
	}
}

func createGetFirewallAliasHandler(client *Client) mcp.ToolHandlerFor[GetFirewallAliasInput, GetFirewallAliasOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[GetFirewallAliasInput]) (*mcp.CallToolResultFor[GetFirewallAliasOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[GetFirewallAliasOutput]("uuid parameter is required"), nil
		}

		data, err := client.Get(ctx, "firewall/alias/getItem/"+input.UUID)
		if err != nil {
			return toolErrorfFor[GetFirewallAliasOutput]("failed to query alias: %s", err.Error()), nil
		}

		var wrapper struct {
			Alias map[string]any `json:"alias"`
		}
		if err := json.Unmarshal(data, &wrapper); err != nil {
			return toolErrorfFor[GetFirewallAliasOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(GetFirewallAliasOutput{Alias: wrapper.Alias}), nil
	}
}

func createAddFirewallAliasHandler(client *Client) mcp.ToolHandlerFor[AddFirewallAliasInput, AddFirewallAliasOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[AddFirewallAliasInput]) (*mcp.CallToolResultFor[AddFirewallAliasOutput], error) {
		input := params.Arguments
		if input.Name == "" {
			return toolErrorFor[AddFirewallAliasOutput]("name parameter is required"), nil
		}
		if input.Type == "" {
			return toolErrorFor[AddFirewallAliasOutput]("type parameter is required (host, network, port, url, urltable, geoip, mac, external, internal)"), nil
		}
		if input.Enabled == "" {
			input.Enabled = "1"
		}

		body := map[string]any{
			"alias": map[string]any{
				"name":        input.Name,
				"type":        input.Type,
				"content":     input.Content,
				"description": input.Description,
				"enabled":     input.Enabled,
			},
		}

		data, err := client.Post(ctx, "firewall/alias/addItem", body)
		if err != nil {
			return toolErrorfFor[AddFirewallAliasOutput]("failed to add alias: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[AddFirewallAliasOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(AddFirewallAliasOutput{
			Result: result.Result,
			UUID:   result.UUID,
		}), nil
	}
}

func createUpdateFirewallAliasHandler(client *Client) mcp.ToolHandlerFor[UpdateFirewallAliasInput, UpdateFirewallAliasOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[UpdateFirewallAliasInput]) (*mcp.CallToolResultFor[UpdateFirewallAliasOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[UpdateFirewallAliasOutput]("uuid parameter is required"), nil
		}

		alias := map[string]any{}
		if input.Name != "" {
			alias["name"] = input.Name
		}
		if input.Type != "" {
			alias["type"] = input.Type
		}
		if input.Content != "" {
			alias["content"] = input.Content
		}
		if input.Description != "" {
			alias["description"] = input.Description
		}
		if input.Enabled != "" {
			alias["enabled"] = input.Enabled
		}
		if len(alias) == 0 {
			return toolErrorFor[UpdateFirewallAliasOutput]("no fields to update (provide at least one of name, type, content, description, enabled)"), nil
		}

		body := map[string]any{"alias": alias}

		data, err := client.Post(ctx, "firewall/alias/setItem/"+input.UUID, body)
		if err != nil {
			return toolErrorfFor[UpdateFirewallAliasOutput]("failed to update alias: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[UpdateFirewallAliasOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(UpdateFirewallAliasOutput{Result: result.Result}), nil
	}
}

func createDeleteFirewallAliasHandler(client *Client) mcp.ToolHandlerFor[DeleteFirewallAliasInput, DeleteFirewallAliasOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[DeleteFirewallAliasInput]) (*mcp.CallToolResultFor[DeleteFirewallAliasOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[DeleteFirewallAliasOutput]("uuid parameter is required"), nil
		}

		data, err := client.Post(ctx, "firewall/alias/delItem/"+input.UUID, nil)
		if err != nil {
			return toolErrorfFor[DeleteFirewallAliasOutput]("failed to delete alias: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[DeleteFirewallAliasOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(DeleteFirewallAliasOutput{Result: result.Result}), nil
	}
}

func createToggleFirewallAliasHandler(client *Client) mcp.ToolHandlerFor[ToggleFirewallAliasInput, ToggleFirewallAliasOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ToggleFirewallAliasInput]) (*mcp.CallToolResultFor[ToggleFirewallAliasOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[ToggleFirewallAliasOutput]("uuid parameter is required"), nil
		}

		path := "firewall/alias/toggleItem/" + input.UUID
		if input.Enabled != "" {
			path += "/" + input.Enabled
		}

		data, err := client.Post(ctx, path, nil)
		if err != nil {
			return toolErrorfFor[ToggleFirewallAliasOutput]("failed to toggle alias: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ToggleFirewallAliasOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(ToggleFirewallAliasOutput{Result: result.Result}), nil
	}
}

// === Registration ===

func registerAliasTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_list_firewall_aliases",
		Description: "Lists OPNsense firewall aliases (host, network, port, geoip, etc.).",
	}, createListFirewallAliasesHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_get_firewall_alias",
		Description: "Retrieves the detailed configuration of a specific firewall alias.",
	}, createGetFirewallAliasHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_add_firewall_alias",
		Description: "Adds a new firewall alias. After adding, you must call ops_apply_changes(module=firewall) to apply.",
	}, createAddFirewallAliasHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_update_firewall_alias",
		Description: "Updates an existing firewall alias (content, type, description, enabled). You must call ops_apply_changes(module=firewall) after updating to apply.",
	}, createUpdateFirewallAliasHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_delete_firewall_alias",
		Description: "Deletes a firewall alias. You must call ops_apply_changes(module=firewall) after deleting to apply.",
	}, createDeleteFirewallAliasHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_toggle_firewall_alias",
		Description: "Enables or disables a firewall alias. You must call ops_apply_changes(module=firewall) after changing to apply.",
	}, createToggleFirewallAliasHandler(client))

	slog.Info("OPNsense alias tools registered: ops_list_firewall_aliases, ops_get_firewall_alias, ops_add_firewall_alias, ops_update_firewall_alias, ops_delete_firewall_alias, ops_toggle_firewall_alias")
}
