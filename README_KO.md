# nagar-mcp

[English](README.md)

홈랩 인프라를 관리하기 위한 Model Context Protocol (MCP) 서버입니다. **Proxmox VE**, **Kubernetes**, **OPNsense**를 84개의 도구로 AI를 통해 통합 제어합니다.

## 기능

| 모듈 | 도구 수 | 기능 |
|------|---------|------|
| **Proxmox VE** | 34 | VM, LXC 컨테이너, Ceph, 스토리지, 태스크, 모니터링 (RRD) |
| **Kubernetes** | 17 | Pod, Deployment, Service, Namespace, Node, Event, 스케일링 |
| **OPNsense** | 33 | 방화벽, NAT, DNS, DHCP, WireGuard VPN, HAProxy, 서비스 |

## 빠른 시작

### 사전 요구사항

- Go 1.23+
- Proxmox VE, Kubernetes 클러스터, OPNsense 중 하나 이상에 대한 접근 권한

### 설치

```bash
go install github.com/agurrrrr/nagar-mcp/cmd/nagar-mcp@latest
```

또는 소스에서 빌드:

```bash
git clone https://github.com/agurrrrr/nagar-mcp.git
cd nagar-mcp
go build -o nagar-mcp ./cmd/nagar-mcp/
```

### 설정

```bash
cp config.example.yaml config.yaml
# config.yaml에 자격 증명을 입력하세요
```

각 모듈은 선택사항입니다. 해당 섹션을 비워두면 비활성화됩니다.

### 실행

```bash
# stdio 모드 (기본값, Claude 등 MCP 클라이언트용)
./nagar-mcp -config config.yaml

# HTTP 모드
./nagar-mcp -config config.yaml -transport http
```

## 설정 방법

### Proxmox VE

Proxmox에서 API 토큰 생성: Datacenter > Permissions > API Tokens

```yaml
proxmox:
  url: "https://your-proxmox-host:8006/api2/json"
  token_id: "user@pam!token-name"
  secret: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  insecure_tls: false
```

환경변수 오버라이드: `PVE_URL`, `PVE_TOKEN_ID`, `PVE_SECRET`, `PVE_INSECURE_TLS`

### Kubernetes

```yaml
kubernetes:
  kubeconfig: "~/.kube/config"
  context: ""  # 비워두면 current-context 사용
```

환경변수 오버라이드: `KUBECONFIG`, `KUBE_CONTEXT`

### OPNsense

OPNsense에서 API 키 생성: System > Access > Users > (사용자) > API keys

```yaml
opnsense:
  url: "https://your-opnsense-host"
  api_key: "your-api-key"
  api_secret: "your-api-secret"
  insecure_tls: false
```

환경변수 오버라이드: `OPS_URL`, `OPS_API_KEY`, `OPS_API_SECRET`, `OPS_INSECURE_TLS`

## Claude Desktop 연동

Claude Desktop 설정 파일 (`~/.claude/settings.json`)에 추가:

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

## 도구 목록

<details>
<summary><strong>Proxmox VE (34개 도구)</strong></summary>

**노드**: `pve_list_nodes`, `pve_node_status`

**VM**: `pve_list_vms`, `pve_vm_status`, `pve_vm_config`, `pve_vm_start`, `pve_vm_stop`, `pve_vm_shutdown`, `pve_vm_reboot`, `pve_vm_snapshot_list`, `pve_vm_snapshot_create`

**LXC 컨테이너**: `pve_list_lxc`, `pve_lxc_status`, `pve_lxc_config`, `pve_lxc_start`, `pve_lxc_stop`, `pve_lxc_shutdown`, `pve_lxc_reboot`

**클러스터 & 스토리지**: `pve_cluster_resources`, `pve_cluster_status`, `pve_list_storage`, `pve_storage_content`

**Ceph**: `pve_ceph_status`, `pve_list_ceph_osd`, `pve_list_ceph_pool`, `pve_list_ceph_mon`, `pve_list_ceph_mds`, `pve_list_ceph_fs`

**모니터링 (RRD)**: `pve_node_rrd`, `pve_vm_rrd`, `pve_lxc_rrd`, `pve_storage_rrd`

**태스크**: `pve_task_status`, `pve_task_log`

</details>

<details>
<summary><strong>Kubernetes (17개 도구)</strong></summary>

**클러스터**: `k8s_cluster_info`, `k8s_list_namespaces`, `k8s_list_nodes`, `k8s_node_status`

**Pod**: `k8s_list_pods`, `k8s_pod_status`, `k8s_pod_logs`

**Deployment**: `k8s_list_deployments`, `k8s_deployment_status`

**Service**: `k8s_list_services`, `k8s_service_status`

**Event**: `k8s_list_events`

**액션**: `k8s_scale_deployment`, `k8s_restart_deployment`, `k8s_delete_pod`, `k8s_cordon_node`, `k8s_uncordon_node`

</details>

<details>
<summary><strong>OPNsense (33개 도구)</strong></summary>

**시스템**: `ops_firmware_status`, `ops_system_status`

**방화벽**: `ops_list_firewall_rules`, `ops_get_firewall_rule`, `ops_add_firewall_rule`, `ops_toggle_firewall_rule`, `ops_delete_firewall_rule`

**NAT / 포트포워딩**: `ops_list_nat_rules`, `ops_get_nat_rule`, `ops_add_nat_rule`, `ops_update_nat_rule`, `ops_delete_nat_rule`, `ops_toggle_nat_rule`

**DNS (Unbound)**: `ops_list_dns_overrides`, `ops_add_dns_override`, `ops_delete_dns_override`, `ops_dns_stats`

**DHCP**: `ops_list_dhcp_leases`

**WireGuard VPN**: `ops_wg_list_instances`, `ops_wg_add_instance`, `ops_wg_list_peers`, `ops_wg_add_peer`, `ops_wg_delete_peer`

**HAProxy**: `ops_ha_list_servers`, `ops_ha_add_server`, `ops_ha_list_backends`, `ops_ha_add_backend`, `ops_ha_list_frontends`, `ops_ha_statistics`, `ops_ha_config_test`

**서비스**: `ops_list_services`, `ops_service_action`, `ops_apply_changes`

</details>

## 라이선스

MIT License - 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

---

> **"내가 그에게 하나님의 영을 충만하게 하여 지혜와 총명과 지식과 여러 가지 재주로 정교한 일을 연구하여 금과 은과 놋으로 만들게 하며"**
>
> *- 출애굽기 31:3-4 (개역개정)*
