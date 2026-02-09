package opnsense

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

type FirmwareStatusInput struct{}

type FirmwareStatusOutput struct {
	Status             string `json:"status"`
	DownloadSize       string `json:"download_size,omitempty"`
	UpdatesAvailable   int    `json:"updates_available"`
	UpgradeNeedsReboot string `json:"upgrade_needs_reboot,omitempty"`
}

type SystemStatusInput struct{}

type SystemStatusOutput struct {
	Info map[string]any `json:"info"`
}

// === Tool Handlers ===

func createFirmwareStatusHandler(client *Client) mcp.ToolHandlerFor[FirmwareStatusInput, FirmwareStatusOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[FirmwareStatusInput]) (*mcp.CallToolResultFor[FirmwareStatusOutput], error) {
		data, err := client.Get(ctx, "core/firmware/status")
		if err != nil {
			return toolErrorfFor[FirmwareStatusOutput]("펌웨어 상태 조회 실패: %s", err.Error()), nil
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return toolErrorfFor[FirmwareStatusOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		output := FirmwareStatusOutput{
			Status: fmt.Sprintf("%v", raw["status"]),
		}
		if v, ok := raw["download_size"]; ok {
			output.DownloadSize = fmt.Sprintf("%v", v)
		}
		if v, ok := raw["new_packages"]; ok {
			if pkgs, ok := v.([]any); ok {
				output.UpdatesAvailable = len(pkgs)
			}
		}
		if v, ok := raw["upgrade_needs_reboot"]; ok {
			output.UpgradeNeedsReboot = fmt.Sprintf("%v", v)
		}

		return toolSuccessFor(output), nil
	}
}

func createSystemStatusHandler(client *Client) mcp.ToolHandlerFor[SystemStatusInput, SystemStatusOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[SystemStatusInput]) (*mcp.CallToolResultFor[SystemStatusOutput], error) {
		data, err := client.Get(ctx, "core/firmware/info")
		if err != nil {
			return toolErrorfFor[SystemStatusOutput]("시스템 정보 조회 실패: %s", err.Error()), nil
		}

		var info map[string]any
		if err := json.Unmarshal(data, &info); err != nil {
			return toolErrorfFor[SystemStatusOutput]("응답 파싱 실패: %s", err.Error()), nil
		}

		return toolSuccessFor(SystemStatusOutput{Info: info}), nil
	}
}

// === Registration ===

func registerSystemTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_firmware_status",
		Description: "OPNsense 펌웨어 및 업데이트 상태를 조회합니다.",
	}, createFirmwareStatusHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_system_status",
		Description: "OPNsense 시스템 정보를 조회합니다. 버전, 아키텍처 등의 정보를 제공합니다.",
	}, createSystemStatusHandler(client))

	fmt.Println("OPNsense system tools registered: ops_firmware_status, ops_system_status")
}
