package proxmox

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// === Ceph Input/Output Types ===

// CephStatusInput contains the node name
type CephStatusInput struct {
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
}

// CephStatusOutput contains Ceph cluster status
type CephStatusOutput struct {
	Health      CephHealthOutput `json:"health"`
	PGMap       CephPGMapOutput  `json:"pgmap"`
	OSDMap      CephOSDMapOutput `json:"osdmap"`
	MonMap      CephMonMapOutput `json:"monmap"`
	QuorumNames []string         `json:"quorum_names"`
	ClusterFSID string           `json:"cluster_fsid"`
}

// CephHealthOutput represents Ceph health status
type CephHealthOutput struct {
	Status string   `json:"status"`
	Checks []string `json:"checks,omitempty"`
}

// CephPGMapOutput represents placement group statistics
type CephPGMapOutput struct {
	PGsPerOSD        float64 `json:"pgs_per_osd"`
	NumPGs           int     `json:"num_pgs"`
	NumPools         int     `json:"num_pools"`
	NumObjects       int64   `json:"num_objects"`
	DataBytes        int64   `json:"data_bytes"`
	BytesUsed        int64   `json:"bytes_used"`
	BytesAvail       int64   `json:"bytes_avail"`
	BytesTotal       int64   `json:"bytes_total"`
	ReadBytesPerSec  int64   `json:"read_bytes_sec"`
	WriteBytesPerSec int64   `json:"write_bytes_sec"`
	ReadOpsPerSec    int64   `json:"read_op_per_sec"`
	WriteOpsPerSec   int64   `json:"write_op_per_sec"`
}

// CephOSDMapOutput represents OSD map summary
type CephOSDMapOutput struct {
	NumOSDs   int  `json:"num_osds"`
	NumUpOSDs int  `json:"num_up_osds"`
	NumInOSDs int  `json:"num_in_osds"`
	Full      bool `json:"full"`
	NearFull  bool `json:"nearfull"`
}

// CephMonMapOutput represents Monitor map summary
type CephMonMapOutput struct {
	NumMons int      `json:"num_mons"`
	Mons    []string `json:"mons,omitempty"`
}

// ListCephOSDInput contains the node name
type ListCephOSDInput struct {
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
}

// ListCephOSDOutput contains list of OSDs
type ListCephOSDOutput struct {
	OSDs []CephOSDDetail `json:"osds"`
}

// CephOSDDetail represents a single OSD with full details
type CephOSDDetail struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Host        string  `json:"host"`
	Status      string  `json:"status"`
	In          bool    `json:"in"`
	Up          bool    `json:"up"`
	DevicePath  string  `json:"device_path,omitempty"`
	Class       string  `json:"class,omitempty"`
	Crush       string  `json:"crush_location,omitempty"`
	UsedBytes   int64   `json:"used_bytes"`
	TotalBytes  int64   `json:"total_bytes"`
	Utilization float64 `json:"utilization"`
}

// ListCephPoolInput contains the node name
type ListCephPoolInput struct {
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
}

// ListCephPoolOutput contains list of Ceph pools
type ListCephPoolOutput struct {
	Pools []CephPoolDetail `json:"pools"`
}

// CephPoolDetail represents a single Ceph pool with full details
type CephPoolDetail struct {
	Name        string  `json:"name"`
	PoolID      int     `json:"pool_id"`
	Size        int     `json:"size"`
	MinSize     int     `json:"min_size"`
	PGNum       int     `json:"pg_num"`
	CrushRule   string  `json:"crush_rule,omitempty"`
	BytesUsed   int64   `json:"bytes_used"`
	PercentUsed float64 `json:"percent_used"`
}

// ListCephMonInput contains the node name
type ListCephMonInput struct {
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
}

// ListCephMonOutput contains list of Ceph monitors
type ListCephMonOutput struct {
	Monitors []CephMonDetail `json:"monitors"`
}

// CephMonDetail represents a single Ceph monitor with full details
type CephMonDetail struct {
	Name   string `json:"name"`
	Addr   string `json:"addr"`
	Host   string `json:"host"`
	Quorum bool   `json:"quorum"`
}

// ListCephMDSInput contains the node name
type ListCephMDSInput struct {
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
}

// ListCephMDSOutput contains list of MDS daemons
type ListCephMDSOutput struct {
	MDSDaemons []CephMDSDetail `json:"mds_daemons"`
}

