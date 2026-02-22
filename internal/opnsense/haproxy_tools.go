package opnsense

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

type HAListServersInput struct{}

type HAListServersOutput struct {
	Servers  []map[string]any `json:"servers"`
	Total    int              `json:"total"`
	RowCount int              `json:"rowCount"`
}

type HAAddServerInput struct {
	Name    string `json:"name" jsonschema:"Server name"`
	Address string `json:"address" jsonschema:"Server address (IP or hostname)"`
	Port    string `json:"port" jsonschema:"Server port"`
	Mode    string `json:"mode,omitempty" jsonschema:"Mode (active, backup, disabled, default: active)"`
	SSL     string `json:"ssl,omitempty" jsonschema:"SSL enabled (1 or 0, default: 0)"`
	Weight  string `json:"weight,omitempty" jsonschema:"Weight (default: 1)"`
}

type HAAddServerOutput struct {
	Result string `json:"result"`
	UUID   string `json:"uuid"`
}

type HAListBackendsInput struct{}

type HAListBackendsOutput struct {
	Backends []map[string]any `json:"backends"`
	Total    int              `json:"total"`
	RowCount int              `json:"rowCount"`
}

type HAAddBackendInput struct {
	Name    string `json:"name" jsonschema:"Backend name"`
	Mode    string `json:"mode,omitempty" jsonschema:"Protocol mode (http, tcp, default: http)"`
	Balance string `json:"balance,omitempty" jsonschema:"Load balancing algorithm (roundrobin, leastconn, source, default: roundrobin)"`
	Servers string `json:"linkedServers,omitempty" jsonschema:"Server UUIDs to link (comma-separated)"`
}

type HAAddBackendOutput struct {
	Result string `json:"result"`
	UUID   string `json:"uuid"`
}

type HAListFrontendsInput struct{}

type HAListFrontendsOutput struct {
	Frontends []map[string]any `json:"frontends"`
	Total     int              `json:"total"`
	RowCount  int              `json:"rowCount"`
}

type HAStatisticsInput struct{}

type HAStatisticsOutput struct {
	Stats map[string]any `json:"stats"`
}

type HAConfigTestInput struct{}

type HAConfigTestOutput struct {
	Result map[string]any `json:"result"`
}

// === Tool Handlers ===

func createHAListServersHandler(client *Client) mcp.ToolHandlerFor[HAListServersInput, HAListServersOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[HAListServersInput]) (*mcp.CallToolResultFor[HAListServersOutput], error) {
		data, err := client.Get(ctx, "haproxy/settings/searchServers")
		if err != nil {
			return toolErrorfFor[HAListServersOutput]("failed to list HAProxy servers: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[HAListServersOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(HAListServersOutput{
			Servers:  result.Rows,
			Total:    result.Total,
			RowCount: result.RowCount,
		}), nil
	}
}

func createHAAddServerHandler(client *Client) mcp.ToolHandlerFor[HAAddServerInput, HAAddServerOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[HAAddServerInput]) (*mcp.CallToolResultFor[HAAddServerOutput], error) {
		input := params.Arguments
		if input.Name == "" {
			return toolErrorFor[HAAddServerOutput]("name parameter is required"), nil
		}
		if input.Address == "" {
			return toolErrorFor[HAAddServerOutput]("address parameter is required"), nil
		}
		if input.Port == "" {
			return toolErrorFor[HAAddServerOutput]("port parameter is required"), nil
		}

		if input.Mode == "" {
			input.Mode = "active"
		}
		if input.SSL == "" {
			input.SSL = "0"
		}
		if input.Weight == "" {
			input.Weight = "1"
		}

		body := map[string]any{
			"server": map[string]any{
				"name":    input.Name,
				"address": input.Address,
				"port":    input.Port,
				"mode":    input.Mode,
				"ssl":     input.SSL,
				"weight":  input.Weight,
			},
		}

		data, err := client.Post(ctx, "haproxy/settings/addServer", body)
		if err != nil {
			return toolErrorfFor[HAAddServerOutput]("failed to add HAProxy server: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[HAAddServerOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(HAAddServerOutput{
			Result: result.Result,
			UUID:   result.UUID,
		}), nil
	}
}

func createHAListBackendsHandler(client *Client) mcp.ToolHandlerFor[HAListBackendsInput, HAListBackendsOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[HAListBackendsInput]) (*mcp.CallToolResultFor[HAListBackendsOutput], error) {
		data, err := client.Get(ctx, "haproxy/settings/searchBackends")
		if err != nil {
			return toolErrorfFor[HAListBackendsOutput]("failed to query HAProxy backends: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[HAListBackendsOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(HAListBackendsOutput{
			Backends: result.Rows,
			Total:    result.Total,
			RowCount: result.RowCount,
		}), nil
	}
}

