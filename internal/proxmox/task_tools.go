package proxmox

import (
	"context"
	"log/slog"

	"github.com/luthermonson/go-proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types ===

// TaskStatusInput contains upid
type TaskStatusInput struct {
	UPID string `json:"upid" jsonschema:"Task UPID"`
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
	UPID  string `json:"upid" jsonschema:"Task UPID"`
	Start int    `json:"start,omitempty" jsonschema:"Start line (default 0)"`
	Limit int    `json:"limit,omitempty" jsonschema:"Maximum number of lines (default 50)"`
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
			return toolErrorFor[TaskStatusOutput]("upid parameter is required"), nil
		}

		// Create task from UPID
		task := proxmox.NewTask(proxmox.UPID(input.UPID), client.Client)

		// Ping the task to update its status
		err := task.Ping(ctx)
		if err != nil {
			return toolErrorfFor[TaskStatusOutput]("failed to query task status: %s", err.Error()), nil
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
			return toolErrorFor[TaskLogOutput]("upid parameter is required"), nil
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
			return toolErrorfFor[TaskLogOutput]("failed to query task log: %s", err.Error()), nil
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
		Description: "Retrieves the status of a Proxmox async task. Operations like VM start/stop run asynchronously, and their status can be checked via UPID.",
	}, createTaskStatusHandler(client))

	// pve_task_log
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_task_log",
		Description: "Retrieves logs of a Proxmox task. View messages and results generated during task execution.",
	}, createTaskLogHandler(client))

	slog.Info("Task tools registered: pve_task_status, pve_task_log")
}