// CephMDSDetail represents a single MDS daemon
type CephMDSDetail struct {
	Name  string `json:"name"`
	Host  string `json:"host"`
	State string `json:"state"`
	Addr  string `json:"addr"`
}

// ListCephFSInput contains the node name
type ListCephFSInput struct {
	Node string `json:"node" jsonschema:"Proxmox 노드 이름"`
}

// ListCephFSOutput contains list of CephFS filesystems
type ListCephFSOutput struct {
	Filesystems []CephFSDetail `json:"filesystems"`
}

// CephFSDetail represents a CephFS filesystem
type CephFSDetail struct {
	Name         string   `json:"name"`
	MetadataPool string   `json:"metadata_pool"`
	DataPools    []string `json:"data_pools"`
}

// === Raw API response types for Ceph ===

type cephStatusResponse struct {
	Health struct {
		Status string `json:"status"`
	} `json:"health"`
	PGMap struct {
		PGsPerOSD        float64 `json:"pgs_per_osd"`
		NumPGs           int     `json:"num_pgs"`
		NumPools         int     `json:"num_pools"`
		NumObjects       int64   `json:"num_objects"`
		DataBytes        int64   `json:"data_bytes"`
		BytesUsed        int64   `json:"bytes_used"`
		BytesAvail       int64   `json:"bytes_avail"`
		BytesTotal       int64   `json:"bytes_total"`
		ReadBytesPerSec  int64   `json:"read_bytes_sec"`
		WriteBytesPerSec int64   `json:"write_bytes_sec"`
		ReadOpsPerSec    int64   `json:"read_op_per_sec"`
		WriteOpsPerSec   int64   `json:"write_op_per_sec"`
	} `json:"pgmap"`
	OSDMap struct {
		NumOSDs   int  `json:"num_osds"`
		NumUpOSDs int  `json:"num_up_osds"`
		NumInOSDs int  `json:"num_in_osds"`
		Full      bool `json:"full"`
		NearFull  bool `json:"nearfull"`
	} `json:"osdmap"`
	MonMap struct {
		NumMons int `json:"num_mons"`
	} `json:"monmap"`
	QuorumNames []string `json:"quorum_names"`
	FSID        string   `json:"fsid"`
}

type cephOSDResponse struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Host        string  `json:"host"`
	Status      string  `json:"status"`
	In          int     `json:"in"`
	Up          int     `json:"up"`
	DevPath     string  `json:"devpath"`
	Class       string  `json:"class"`
	Crush       string  `json:"crush_location"`
	TotalSpace  int64   `json:"total_space"`
	UsedSpace   int64   `json:"used_space"`
	Utilization float64 `json:"utilization"`
}

type cephPoolResponse struct {
	PoolName    string  `json:"pool_name"`
	Pool        int     `json:"pool"`
	Size        int     `json:"size"`
	MinSize     int     `json:"min_size"`
	PGNum       int     `json:"pg_num"`
	CrushRule   string  `json:"crush_rule"`
	BytesUsed   int64   `json:"bytes_used"`
	PercentUsed float64 `json:"percent_used"`
}

type cephMonResponse struct {
	Name   string `json:"name"`
	Addr   string `json:"addr"`
	Host   string `json:"host"`
	Quorum int    `json:"quorum"`
}

type cephMDSResponse struct {
	Name  string `json:"name"`
	Host  string `json:"host"`
	State string `json:"state"`
	Addr  string `json:"addr"`
}

type cephFSResponse struct {
	Name         string   `json:"name"`
	MetadataPool string   `json:"metadata_pool"`
	DataPool     []string `json:"data_pool"`
}

// === Tool Handlers ===

