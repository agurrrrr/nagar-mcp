package opnsense

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

type WGListInstancesInput struct{}

type WGListInstancesOutput struct {
	Instances []map[string]any `json:"instances"`
	Total     int              `json:"total"`
	RowCount  int              `json:"rowCount"`
}

type WGAddInstanceInput struct {
	Name       string `json:"name" jsonschema:"Instance name"`
	Port       string `json:"port,omitempty" jsonschema:"Listening port (default: 51820)"`
	TunnelAddr string `json:"tunneladdress,omitempty" jsonschema:"Tunnel address (e.g. 10.10.10.1/24)"`
	Enabled    string `json:"enabled,omitempty" jsonschema:"Enabled (1 or 0, default: 1)"`
}

type WGAddInstanceOutput struct {
	Result string `json:"result"`
	UUID   string `json:"uuid"`
}

type WGListPeersInput struct{}

type WGListPeersOutput struct {
	Peers    []map[string]any `json:"peers"`
	Total    int              `json:"total"`
	RowCount int              `json:"rowCount"`
}

type WGAddPeerInput struct {
	Name       string `json:"name" jsonschema:"Peer name"`
	PublicKey  string `json:"pubkey" jsonschema:"WireGuard public key of the peer"`
	TunnelAddr string `json:"tunneladdress,omitempty" jsonschema:"Peer tunnel address (e.g. 10.10.10.2/32)"`
	Endpoint   string `json:"endpoint,omitempty" jsonschema:"Peer endpoint (e.g. peer.example.com:51820)"`
	AllowedIPs string `json:"allowedips,omitempty" jsonschema:"Allowed IP ranges (e.g. 10.10.10.2/32)"`
	KeepAlive  string `json:"keepalive,omitempty" jsonschema:"Keepalive interval in seconds (default: 25)"`
	Enabled    string `json:"enabled,omitempty" jsonschema:"Enabled (1 or 0, default: 1)"`
}

type WGAddPeerOutput struct {
	Result string `json:"result"`
	UUID   string `json:"uuid"`
}

type WGDeletePeerInput struct {
	UUID string `json:"uuid" jsonschema:"UUID of the peer to delete"`
}

type WGDeletePeerOutput struct {
	Result string `json:"result"`
}

// === Tool Handlers ===

func createWGListInstancesHandler(client *Client) mcp.ToolHandlerFor[WGListInstancesInput, WGListInstancesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[WGListInstancesInput]) (*mcp.CallToolResultFor[WGListInstancesOutput], error) {
		data, err := client.Get(ctx, "wireguard/server/searchServer")
		if err != nil {
			return toolErrorfFor[WGListInstancesOutput]("failed to query WireGuard instances: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[WGListInstancesOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(WGListInstancesOutput{
			Instances: result.Rows,
			Total:     result.Total,
			RowCount:  result.RowCount,
		}), nil
	}
}

func createWGAddInstanceHandler(client *Client) mcp.ToolHandlerFor[WGAddInstanceInput, WGAddInstanceOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[WGAddInstanceInput]) (*mcp.CallToolResultFor[WGAddInstanceOutput], error) {
		input := params.Arguments
		if input.Name == "" {
			return toolErrorFor[WGAddInstanceOutput]("name parameter is required"), nil
		}

		if input.Enabled == "" {
			input.Enabled = "1"
		}
		if input.Port == "" {
			input.Port = "51820"
		}

		body := map[string]any{
			"server": map[string]any{
				"enabled":       input.Enabled,
				"name":          input.Name,
				"port":          input.Port,
				"tunneladdress": input.TunnelAddr,
			},
		}

		data, err := client.Post(ctx, "wireguard/server/addServer", body)
		if err != nil {
			return toolErrorfFor[WGAddInstanceOutput]("failed to create WireGuard instance: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[WGAddInstanceOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(WGAddInstanceOutput{
			Result: result.Result,
			UUID:   result.UUID,
		}), nil
	}
}

func createWGListPeersHandler(client *Client) mcp.ToolHandlerFor[WGListPeersInput, WGListPeersOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[WGListPeersInput]) (*mcp.CallToolResultFor[WGListPeersOutput], error) {
		data, err := client.Get(ctx, "wireguard/client/searchClient")
		if err != nil {
			return toolErrorfFor[WGListPeersOutput]("failed to query WireGuard peers: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[WGListPeersOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(WGListPeersOutput{
			Peers:    result.Rows,
			Total:    result.Total,
			RowCount: result.RowCount,
		}), nil
	}
}

func createWGAddPeerHandler(client *Client) mcp.ToolHandlerFor[WGAddPeerInput, WGAddPeerOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[WGAddPeerInput]) (*mcp.CallToolResultFor[WGAddPeerOutput], error) {
		input := params.Arguments
		if input.Name == "" {
			return toolErrorFor[WGAddPeerOutput]("name parameter is required"), nil
		}
		if input.PublicKey == "" {
			return toolErrorFor[WGAddPeerOutput]("pubkey (public key) parameter is required"), nil
		}

		if input.Enabled == "" {
			input.Enabled = "1"
		}
		if input.KeepAlive == "" {
			input.KeepAlive = "25"
		}

		body := map[string]any{
			"client": map[string]any{
				"enabled":       input.Enabled,
				"name":          input.Name,
				"pubkey":        input.PublicKey,
				"tunneladdress": input.TunnelAddr,
				"endpoint":      input.Endpoint,
				"allowedips":    input.AllowedIPs,
				"keepalive":     input.KeepAlive,
			},
		}

		data, err := client.Post(ctx, "wireguard/client/addClient", body)
		if err != nil {
			return toolErrorfFor[WGAddPeerOutput]("failed to add WireGuard peer: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[WGAddPeerOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(WGAddPeerOutput{
			Result: result.Result,
			UUID:   result.UUID,
		}), nil
	}
}

func createWGDeletePeerHandler(client *Client) mcp.ToolHandlerFor[WGDeletePeerInput, WGDeletePeerOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[WGDeletePeerInput]) (*mcp.CallToolResultFor[WGDeletePeerOutput], error) {
		input := params.Arguments
		if input.UUID == "" {
			return toolErrorFor[WGDeletePeerOutput]("uuid parameter is required"), nil
		}

		data, err := client.Post(ctx, "wireguard/client/delClient/"+input.UUID, nil)
		if err != nil {
			return toolErrorfFor[WGDeletePeerOutput]("failed to delete WireGuard peer: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[WGDeletePeerOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(WGDeletePeerOutput{Result: result.Result}), nil
	}
}

// === Registration ===

func registerWireGuardTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_wg_list_instances",
		Description: "Lists WireGuard VPN instances (servers).",
	}, createWGListInstancesHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_wg_add_instance",
		Description: "Creates a new WireGuard VPN instance. Key pairs are auto-generated. You must call ops_apply_changes(module=wireguard) after creating to apply.",
	}, createWGAddInstanceHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_wg_list_peers",
		Description: "Lists WireGuard VPN peers (clients).",
	}, createWGListPeersHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_wg_add_peer",
		Description: "Adds a new WireGuard VPN peer. You must call ops_apply_changes(module=wireguard) after adding to apply.",
	}, createWGAddPeerHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_wg_delete_peer",
		Description: "Deletes a WireGuard VPN peer. You must call ops_apply_changes(module=wireguard) after deleting to apply.",
	}, createWGDeletePeerHandler(client))

	slog.Info("OPNsense WireGuard tools registered: ops_wg_list_instances, ops_wg_add_instance, ops_wg_list_peers, ops_wg_add_peer, ops_wg_delete_peer")
}
