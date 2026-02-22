# nagar-mcp

[한국어](README_KO.md)

A Model Context Protocol (MCP) server for managing homelab infrastructure. Provides unified AI-assisted control over **Proxmox VE**, **Kubernetes**, and **OPNsense** through 84 tools.

## Features

| Module | Tools | Capabilities |
|--------|-------|--------------|
| **Proxmox VE** | 34 | VMs, LXC containers, Ceph, storage, tasks, monitoring (RRD) |
| **Kubernetes** | 17 | Pods, deployments, services, namespaces, nodes, events, scaling |
| **OPNsense** | 33 | Firewall, NAT, DNS, DHCP, WireGuard VPN, HAProxy, services |

## Quick Start

### Prerequisites

- Go 1.23+
- Access to at least one of: Proxmox VE, Kubernetes cluster, OPNsense

### Install

```bash
go install github.com/agurrrrr/nagar-mcp/cmd/nagar-mcp@latest
```

Or build from source:

```bash
git clone https://github.com/agurrrrr/nagar-mcp.git
cd nagar-mcp
go build -o nagar-mcp ./cmd/nagar-mcp/
```

### Configure

```bash
cp config.example.yaml config.yaml
# Edit config.yaml with your credentials
```

Each module is optional. Leave a section empty to disable it.

### Run

```bash
# stdio mode (default, for MCP clients like Claude)
./nagar-mcp -config config.yaml

# HTTP mode
./nagar-mcp -config config.yaml -transport http
```

## Configuration

### Proxmox VE

Create an API token in Proxmox: Datacenter > Permissions > API Tokens.

```yaml
proxmox:
  url: "https://your-proxmox-host:8006/api2/json"
  token_id: "user@pam!token-name"
  secret: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  insecure_tls: false
```

Environment variable overrides: `PVE_URL`, `PVE_TOKEN_ID`, `PVE_SECRET`, `PVE_INSECURE_TLS`

### Kubernetes

```yaml
kubernetes:
  kubeconfig: "~/.kube/config"
  context: ""  # leave empty for current-context
```

Environment variable overrides: `KUBECONFIG`, `KUBE_CONTEXT`

### OPNsense

Create an API key in OPNsense: System > Access > Users > (your user) > API keys.

```yaml
opnsense:
  url: "https://your-opnsense-host"
  api_key: "your-api-key"
  api_secret: "your-api-secret"
  insecure_tls: false
```

Environment variable overrides: `OPS_URL`, `OPS_API_KEY`, `OPS_API_SECRET`, `OPS_INSECURE_TLS`

## Claude Desktop Integration

Add to your Claude Desktop config (`~/.claude/settings.json`):

```json
{
  "mcpServers": {
    "nagar-mcp": {
      "command": "nagar-mcp",
      "args": ["-config", "/path/to/config.yaml"]
    }
  }
}
```

## Tool Reference

<details>
<summary><strong>Proxmox VE (34 tools)</strong></summary>

**Nodes**: `pve_list_nodes`, `pve_node_status`

**VMs**: `pve_list_vms`, `pve_vm_status`, `pve_vm_config`, `pve_vm_start`, `pve_vm_stop`, `pve_vm_shutdown`, `pve_vm_reboot`, `pve_vm_snapshot_list`, `pve_vm_snapshot_create`

**LXC Containers**: `pve_list_lxc`, `pve_lxc_status`, `pve_lxc_config`, `pve_lxc_start`, `pve_lxc_stop`, `pve_lxc_shutdown`, `pve_lxc_reboot`

**Cluster & Storage**: `pve_cluster_resources`, `pve_cluster_status`, `pve_list_storage`, `pve_storage_content`

**Ceph**: `pve_ceph_status`, `pve_list_ceph_osd`, `pve_list_ceph_pool`, `pve_list_ceph_mon`, `pve_list_ceph_mds`, `pve_list_ceph_fs`

**Monitoring (RRD)**: `pve_node_rrd`, `pve_vm_rrd`, `pve_lxc_rrd`, `pve_storage_rrd`

**Tasks**: `pve_task_status`, `pve_task_log`

</details>

<details>
<summary><strong>Kubernetes (17 tools)</strong></summary>

**Cluster**: `k8s_cluster_info`, `k8s_list_namespaces`, `k8s_list_nodes`, `k8s_node_status`

**Pods**: `k8s_list_pods`, `k8s_pod_status`, `k8s_pod_logs`

**Deployments**: `k8s_list_deployments`, `k8s_deployment_status`

**Services**: `k8s_list_services`, `k8s_service_status`

**Events**: `k8s_list_events`

**Actions**: `k8s_scale_deployment`, `k8s_restart_deployment`, `k8s_delete_pod`, `k8s_cordon_node`, `k8s_uncordon_node`

</details>

<details>
<summary><strong>OPNsense (33 tools)</strong></summary>

**System**: `ops_firmware_status`, `ops_system_status`

**Firewall**: `ops_list_firewall_rules`, `ops_get_firewall_rule`, `ops_add_firewall_rule`, `ops_toggle_firewall_rule`, `ops_delete_firewall_rule`

**NAT / Port Forwarding**: `ops_list_nat_rules`, `ops_get_nat_rule`, `ops_add_nat_rule`, `ops_update_nat_rule`, `ops_delete_nat_rule`, `ops_toggle_nat_rule`

**DNS (Unbound)**: `ops_list_dns_overrides`, `ops_add_dns_override`, `ops_delete_dns_override`, `ops_dns_stats`

**DHCP**: `ops_list_dhcp_leases`

**WireGuard VPN**: `ops_wg_list_instances`, `ops_wg_add_instance`, `ops_wg_list_peers`, `ops_wg_add_peer`, `ops_wg_delete_peer`

**HAProxy**: `ops_ha_list_servers`, `ops_ha_add_server`, `ops_ha_list_backends`, `ops_ha_add_backend`, `ops_ha_list_frontends`, `ops_ha_statistics`, `ops_ha_config_test`

**Services**: `ops_list_services`, `ops_service_action`, `ops_apply_changes`

</details>

## License

MIT License - see [LICENSE](LICENSE) for details.

---

> **"I have filled him with the Spirit of God, with wisdom, with understanding, with knowledge and with all kinds of skills, to make artistic designs for work in gold, silver and bronze"**
>
> *- Exodus 31:3-4 (NIV)*