func createCephStatusHandler(client *Client) mcp.ToolHandlerFor[CephStatusInput, CephStatusOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[CephStatusInput]) (*mcp.CallToolResultFor[CephStatusOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return toolErrorFor[CephStatusOutput]("node 파라미터가 필요합니다"), nil
		}

		var resp cephStatusResponse
		path := fmt.Sprintf("/nodes/%s/ceph/status", input.Node)
		if err := client.Get(ctx, path, &resp); err != nil {
			return toolErrorfFor[CephStatusOutput]("Ceph 상태 조회 실패: %s", err.Error()), nil
		}

		output := CephStatusOutput{
			Health: CephHealthOutput{
				Status: resp.Health.Status,
			},
			PGMap: CephPGMapOutput{
				PGsPerOSD:        resp.PGMap.PGsPerOSD,
				NumPGs:           resp.PGMap.NumPGs,
				NumPools:         resp.PGMap.NumPools,
				NumObjects:       resp.PGMap.NumObjects,
				DataBytes:        resp.PGMap.DataBytes,
				BytesUsed:        resp.PGMap.BytesUsed,
				BytesAvail:       resp.PGMap.BytesAvail,
				BytesTotal:       resp.PGMap.BytesTotal,
				ReadBytesPerSec:  resp.PGMap.ReadBytesPerSec,
				WriteBytesPerSec: resp.PGMap.WriteBytesPerSec,
				ReadOpsPerSec:    resp.PGMap.ReadOpsPerSec,
				WriteOpsPerSec:   resp.PGMap.WriteOpsPerSec,
			},
			OSDMap: CephOSDMapOutput{
				NumOSDs:   resp.OSDMap.NumOSDs,
				NumUpOSDs: resp.OSDMap.NumUpOSDs,
				NumInOSDs: resp.OSDMap.NumInOSDs,
				Full:      resp.OSDMap.Full,
				NearFull:  resp.OSDMap.NearFull,
			},
			MonMap: CephMonMapOutput{
				NumMons: resp.MonMap.NumMons,
			},
			QuorumNames: resp.QuorumNames,
			ClusterFSID: resp.FSID,
		}

		return toolSuccessFor(output), nil
	}
}

