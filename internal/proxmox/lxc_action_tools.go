package proxmox

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types for LXC Actions ===

// LXCStartInput contains parameters for starting an LXC container
type LXCStartInput struct {
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
	VMID int    `json:"vmid" jsonschema:"LXC 컨테이너 ID"`
}

// LXCStartOutput contains the result of starting an LXC container
type LXCStartOutput struct {
	UPID    string `json:"upid"`
	Message string `json:"message"`
}

// LXCStopInput contains parameters for stopping an LXC container
type LXCStopInput struct {
	Node    string `json:"node" jsonschema:"Proxmox 노드 이름"`
	VMID    int    `json:"vmid" jsonschema:"LXC 컨테이너 ID"`
	Confirm bool   `json:"confirm" jsonschema:"true로 설정해야 실제 정지 실행"`
}

// LXCStopOutput contains the result of stopping an LXC container
type LXCStopOutput struct {
	UPID    string `json:"upid,omitempty"`
	Message string `json:"message"`
}

// LXCShutdownInput contains parameters for gracefully shutting down an LXC container
type LXCShutdownInput struct {
	Node    string `json:"node" jsonschema:"Proxmox 노드 이름"`
	VMID    int    `json:"vmid" jsonschema:"LXC 컨테이너 ID"`
	Timeout int    `json:"timeout,omitempty" jsonschema:"종료 대기 시간 (초)"`
}

// LXCShutdownOutput contains the result of shutting down an LXC container
type LXCShutdownOutput struct {
	UPID    string `json:"upid"`
	Message string `json:"message"`
}

// LXCRebootInput contains parameters for rebooting an LXC container
type LXCRebootInput struct {
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
	VMID int    `json:"vmid" jsonschema:"LXC 컨테이너 ID"`
}

// LXCRebootOutput contains the result of rebooting an LXC container
type LXCRebootOutput struct {
	UPID    string `json:"upid"`
	Message string `json:"message"`
}

// === Tool Handlers ===

func createLXCStartHandler(client *Client) mcp.ToolHandlerFor[LXCStartInput, LXCStartOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[LXCStartInput]) (*mcp.CallToolResultFor[LXCStartOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[LXCStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[LXCStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[LXCStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("노드 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		ct, err := node.Container(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[LXCStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("LXC 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := ct.Start(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[LXCStartOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("LXC 시작 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[LXCStartOutput]{
			StructuredContent: LXCStartOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("LXC %d 시작 요청됨 (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

func createLXCStopHandler(client *Client) mcp.ToolHandlerFor[LXCStopInput, LXCStopOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[LXCStopInput]) (*mcp.CallToolResultFor[LXCStopOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[LXCStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[LXCStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}
		if !input.Confirm {
			return &mcp.CallToolResultFor[LXCStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "⚠️ LXC 강제 정지는 데이터 손실을 유발할 수 있습니다. confirm: true를 설정해주세요. 정상 종료를 원하면 pve_lxc_shutdown을 사용하세요."}},
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[LXCStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("노드 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		ct, err := node.Container(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[LXCStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("LXC 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := ct.Stop(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[LXCStopOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("LXC 정지 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[LXCStopOutput]{
			StructuredContent: LXCStopOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("LXC %d 정지 요청됨 (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

func createLXCShutdownHandler(client *Client) mcp.ToolHandlerFor[LXCShutdownInput, LXCShutdownOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[LXCShutdownInput]) (*mcp.CallToolResultFor[LXCShutdownOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[LXCShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[LXCShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[LXCShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("노드 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		ct, err := node.Container(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[LXCShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("LXC 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		// Shutdown with optional timeout
		timeout := input.Timeout
		if timeout == 0 {
			timeout = 30
		}
		task, err := ct.Shutdown(ctx, false, timeout)
		if err != nil {
			return &mcp.CallToolResultFor[LXCShutdownOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("LXC 종료 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[LXCShutdownOutput]{
			StructuredContent: LXCShutdownOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("LXC %d 정상 종료 요청됨 (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

func createLXCRebootHandler(client *Client) mcp.ToolHandlerFor[LXCRebootInput, LXCRebootOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[LXCRebootInput]) (*mcp.CallToolResultFor[LXCRebootOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return &mcp.CallToolResultFor[LXCRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "node 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}
		if input.VMID == 0 {
			return &mcp.CallToolResultFor[LXCRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: "vmid 파라미터가 필요합니다"}},
				IsError: true,
			}, nil
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return &mcp.CallToolResultFor[LXCRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("노드 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		ct, err := node.Container(ctx, input.VMID)
		if err != nil {
			return &mcp.CallToolResultFor[LXCRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("LXC 조회 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		task, err := ct.Reboot(ctx)
		if err != nil {
			return &mcp.CallToolResultFor[LXCRebootOutput]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("LXC 재시작 실패: %s", err.Error())}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResultFor[LXCRebootOutput]{
			StructuredContent: LXCRebootOutput{
				UPID:    string(task.UPID),
				Message: fmt.Sprintf("LXC %d 재시작 요청됨 (UPID: %s)", input.VMID, string(task.UPID)),
			},
		}, nil
	}
}

// === Registration ===

func registerLXCActionTools(server *mcp.Server, client *Client) {
	// pve_lxc_start
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_lxc_start",
		Description: "LXC 컨테이너를 시작합니다. 비동기 작업으로 UPID가 반환됩니다.",
	}, createLXCStartHandler(client))

	// pve_lxc_stop
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_lxc_stop",
		Description: "LXC 컨테이너를 강제 정지합니다. 데이터 손실 위험이 있으므로 confirm: true가 필요합니다.",
	}, createLXCStopHandler(client))

	// pve_lxc_shutdown
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_lxc_shutdown",
		Description: "LXC 컨테이너를 정상 종료합니다.",
	}, createLXCShutdownHandler(client))

	// pve_lxc_reboot
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_lxc_reboot",
		Description: "LXC 컨테이너를 재시작합니다.",
	}, createLXCRebootHandler(client))

	fmt.Println("LXC Action tools registered: pve_lxc_start, pve_lxc_stop, pve_lxc_shutdown, pve_lxc_reboot")
}
