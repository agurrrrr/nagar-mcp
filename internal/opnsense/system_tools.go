package opnsense

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

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
			return toolErrorfFor[FirmwareStatusOutput]("failed to query firmware status: %s", err.Error()), nil
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return toolErrorfFor[FirmwareStatusOutput]("failed to parse response: %s", err.Error()), nil
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
			return toolErrorfFor[SystemStatusOutput]("failed to query system info: %s", err.Error()), nil
		}

		var info map[string]any
		if err := json.Unmarshal(data, &info); err != nil {
			return toolErrorfFor[SystemStatusOutput]("failed to parse response: %s", err.Error()), nil
		}

		return toolSuccessFor(SystemStatusOutput{Info: info}), nil
	}
}

// === Registration ===

func registerSystemTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_firmware_status",
		Description: "Queries OPNsense firmware and update status.",
	}, createFirmwareStatusHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ops_system_status",
		Description: "Queries OPNsense system information. Provides version, architecture, and other details.",
	}, createSystemStatusHandler(client))

	slog.Info("OPNsense system tools registered: ops_firmware_status, ops_system_status")
}
