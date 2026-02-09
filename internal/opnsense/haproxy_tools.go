package opnsense

import (
	"context"
	"encoding/json"
	"fmt"

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
	Name    string `json:"name" jsonschema:"서버 이름"`
	Address string `json:"address" jsonschema:"서버 주소 (IP 또는 호스트명)"`
	Port    string `json:"port" jsonschema:"서버 포트"`
	Mode    string `json:"mode,omitempty" jsonschema:"모드 (active, backup, disabled, 기본값: active)"`
	SSL     string `json:"ssl,omitempty" jsonschema:"SSL 사용 여부 (1 또는 0, 기본값: 0)"`
	Weight  string `json:"weight,omitempty" jsonschema:"가중치 (기본값: 1)"`
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
	Name    string `json:"name" jsonschema:"백엔드 이름"`
	Mode    string `json:"mode,omitempty" jsonschema:"프로토콜 모드 (http, tcp, 기본값: http)"`
	Balance string `json:"balance,omitempty" jsonschema:"로드밸런싱 알고리즘 (roundrobin, leastconn, source, 기본값: roundrobin)"`
	Servers string `json:"linkedServers,omitempty" jsonschema:"연결할 서버 UUID (쉼표로 구분)"`
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
			return toolErrorfFor[HAListServersOutput]("HAProxy 서버 목록 조회 실패: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[HAListServersOutput]("응답 파싱 실패: %s", err.Error()), nil
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
			return toolErrorFor[HAAddServerOutput]("name 파라미터가 필요합니다"), nil
		}
		if input.Address == "" {
			return toolErrorFor[HAAddServerOutput]("address 파라미터가 필요합니다"), nil
		}
		if input.Port == "" {
			return toolErrorFor[HAAddServerOutput]("port 파라미터가 필요합니다"), nil
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
			return toolErrorfFor[HAAddServerOutput]("HAProxy 서버 추가 실패: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[HAAddServerOutput]("응답 파싱 실패: %s", err.Error()), nil
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
			return toolErrorfFor[HAListBackendsOutput]("HAProxy 백엔드 조회 실패: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[HAListBackendsOutput]("응답 파싱 실패: %s", err.Error()), nil
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
			return toolErrorFor[HAAddBackendOutput]("name 파라미터가 필요합니다"), nil
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
			return toolErrorfFor[HAAddBackendOutput]("HAProxy 백엔드 생성 실패: %s", err.Error()), nil
		}

		var result UUIDResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[HAAddBackendOutput]("응답 파싱 실패: %s", err.Error()), nil
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
			return toolErrorfFor[HAListFrontendsOutput]("HAProxy 프론트엔드 조회 실패: %s", err.Error()), nil
		}

		var result SearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[HAListFrontendsOutput]("응답 파싱 실패: %s", err.Error()), nil
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
			return toolErrorfFor[HAStatisticsOutput]("HAProxy 통계 조회 실패: %s", err.Error()), nil
		}

		var stats map[string]any
		if err := json.Unmarshal(data, &stats); err != nil {
			return toolErrorfFor[HAStatisticsOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(HAStatisticsOutput{Stats: stats}), nil
	}
}

func createHAConfigTestHandler(client *Client) mcp.ToolHandlerFor[HAConfigTestInput, HAConfigTestOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[HAConfigTestInput]) (*mcp.CallToolResultFor[HAConfigTestOutput], error) {
		data, err := client.Get(ctx, "haproxy/service/configtest")
		if err != nil {
			return toolErrorfFor[HAConfigTestOutput]("HAProxy 설정 검증 실패: %s", err.Error()), nil
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			return toolErrorfFor[HAConfigTestOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(HAConfigTestOutput{Result: result}), nil
	}
}

// === Registration ===

func registerHAProxyTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_ha_list_servers",
		Description: "HAProxy 백엔드 서버 목록을 조회합니다.",
	}, createHAListServersHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_ha_add_server",
		Description: "HAProxy 백엔드 서버를 추가합니다. 추가 후 ops_ha_config_test로 검증하고 ops_apply_changes(module=haproxy)를 호출해야 적용됩니다.",
	}, createHAAddServerHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_ha_list_backends",
		Description: "HAProxy 백엔드 풀 목록을 조회합니다.",
	}, createHAListBackendsHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_ha_add_backend",
		Description: "HAProxy 백엔드 풀을 생성합니다. 서버를 연결하고 로드밸런싱 알고리즘을 설정합니다.",
	}, createHAAddBackendHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_ha_list_frontends",
		Description: "HAProxy 프론트엔드(리스너) 목록을 조회합니다.",
	}, createHAListFrontendsHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_ha_statistics",
		Description: "HAProxy 성능 통계를 조회합니다. 연결 수, 바이트, 응답 시간 등의 카운터를 제공합니다.",
	}, createHAStatisticsHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_ha_config_test",
		Description: "HAProxy 설정의 유효성을 검증합니다. 변경사항 적용 전에 반드시 호출하세요.",
	}, createHAConfigTestHandler(client))

	fmt.Println("OPNsense HAProxy tools registered: ops_ha_list_servers, ops_ha_add_server, ops_ha_list_backends, ops_ha_add_backend, ops_ha_list_frontends, ops_ha_statistics, ops_ha_config_test")
}
