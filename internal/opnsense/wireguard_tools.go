package opnsense

import (
	"context"
	"encoding/json"
	"fmt"

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
	Name       string `json:"name" jsonschema:"인스턴스 이름"`
	Port       string `json:"port,omitempty" jsonschema:"리스닝 포트 (기본값: 51820)"`
	TunnelAddr string `json:"tunneladdress,omitempty" jsonschema:"터널 주소 (예: 10.10.10.1/24)"`
	Enabled    string `json:"enabled,omitempty" jsonschema:"활성화 여부 (1 또는 0, 기본값: 1)"`
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
	Name       string `json:"name" jsonschema:"피어 이름"`
	PublicKey  string `json:"pubkey" jsonschema:"피어의 WireGuard 공개키"`
	TunnelAddr string `json:"tunneladdress,omitempty" jsonschema:"피어 터널 주소 (예: 10.10.10.2/32)"`
	Endpoint   string `json:"endpoint,omitempty" jsonschema:"피어 엔드포인트 (예: peer.example.com:51820)"`
	AllowedIPs string `json:"allowedips,omitempty" jsonschema:"허용 IP 대역 (예: 10.10.10.2/32)"`
	KeepAlive  string `json:"keepalive,omitempty" jsonschema:"Keepalive 간격 (초, 기본값: 25)"`
	Enabled    string `json:"enabled,omitempty" jsonschema:"활성화 여부 (1 또는 0, 기본값: 1)"`
}

type WGAddPeerOutput struct {
	Result string `json:"result"`
	UUID   string `json:"uuid"`
}

type WGDeletePeerInput struct {
	UUID string `json:"uuid" jsonschema:"삭제할 피어 UUID"`
}

type WGDeletePeerOutput struct {
	Result string `json:"result"`
}

// === Tool Handlers ===

func createWGListInstancesHandler(client *Client) mcp.ToolHandlerFor[WGListInstancesInput, WGListInstancesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[WGListInstancesInput]) (*mcp.CallToolResultFor[WGListInstancesOutput], error) {
		data, err := client.Get(ctx, "wireguard/server/searchServer")
		if err != nil {
			return toolErrorfFor[WGListInstancesOutput]("WireGuard 인스턴스 조회 실패: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[WGListInstancesOutput]("응답 파싱 실패: %s", err.Error()), nil
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
			return toolErrorFor[WGAddInstanceOutput]("name 파라미터가 필요합니다"), nil
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
			return toolErrorfFor[WGAddInstanceOutput]("WireGuard 인스턴스 생성 실패: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[WGAddInstanceOutput]("응답 파싱 실패: %s", err.Error()), nil
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
			return toolErrorfFor[WGListPeersOutput]("WireGuard 피어 조회 실패: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[WGListPeersOutput]("응답 파싱 실패: %s", err.Error()), nil
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
			return toolErrorFor[WGAddPeerOutput]("name 파라미터가 필요합니다"), nil
		}
		if input.PublicKey == "" {
			return toolErrorFor[WGAddPeerOutput]("pubkey (공개키) 파라미터가 필요합니다"), nil
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
			return toolErrorfFor[WGAddPeerOutput]("WireGuard 피어 추가 실패: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[WGAddPeerOutput]("응답 파싱 실패: %s", err.Error()), nil
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
			return toolErrorFor[WGDeletePeerOutput]("uuid 파라미터가 필요합니다"), nil
		}

		data, err := client.Post(ctx, "wireguard/client/delClient/"+input.UUID, nil)
		if err != nil {
			return toolErrorfFor[WGDeletePeerOutput]("WireGuard 피어 삭제 실패: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[WGDeletePeerOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(WGDeletePeerOutput{Result: result.Result}), nil
	}
}

// === Registration ===

func registerWireGuardTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_wg_list_instances",
		Description: "WireGuard VPN 인스턴스(서버) 목록을 조회합니다.",
	}, createWGListInstancesHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_wg_add_instance",
		Description: "새 WireGuard VPN 인스턴스를 생성합니다. 키페어는 자동 생성됩니다. 생성 후 ops_apply_changes(module=wireguard)를 호출해야 적용됩니다.",
	}, createWGAddInstanceHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_wg_list_peers",
		Description: "WireGuard VPN 피어(클라이언트) 목록을 조회합니다.",
	}, createWGListPeersHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_wg_add_peer",
		Description: "새 WireGuard VPN 피어를 추가합니다. 추가 후 ops_apply_changes(module=wireguard)를 호출해야 적용됩니다.",
	}, createWGAddPeerHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_wg_delete_peer",
		Description: "WireGuard VPN 피어를 삭제합니다. 삭제 후 ops_apply_changes(module=wireguard)를 호출해야 적용됩니다.",
	}, createWGDeletePeerHandler(client))

	fmt.Println("OPNsense WireGuard tools registered: ops_wg_list_instances, ops_wg_add_instance, ops_wg_list_peers, ops_wg_add_peer, ops_wg_delete_peer")
}