func createHAAddBackendHandler(client *Client) mcp.ToolHandlerFor[HAAddBackendInput, HAAddBackendOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[HAAddBackendInput]) (*mcp.CallToolResultFor[HAAddBackendOutput], error) {
		input := params.Arguments
		if input.Name == "" {
			return toolErrorFor[HAAddBackendOutput]("name parameter is required"), nil
		}

		if input.Mode == "" {
			input.Mode = "http"
		}
		if input.Balance == "" {
			input.Balance = "roundrobin"
		}

		body := map[string]any{
			"backend": map[string]any{
				"enabled":       "1",
				"name":          input.Name,
				"mode":          input.Mode,
				"algorithm":     input.Balance,
				"linkedServers": input.Servers,
			},
		}

		data, err := client.Post(ctx, "haproxy/settings/addBackend", body)
		if err != nil {
			return toolErrorfFor[HAAddBackendOutput]("failed to create HAProxy backend: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[HAAddBackendOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(HAAddBackendOutput{
			Result: result.Result,
			UUID:   result.UUID,
		}), nil
	}
}

func createHAListFrontendsHandler(client *Client) mcp.ToolHandlerFor[HAListFrontendsInput, HAListFrontendsOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[HAListFrontendsInput]) (*mcp.CallToolResultFor[HAListFrontendsOutput], error) {
		data, err := client.Get(ctx, "haproxy/settings/searchFrontends")
		if err != nil {
			return toolErrorfFor[HAListFrontendsOutput]("failed to query HAProxy frontends: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[HAListFrontendsOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(HAListFrontendsOutput{
			Frontends: result.Rows,
			Total:     result.Total,
			RowCount:  result.RowCount,
		}), nil
	}
}

func createHAStatisticsHandler(client *Client) mcp.ToolHandlerFor[HAStatisticsInput, HAStatisticsOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[HAStatisticsInput]) (*mcp.CallToolResultFor[HAStatisticsOutput], error) {
		data, err := client.Get(ctx, "haproxy/statistics/counters")
		if err != nil {
			return toolErrorfFor[HAStatisticsOutput]("failed to query HAProxy statistics: %s", err.Error()), nil
		}

		var stats map[string]any
		if err := json.Unmarshal(data, &stats); err != nil {
			return toolErrorfFor[HAStatisticsOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(HAStatisticsOutput{Stats: stats}), nil
	}
}

func createHAConfigTestHandler(client *Client) mcp.ToolHandlerFor[HAConfigTestInput, HAConfigTestOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[HAConfigTestInput]) (*mcp.CallToolResultFor[HAConfigTestOutput], error) {
		data, err := client.Get(ctx, "haproxy/service/configtest")
		if err != nil {
			return toolErrorfFor[HAConfigTestOutput]("failed to validate HAProxy configuration: %s", err.Error()), nil
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[HAConfigTestOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(HAConfigTestOutput{Result: result}), nil
	}
}

// === Registration ===

func registerHAProxyTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_ha_list_servers",
		Description: "Lists HAProxy backend servers.",
	}, createHAListServersHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_ha_add_server",
		Description: "Adds an HAProxy backend server. After adding, validate with ops_ha_config_test and call ops_apply_changes(module=haproxy) to apply.",
	}, createHAAddServerHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_ha_list_backends",
		Description: "Lists HAProxy backend pools.",
	}, createHAListBackendsHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_ha_add_backend",
		Description: "Creates an HAProxy backend pool. Links servers and configures the load balancing algorithm.",
	}, createHAAddBackendHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_ha_list_frontends",
		Description: "Lists HAProxy frontends (listeners).",
	}, createHAListFrontendsHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_ha_statistics",
		Description: "Retrieves HAProxy performance statistics. Provides counters for connections, bytes, response times, etc.",
	}, createHAStatisticsHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_ha_config_test",
		Description: "Validates HAProxy configuration. Must be called before applying changes.",
	}, createHAConfigTestHandler(client))

	slog.Info("OPNsense HAProxy tools registered: ops_ha_list_servers, ops_ha_add_server, ops_ha_list_backends, ops_ha_add_backend, ops_ha_list_frontends, ops_ha_statistics, ops_ha_config_test")
}
