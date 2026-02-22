package opnsense

import (
	"context"
	"encoding/json"
	"log/slog"

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
	Hostname string `json:"hostname" jsonschema:"Hostname (e.g. myserver)"`
	Domain   string `json:"domain,omitempty" jsonschema:"Domain (e.g. local)"`
	Server   string `json:"server" jsonschema:"IP address (e.g. 192.168.1.100)"`
	Descr    string `json:"description,omitempty" jsonschema:"Description"`
}

type AddDNSOverrideOutput struct {
	Result string `json:"result"`
	UUID   string `json:"uuid"`
}

type DeleteDNSOverrideInput struct {
	UUID string `json:"uuid" jsonschema:"UUID of the DNS override to delete"`
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
			return toolErrorfFor[ListDNSOverridesOutput]("failed to query DNS overrides: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ListDNSOverridesOutput]("failed to parse response: %s", err.Error()), nil
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
			return toolErrorFor[AddDNSOverrideOutput]("hostname parameter is required"), nil
		}
		if input.Server == "" {
			return toolErrorFor[AddDNSOverrideOutput]("server (IP address) parameter is required"), nil
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
			return toolErrorfFor[AddDNSOverrideOutput]("failed to add DNS override: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[AddDNSOverrideOutput]("failed to parse response: %s", err.Error()), nil
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
			return toolErrorFor[DeleteDNSOverrideOutput]("uuid parameter is required"), nil
		}

		data, err := client.Post(ctx, "unbound/settings/delHostOverride/"+input.UUID, nil)
		if err != nil {
			return toolErrorfFor[DeleteDNSOverrideOutput]("failed to delete DNS override: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[DeleteDNSOverrideOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(DeleteDNSOverrideOutput{Result: result.Result}), nil
	}
}

func createDNSStatsHandler(client *Client) mcp.ToolHandlerFor[DNSStatsInput, DNSStatsOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[DNSStatsInput]) (*mcp.CallToolResultFor[DNSStatsOutput], error) {
		data, err := client.Get(ctx, "unbound/diagnostics/stats")
		if err != nil {
			return toolErrorfFor[DNSStatsOutput]("failed to query DNS statistics: %s", err.Error()), nil
		}

		var stats map[string]any
		if err := json.Unmarshal(data, &stats); err != nil {
			return toolErrorfFor[DNSStatsOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(DNSStatsOutput{Stats: stats}), nil
	}
}

// === Registration ===

func registerDNSTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_list_dns_overrides",
		Description: "Lists OPNsense Unbound DNS host overrides.",
	}, createListDNSOverridesHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_add_dns_override",
		Description: "Adds a DNS host override. You must call ops_apply_changes(module=unbound) after adding to apply.",
	}, createAddDNSOverrideHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_delete_dns_override",
		Description: "Deletes a DNS host override. You must call ops_apply_changes(module=unbound) after deleting to apply.",
	}, createDeleteDNSOverrideHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_dns_stats",
		Description: "Retrieves Unbound DNS server statistics. Provides query count, cache hit rate, and other information.",
	}, createDNSStatsHandler(client))

	slog.Info("OPNsense DNS tools registered: ops_list_dns_overrides, ops_add_dns_override, ops_delete_dns_override, ops_dns_stats")
}
