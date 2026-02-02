package proxmox

// NodeInfo represents basic node information
type NodeInfo struct {
	Node   string  `json:"node"`
	Status string  `json:"status"`
	CPU    float64 `json:"cpu"`
	MaxCPU int     `json:"maxcpu"`
	Mem    int64   `json:"mem"`
	MaxMem int64   `json:"maxmem"`
	Disk   int64   `json:"disk"`
	MaxDisk int64  `json:"maxdisk"`
	Uptime int64   `json:"uptime"`
}

// VMInfo represents basic VM information
type VMInfo struct {
	VMID      int     `json:"vmid"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	CPU       float64 `json:"cpu"`
	MaxCPU    int     `json:"maxcpu"`
	Mem       int64   `json:"mem"`
	MaxMem    int64   `json:"maxmem"`
	Disk      int64   `json:"disk"`
	MaxDisk   int64   `json:"maxdisk"`
	Uptime    int64   `json:"uptime"`
	NetIn     int64   `json:"netin"`
	NetOut    int64   `json:"netout"`
	DiskRead  int64   `json:"diskread"`
	DiskWrite int64   `json:"diskwrite"`
}

// ContainerInfo represents basic LXC container information
type ContainerInfo struct {
	VMID    int    `json:"vmid"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	CPUs    int    `json:"cpus"`
	MaxMem  int64  `json:"maxmem"`
	MaxDisk int64  `json:"maxdisk"`
	MaxSwap int64  `json:"maxswap"`
	Uptime  int64  `json:"uptime"`
}

// StorageInfo represents storage information
type StorageInfo struct {
	Storage  string  `json:"storage"`
	Type     string  `json:"type"`
	Content  string  `json:"content"`
	Active   int     `json:"active"`
	Enabled  int     `json:"enabled"`
	Shared   int     `json:"shared"`
	Used     int64   `json:"used"`
	Avail    int64   `json:"avail"`
	Total    int64   `json:"total"`
	UsedFraction float64 `json:"used_fraction"`
}

// TaskInfo represents task information
type TaskInfo struct {
	UPID      string `json:"upid"`
	Node      string `json:"node"`
	Status    string `json:"status"`
	ExitStatus string `json:"exitstatus"`
	Type      string `json:"type"`
	User      string `json:"user"`
	StartTime int64  `json:"starttime"`
	EndTime   int64  `json:"endtime,omitempty"`
}

// SnapshotInfo represents VM/LXC snapshot information
type SnapshotInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SnapTime    int64  `json:"snaptime,omitempty"`
	Parent      string `json:"parent,omitempty"`
}

// ClusterResourceInfo represents a cluster resource
type ClusterResourceInfo struct {
	ID       string  `json:"id"`
	Type     string  `json:"type"`
	Node     string  `json:"node"`
	Status   string  `json:"status"`
	Name     string  `json:"name,omitempty"`
	VMID     int     `json:"vmid,omitempty"`
	CPU      float64 `json:"cpu,omitempty"`
	MaxCPU   int     `json:"maxcpu,omitempty"`
	Mem      int64   `json:"mem,omitempty"`
	MaxMem   int64   `json:"maxmem,omitempty"`
	Disk     int64   `json:"disk,omitempty"`
	MaxDisk  int64   `json:"maxdisk,omitempty"`
	Uptime   int64   `json:"uptime,omitempty"`
}

// RRDDataPoint represents a single RRD data point
type RRDDataPoint struct {
	Time      int64   `json:"time"`
	CPU       float64 `json:"cpu,omitempty"`
	MaxCPU    int     `json:"maxcpu,omitempty"`
	Mem       int64   `json:"mem,omitempty"`
	MaxMem    int64   `json:"maxmem,omitempty"`
	NetIn     int64   `json:"netin,omitempty"`
	NetOut    int64   `json:"netout,omitempty"`
	DiskRead  int64   `json:"diskread,omitempty"`
	DiskWrite int64   `json:"diskwrite,omitempty"`
	Used      int64   `json:"used,omitempty"`
	Total     int64   `json:"total,omitempty"`
}

// CephStatusInfo represents Ceph cluster status
type CephStatusInfo struct {
	Health  CephHealthInfo  `json:"health"`
	MonMap  CephMonMapInfo  `json:"monmap"`
	OSDMap  CephOSDMapInfo  `json:"osdmap"`
	PGMap   CephPGMapInfo   `json:"pgmap"`
}

// CephHealthInfo represents Ceph health information
type CephHealthInfo struct {
	Status string                 `json:"status"`
	Checks map[string]interface{} `json:"checks,omitempty"`
}

// CephMonMapInfo represents Ceph monitor map
type CephMonMapInfo struct {
	Epoch int             `json:"epoch"`
	Mons  []CephMonInfo   `json:"mons"`
}

// CephMonInfo represents a single Ceph monitor
type CephMonInfo struct {
	Name string `json:"name"`
	Rank int    `json:"rank"`
	Addr string `json:"addr"`
}

// CephOSDMapInfo represents Ceph OSD map summary
type CephOSDMapInfo struct {
	NumOSDs    int `json:"num_osds"`
	NumUpOSDs  int `json:"num_up_osds"`
	NumInOSDs  int `json:"num_in_osds"`
}

// CephPGMapInfo represents Ceph placement group map summary
type CephPGMapInfo struct {
	NumPGs     int   `json:"num_pgs"`
	BytesUsed  int64 `json:"bytes_used"`
	BytesAvail int64 `json:"bytes_avail"`
	BytesTotal int64 `json:"bytes_total"`
}

// CephOSDInfo represents a single OSD
type CephOSDInfo struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Host   string `json:"host"`
	Class  string `json:"class"`
	Status string `json:"status"`
	In     bool   `json:"in"`
	Up     bool   `json:"up"`
}

// CephPoolInfo represents a Ceph pool
type CephPoolInfo struct {
	Name        string `json:"name"`
	Size        int    `json:"size"`
	MinSize     int    `json:"min_size"`
	PGNum       int    `json:"pg_num"`
	CrushRule   string `json:"crush_rule"`
	Application string `json:"application,omitempty"`
}
