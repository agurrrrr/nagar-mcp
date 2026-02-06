package proxmox

import (
	"context"
	"fmt"

	"github.com/luthermonson/go-proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

// TaskStatusInput contains upid
type TaskStatusInput struct {
	UPID string `json:"upid" jsonschema:"태스크 UPID"`
}

// TaskStatusOutput contains task status
type TaskStatusOutput struct {
	UPID       string `json:"upid"`
	Node       string `json:"node"`
	Status     string `json:"status"`
	ExitStatus string `json:"exitstatus"`
	Type       string `json:"type"`
	User       string `json:"user"`
	IsRunning  bool   `json:"is_running"`
	IsComplete bool   `json:"is_complete"`
	IsFailed   bool   `json:"is_failed"`
}

// TaskLogInput contains upid
type TaskLogInput struct {
	UPID  string `json:"upid" jsonschema:"태스크 UPID"`
	Start int    `json:"start,omitempty" jsonschema:"시작 라인 (기본 0)"`
	Limit int    `json:"limit,omitempty" jsonschema:"최대 라인 수 (기본 50)"`
}

// TaskLogOutput contains task log
type TaskLogOutput struct {
	UPID  string        `json:"upid"`
	Lines []TaskLogLine `json:"lines"`
	Total int           `json:"total"`
}

// TaskLogLine represents a single log line
type TaskLogLine struct {
	N int    `json:"n"`
	T string `json:"t"`
}

// === Tool Handlers ===

func createTaskStatusHandler(client *Client) mcp.ToolHandlerFor[TaskStatusInput, TaskStatusOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[TaskStatusInput]) (*mcp.CallToolResultFor[TaskStatusOutput], error) {
		input := params.Arguments

		if input.UPID == "" {
			return toolErrorFor[TaskStatusOutput]("upid 파라미터가 필요합니다"), nil
		}

		// Create task from UPID
		task := proxmox.NewTask(proxmox.UPID(input.UPID), client.Client)

		// Ping the task to update its status
		err := task.Ping(ctx)
		if err != nil {
			return toolErrorfFor[TaskStatusOutput]("태스크 상태 조회 실패: %s", err.Error()), nil
		}

		output := TaskStatusOutput{
			UPID:       string(task.UPID),
			Node:       task.Node,
			Status:     task.Status,
			ExitStatus: task.ExitStatus,
			Type:       task.Type,
			User:       task.User,
			IsRunning:  task.IsRunning,
			IsComplete: task.IsCompleted,
			IsFailed:   task.IsFailed,
		}

		return toolSuccessFor(output), nil
	}
}

func createTaskLogHandler(client *Client) mcp.ToolHandlerFor[TaskLogInput, TaskLogOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[TaskLogInput]) (*mcp.CallToolResultFor[TaskLogOutput], error) {
		input := params.Arguments

		if input.UPID == "" {
			return toolErrorFor[TaskLogOutput]("upid 파라미터가 필요합니다"), nil
		}

		// Create task from UPID
		task := proxmox.NewTask(proxmox.UPID(input.UPID), client.Client)

		// Set defaults
		start := input.Start
		limit := input.Limit
		if limit == 0 {
			limit = 50
		}

		logs, err := task.Log(ctx, start, limit)
		if err != nil {
			return toolErrorfFor[TaskLogOutput]("태스크 로그 조회 실패: %s", err.Error()), nil
		}

		output := TaskLogOutput{
			UPID:  string(task.UPID),
			Lines: make([]TaskLogLine, 0, len(logs)),
			Total: len(logs),
		}

		for n, text := range logs {
			output.Lines = append(output.Lines, TaskLogLine{
				N: n,
				T: text,
			})
		}

		return toolSuccessFor(output), nil
	}
}

// === Registration ===

func registerTaskTools(server *mcp.Server, client *Client) {
	// pve_task_status
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_task_status",
		Description: "Proxmox 비동기 태스크의 상태를 조회합니다. VM 시작/정지 등의 작업은 비동기로 실행되며, UPID를 통해 상태를 확인할 수 있습니다.",
	}, createTaskStatusHandler(client))

	// pve_task_log
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_task_log",
		Description: "Proxmox 태스크의 로그를 조회합니다. 태스크 실행 중 발생한 메시지와 결과를 확인할 수 있습니다.",
	}, createTaskLogHandler(client))

	fmt.Println("Task tools registered: pve_task_status, pve_task_log")
}
