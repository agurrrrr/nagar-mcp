package opnsense

import (
	"context"
	"encoding/json"
	"log/slog"

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
	Service string `json:"service" jsonschema:"Service name (e.g. unbound, wireguard, haproxy)"`
	Action  string `json:"action" jsonschema:"Action (start, stop, restart, status)"`
}

type ServiceActionOutput struct {
	Result map[string]any `json:"result"`
}

type ApplyChangesInput struct {
	Module string `json:"module" jsonschema:"Module name (e.g. firewall, unbound, wireguard, haproxy)"`
}

type ApplyChangesOutput struct {
	Result map[string]any `json:"result"`
}

// === Tool Handlers ===

func createListServicesHandler(client *Client) mcp.ToolHandlerFor[ListServicesInput, ListServicesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListServicesInput]) (*mcp.CallToolResultFor[ListServicesOutput], error) {
		data, err := client.Get(ctx, "core/service/search")
		if err != nil {
			return toolErrorfFor[ListServicesOutput]("failed to list services: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ListServicesOutput]("failed to parse response: %s", err.Error()), nil
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
			return toolErrorFor[ServiceActionOutput]("service parameter is required"), nil
		}
		if input.Action == "" {
			return toolErrorFor[ServiceActionOutput]("action parameter is required (start, stop, restart, status)"), nil
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
			return toolErrorfFor[ServiceActionOutput]("service %s %s failed: %s", input.Service, input.Action, err.Error()), nil
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ServiceActionOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(ServiceActionOutput{Result: result}), nil
	}
}

func createApplyChangesHandler(client *Client) mcp.ToolHandlerFor[ApplyChangesInput, ApplyChangesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ApplyChangesInput]) (*mcp.CallToolResultFor[ApplyChangesOutput], error) {
		input := params.Arguments
		if input.Module == "" {
			return toolErrorFor[ApplyChangesOutput]("module parameter is required"), nil
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
			return toolErrorfFor[ApplyChangesOutput]("failed to apply changes: %s", err.Error()), nil
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[ApplyChangesOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(ApplyChangesOutput{Result: result}), nil
	}
}

// === Registration ===

func registerServiceTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_list_services",
		Description: "Lists running services and their status on OPNsense.",
	}, createListServicesHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_service_action",
		Description: "Controls an OPNsense service (start, stop, restart, status).",
	}, createServiceActionHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_apply_changes",
		Description: "Applies configuration changes. Must be called after modifying firewall rules, DNS, WireGuard, HAProxy, etc.",
	}, createApplyChangesHandler(client))

	slog.Info("OPNsense service tools registered: ops_list_services, ops_service_action, ops_apply_changes")
}