func createListCephOSDHandler(client *Client) mcp.ToolHandlerFor[ListCephOSDInput, ListCephOSDOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListCephOSDInput]) (*mcp.CallToolResultFor[ListCephOSDOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return toolErrorFor[ListCephOSDOutput]("node 파라미터가 필요합니다"), nil
		}

		var osds []cephOSDResponse
		path := fmt.Sprintf("/nodes/%s/ceph/osd", input.Node)
		if err := client.Get(ctx, path, &osds); err != nil {
			return toolErrorfFor[ListCephOSDOutput]("Ceph OSD 조회 실패: %s", err.Error()), nil
		}

		output := ListCephOSDOutput{
			OSDs: make([]CephOSDDetail, 0, len(osds)),
		}

		for _, osd := range osds {
			output.OSDs = append(output.OSDs, CephOSDDetail{
				ID:          osd.ID,
				Name:        osd.Name,
				Host:        osd.Host,
				Status:      osd.Status,
				In:          osd.In == 1,
				Up:          osd.Up == 1,
				DevicePath:  osd.DevPath,
				Class:       osd.Class,
				Crush:       osd.Crush,
				UsedBytes:   osd.UsedSpace,
				TotalBytes:  osd.TotalSpace,
				Utilization: osd.Utilization,
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createListCephPoolHandler(client *Client) mcp.ToolHandlerFor[ListCephPoolInput, ListCephPoolOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListCephPoolInput]) (*mcp.CallToolResultFor[ListCephPoolOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return toolErrorFor[ListCephPoolOutput]("node 파라미터가 필요합니다"), nil
		}

		var pools []cephPoolResponse
		path := fmt.Sprintf("/nodes/%s/ceph/pool", input.Node)
		if err := client.Get(ctx, path, &pools); err != nil {
			return toolErrorfFor[ListCephPoolOutput]("Ceph Pool 조회 실패: %s", err.Error()), nil
		}

		output := ListCephPoolOutput{
			Pools: make([]CephPoolDetail, 0, len(pools)),
		}

		for _, pool := range pools {
			output.Pools = append(output.Pools, CephPoolDetail{
				Name:        pool.PoolName,
				PoolID:      pool.Pool,
				Size:        pool.Size,
				MinSize:     pool.MinSize,
				PGNum:       pool.PGNum,
				CrushRule:   pool.CrushRule,
				BytesUsed:   pool.BytesUsed,
				PercentUsed: pool.PercentUsed,
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createListCephMonHandler(client *Client) mcp.ToolHandlerFor[ListCephMonInput, ListCephMonOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListCephMonInput]) (*mcp.CallToolResultFor[ListCephMonOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return toolErrorFor[ListCephMonOutput]("node 파라미터가 필요합니다"), nil
		}

		var mons []cephMonResponse
		path := fmt.Sprintf("/nodes/%s/ceph/mon", input.Node)
		if err := client.Get(ctx, path, &mons); err != nil {
			return toolErrorfFor[ListCephMonOutput]("Ceph Monitor 조회 실패: %s", err.Error()), nil
		}

		output := ListCephMonOutput{
			Monitors: make([]CephMonDetail, 0, len(mons)),
		}

		for _, mon := range mons {
			output.Monitors = append(output.Monitors, CephMonDetail{
				Name:   mon.Name,
				Addr:   mon.Addr,
				Host:   mon.Host,
				Quorum: mon.Quorum == 1,
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createListCephMDSHandler(client *Client) mcp.ToolHandlerFor[ListCephMDSInput, ListCephMDSOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListCephMDSInput]) (*mcp.CallToolResultFor[ListCephMDSOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return toolErrorFor[ListCephMDSOutput]("node 파라미터가 필요합니다"), nil
		}

		var mdsList []cephMDSResponse
		path := fmt.Sprintf("/nodes/%s/ceph/mds", input.Node)
		if err := client.Get(ctx, path, &mdsList); err != nil {
			return toolErrorfFor[ListCephMDSOutput]("Ceph MDS 조회 실패: %s", err.Error()), nil
		}

		output := ListCephMDSOutput{
			MDSDaemons: make([]CephMDSDetail, 0, len(mdsList)),
		}

		for _, mds := range mdsList {
			output.MDSDaemons = append(output.MDSDaemons, CephMDSDetail{
				Name:  mds.Name,
				Host:  mds.Host,
				State: mds.State,
				Addr:  mds.Addr,
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createListCephFSHandler(client *Client) mcp.ToolHandlerFor[ListCephFSInput, ListCephFSOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListCephFSInput]) (*mcp.CallToolResultFor[ListCephFSOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return toolErrorFor[ListCephFSOutput]("node 파라미터가 필요합니다"), nil
		}

		var fsList []cephFSResponse
		path := fmt.Sprintf("/nodes/%s/ceph/fs", input.Node)
		if err := client.Get(ctx, path, &fsList); err != nil {
			return toolErrorfFor[ListCephFSOutput]("CephFS 조회 실패: %s", err.Error()), nil
		}

		output := ListCephFSOutput{
			Filesystems: make([]CephFSDetail, 0, len(fsList)),
		}

		for _, fs := range fsList {
			output.Filesystems = append(output.Filesystems, CephFSDetail{
				Name:         fs.Name,
				MetadataPool: fs.MetadataPool,
				DataPools:    fs.DataPool,
			})
		}

		return toolSuccessFor(output), nil
	}
}

// === Registration ===

func registerCephTools(server *mcp.Server, client *Client) {
	// pve_ceph_status
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_ceph_status",
		Description: "Ceph 클러스터의 전체 상태를 조회합니다. 건강 상태, PG, OSD, Monitor 정보를 포함합니다.",
	}, createCephStatusHandler(client))

	// pve_list_ceph_osd
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_list_ceph_osd",
		Description: "Ceph OSD 목록과 상태를 조회합니다. 각 OSD의 용량, 사용률, 상태 정보를 포함합니다.",
	}, createListCephOSDHandler(client))

	// pve_list_ceph_pool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_list_ceph_pool",
		Description: "Ceph Pool 목록을 조회합니다. 각 풀의 크기, PG 수, 사용량 정보를 포함합니다.",
	}, createListCephPoolHandler(client))

	// pve_list_ceph_mon
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_list_ceph_mon",
		Description: "Ceph Monitor 목록을 조회합니다. 각 모니터의 주소와 쿼럼 상태를 포함합니다.",
	}, createListCephMonHandler(client))

	// pve_list_ceph_mds
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_list_ceph_mds",
		Description: "Ceph MDS (Metadata Server) 목록을 조회합니다. CephFS를 위한 MDS 상태를 확인합니다.",
	}, createListCephMDSHandler(client))

	// pve_list_ceph_fs
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pve_list_ceph_fs",
		Description: "CephFS 파일시스템 목록을 조회합니다.",
	}, createListCephFSHandler(client))

	fmt.Println("Ceph tools registered: pve_ceph_status, pve_list_ceph_osd, pve_list_ceph_pool, pve_list_ceph_mon, pve_list_ceph_mds, pve_list_ceph_fs")
}
