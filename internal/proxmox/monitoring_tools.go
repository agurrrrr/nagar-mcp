package proxmox

import (
	"context"
	"fmt"

	"github.com/luthermonson/go-proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Input/Output Types for Monitoring ===

// NodeRRDInput contains parameters for node RRD data query
type NodeRRDInput struct {
	Node      string `json:"node" jsonschema:"description=Proxmox 노드 이름"`
	Timeframe string `json:"timeframe" jsonschema:"description=데이터 기간 (hour, day, week, month, year)"`
}

// NodeRRDOutput contains RRD data for a node
type NodeRRDOutput struct {
	Node       string              `json:"node"`
	Timeframe  string              `json:"timeframe"`
	DataPoints []RRDNodeDataPoint  `json:"data_points"`
}

// RRDNodeDataPoint represents a single RRD data point for a node
type RRDNodeDataPoint struct {
	Time       int64   `json:"time"`
	CPU        float64 `json:"cpu,omitempty"`
	IOWait     float64 `json:"iowait,omitempty"`
	LoadAvg    float64 `json:"loadavg,omitempty"`
	MemTotal   int64   `json:"memtotal,omitempty"`
	MemUsed    int64   `json:"memused,omitempty"`
	RootTotal  int64   `json:"roottotal,omitempty"`
	RootUsed   int64   `json:"rootused,omitempty"`
	SwapTotal  int64   `json:"swaptotal,omitempty"`
	SwapUsed   int64   `json:"swapused,omitempty"`
	NetIn      int64   `json:"netin,omitempty"`
	NetOut     int64   `json:"netout,omitempty"`
}

// VMRRDInput contains parameters for VM RRD data query
type VMRRDInput struct {
	Node      string `json:"node" jsonschema:"description=Proxmox 노드 이름"`
	VMID      int    `json:"vmid" jsonschema:"description=VM ID"`
	Timeframe string `json:"timeframe" jsonschema:"description=데이터 기간 (hour, day, week, month, year)"`
}

// VMRRDOutput contains RRD data for a VM
type VMRRDOutput struct {
	VMID       int                `json:"vmid"`
	Timeframe  string             `json:"timeframe"`
	DataPoints []RRDVMDataPoint   `json:"data_points"`
}

// RRDVMDataPoint represents a single RRD data point for a VM
type RRDVMDataPoint struct {
	Time      int64   `json:"time"`
	CPU       float64 `json:"cpu,omitempty"`
	MaxCPU    int     `json:"maxcpu,omitempty"`
	Mem       int64   `json:"mem,omitempty"`
	MaxMem    int64   `json:"maxmem,omitempty"`
	Disk      int64   `json:"disk,omitempty"`
	MaxDisk   int64   `json:"maxdisk,omitempty"`
	NetIn     int64   `json:"netin,omitempty"`
	NetOut    int64   `json:"netout,omitempty"`
	DiskRead  int64   `json:"diskread,omitempty"`
	DiskWrite int64   `json:"diskwrite,omitempty"`
}

// LXCRRDInput contains parameters for LXC RRD data query
type LXCRRDInput struct {
	Node      string `json:"node" jsonschema:"description=Proxmox 노드 이름"`
	VMID      int    `json:"vmid" jsonschema:"description=LXC 컨테이너 ID"`
	Timeframe string `json:"timeframe" jsonschema:"description=데이터 기간 (hour, day, week, month, year)"`
}

// LXCRRDOutput contains RRD data for an LXC container
type LXCRRDOutput struct {
	VMID       int               `json:"vmid"`
	Timeframe  string            `json:"timeframe"`
	DataPoints []RRDLXCDataPoint `json:"data_points"`
}

// RRDLXCDataPoint represents a single RRD data point for an LXC container
type RRDLXCDataPoint struct {
	Time      int64   `json:"time"`
	CPU       float64 `json:"cpu,omitempty"`
	MaxCPU    int     `json:"maxcpu,omitempty"`
	Mem       int64   `json:"mem,omitempty"`
	MaxMem    int64   `json:"maxmem,omitempty"`
	Disk      int64   `json:"disk,omitempty"`
	MaxDisk   int64   `json:"maxdisk,omitempty"`
	NetIn     int64   `json:"netin,omitempty"`
	NetOut    int64   `json:"netout,omitempty"`
	DiskRead  int64   `json:"diskread,omitempty"`
	DiskWrite int64   `json:"diskwrite,omitempty"`
}

// StorageRRDInput contains parameters for storage RRD data query
type StorageRRDInput struct {
	Node      string `json:"node" jsonschema:"description=Proxmox 노드 이름"`
	Storage   string `json:"storage" jsonschema:"description=스토리지 이름"`
	Timeframe string `json:"timeframe" jsonschema:"description=데이터 기간 (hour, day, week, month, year)"`
}

// StorageRRDOutput contains RRD data for storage
type StorageRRDOutput struct {
	Storage    string                  `json:"storage"`
	Timeframe  string                  `json:"timeframe"`
	DataPoints []RRDStorageDataPoint   `json:"data_points"`
}

// RRDStorageDataPoint represents a single RRD data point for storage
type RRDStorageDataPoint struct {
	Time  int64 `json:"time"`
	Used  int64 `json:"used,omitempty"`
	Total int64 `json:"total,omitempty"`
}

// === Raw API response types ===

type nodeRRDResponse struct {
	Time      uint64  `json:"time"`
	CPU       float64 `json:"cpu"`
	IOWait    float64 `json:"iowait"`
	LoadAvg   float64 `json:"loadavg"`
	MemTotal  uint64  `json:"memtotal"`
	MemUsed   uint64  `json:"memused"`
	RootTotal uint64  `json:"roottotal"`
	RootUsed  uint64  `json:"rootused"`
	SwapTotal uint64  `json:"swaptotal"`
	SwapUsed  uint64  `json:"swapused"`
	NetIn     uint64  `json:"netin"`
	NetOut    uint64  `json:"netout"`
}

type storageRRDResponse struct {
	Time  uint64 `json:"time"`
	Used  uint64 `json:"used"`
	Total uint64 `json:"total"`
}

// === Helper functions ===

func parseTimeframe(tf string) proxmox.Timeframe {
	switch tf {
	case "hour":
		return proxmox.TimeframeHour
	case "day":
		return proxmox.TimeframeDay
	case "week":
		return proxmox.TimeframeWeek
	case "month":
		return proxmox.TimeframeMonth
	case "year":
		return proxmox.TimeframeYear
	default:
		return proxmox.TimeframeHour
	}
}

// === Tool Handlers ===

func createNodeRRDHandler(client *Client) mcp.ToolHandlerFor[NodeRRDInput, NodeRRDOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[NodeRRDInput]) (*mcp.CallToolResultFor[NodeRRDOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return toolErrorFor[NodeRRDOutput]("node 파라미터가 필요합니다"), nil
		}

		timeframe := input.Timeframe
		if timeframe == "" {
			timeframe = "hour"
		}

		var data []nodeRRDResponse
		path := fmt.Sprintf("/nodes/%s/rrddata?timeframe=%s", input.Node, timeframe)
		if err := client.Get(ctx, path, &data); err != nil {
			return toolErrorfFor[NodeRRDOutput]("노드 RRD 데이터 조회 실패: %s", err.Error()), nil
		}

		output := NodeRRDOutput{
			Node:       input.Node,
			Timeframe:  timeframe,
			DataPoints: make([]RRDNodeDataPoint, 0, len(data)),
		}

		for _, d := range data {
			output.DataPoints = append(output.DataPoints, RRDNodeDataPoint{
				Time:      int64(d.Time),
				CPU:       d.CPU,
				IOWait:    d.IOWait,
				LoadAvg:   d.LoadAvg,
				MemTotal:  int64(d.MemTotal),
				MemUsed:   int64(d.MemUsed),
				RootTotal: int64(d.RootTotal),
				RootUsed:  int64(d.RootUsed),
				SwapTotal: int64(d.SwapTotal),
				SwapUsed:  int64(d.SwapUsed),
				NetIn:     int64(d.NetIn),
				NetOut:    int64(d.NetOut),
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createVMRRDHandler(client *Client) mcp.ToolHandlerFor[VMRRDInput, VMRRDOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[VMRRDInput]) (*mcp.CallToolResultFor[VMRRDOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return toolErrorFor[VMRRDOutput]("node 파라미터가 필요합니다"), nil
		}
		if input.VMID == 0 {
			return toolErrorFor[VMRRDOutput]("vmid 파라미터가 필요합니다"), nil
		}

		timeframe := input.Timeframe
		if timeframe == "" {
			timeframe = "hour"
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[VMRRDOutput]("노드 조회 실패: %s", err.Error()), nil
		}

		vm, err := node.VirtualMachine(ctx, input.VMID)
		if err != nil {
			return toolErrorfFor[VMRRDOutput]("VM 조회 실패: %s", err.Error()), nil
		}

		rrddata, err := vm.RRDData(ctx, parseTimeframe(timeframe))
		if err != nil {
			return toolErrorfFor[VMRRDOutput]("VM RRD 데이터 조회 실패: %s", err.Error()), nil
		}

		output := VMRRDOutput{
			VMID:       input.VMID,
			Timeframe:  timeframe,
			DataPoints: make([]RRDVMDataPoint, 0, len(rrddata)),
		}

		for _, d := range rrddata {
			output.DataPoints = append(output.DataPoints, RRDVMDataPoint{
				Time:    int64(d.Time),
				MaxCPU:  d.MaxCPU,
				MaxMem:  int64(d.MaxMem),
				Disk:    int64(d.Disk),
				MaxDisk: int64(d.MaxDisk),
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createLXCRRDHandler(client *Client) mcp.ToolHandlerFor[LXCRRDInput, LXCRRDOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[LXCRRDInput]) (*mcp.CallToolResultFor[LXCRRDOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return toolErrorFor[LXCRRDOutput]("node 파라미터가 필요합니다"), nil
		}
		if input.VMID == 0 {
			return toolErrorFor[LXCRRDOutput]("vmid 파라미터가 필요합니다"), nil
		}

		timeframe := input.Timeframe
		if timeframe == "" {
			timeframe = "hour"
		}

		node, err := client.Node(ctx, input.Node)
		if err != nil {
			return toolErrorfFor[LXCRRDOutput]("노드 조회 실패: %s", err.Error()), nil
		}

		ct, err := node.Container(ctx, input.VMID)
		if err != nil {
			return toolErrorfFor[LXCRRDOutput]("LXC 조회 실패: %s", err.Error()), nil
		}

		rrddata, err := ct.RRDData(ctx, parseTimeframe(timeframe), proxmox.AVERAGE)
		if err != nil {
			return toolErrorfFor[LXCRRDOutput]("LXC RRD 데이터 조회 실패: %s", err.Error()), nil
		}

		output := LXCRRDOutput{
			VMID:       input.VMID,
			Timeframe:  timeframe,
			DataPoints: make([]RRDLXCDataPoint, 0, len(rrddata)),
		}

		for _, d := range rrddata {
			output.DataPoints = append(output.DataPoints, RRDLXCDataPoint{
				Time:    int64(d.Time),
				MaxCPU:  d.MaxCPU,
				MaxMem:  int64(d.MaxMem),
				Disk:    int64(d.Disk),
				MaxDisk: int64(d.MaxDisk),
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createStorageRRDHandler(client *Client) mcp.ToolHandlerFor[StorageRRDInput, StorageRRDOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[StorageRRDInput]) (*mcp.CallToolResultFor[StorageRRDOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return toolErrorFor[StorageRRDOutput]("node 파라미터가 필요합니다"), nil
		}
		if input.Storage == "" {
			return toolErrorFor[StorageRRDOutput]("storage 파라미터가 필요합니다"), nil
		}

		timeframe := input.Timeframe
		if timeframe == "" {
			timeframe = "hour"
		}

		var data []storageRRDResponse
		path := fmt.Sprintf("/nodes/%s/storage/%s/rrddata?timeframe=%s", input.Node, input.Storage, timeframe)
		if err := client.Get(ctx, path, &data); err != nil {
			return toolErrorfFor[StorageRRDOutput]("스토리지 RRD 데이터 조회 실패: %s", err.Error()), nil
		}

		output := StorageRRDOutput{
			Storage:    input.Storage,
			Timeframe:  timeframe,
			DataPoints: make([]RRDStorageDataPoint, 0, len(data)),
		}

		for _, d := range data {
			output.DataPoints = append(output.DataPoints, RRDStorageDataPoint{
				Time:  int64(d.Time),
				Used:  int64(d.Used),
				Total: int64(d.Total),
			})
		}

		return toolSuccessFor(output), nil
	}
}

// === Registration ===

func registerMonitoringTools(server *mcp.Server, client *Client) {
	// pve_node_rrd
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_node_rrd",
		Description: "노드의 RRD 모니터링 데이터를 조회합니다. CPU, 메모리, 네트워크 사용량의 시계열 데이터를 제공합니다.",
	}, createNodeRRDHandler(client))

	// pve_vm_rrd
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_vm_rrd",
		Description: "VM의 RRD 모니터링 데이터를 조회합니다. CPU, 메모리, 디스크 사용량의 시계열 데이터를 제공합니다.",
	}, createVMRRDHandler(client))

	// pve_lxc_rrd
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_lxc_rrd",
		Description: "LXC 컨테이너의 RRD 모니터링 데이터를 조회합니다. CPU, 메모리, 디스크 사용량의 시계열 데이터를 제공합니다.",
	}, createLXCRRDHandler(client))

	// pve_storage_rrd
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_storage_rrd",
		Description: "스토리지의 RRD 모니터링 데이터를 조회합니다. 사용량의 시계열 데이터를 제공합니다.",
	}, createStorageRRDHandler(client))

	fmt.Println("Monitoring tools registered: pve_node_rrd, pve_vm_rrd, pve_lxc_rrd, pve_storage_rrd")
}
