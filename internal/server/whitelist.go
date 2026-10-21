package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AllowedTools는 tereo의 NagarTool enum과 동일한 허용셋이다.
//
// tereo enum은 23개다(crates/tereo-adapters/src/nagar.rs). nagar-mcp에 등록된
// 도구와 대조하면 22개만 존재한다. pve_vm_snapshot_delete는 nagar-mcp에 등록되어
// 있지 않다(pve_vm_snapshot_create, pve_vm_snapshot_list만 있다). 따라서 허용셋은
// 22개다. tereo enum의 정합성 문제는 보고한다 — enum은 축소하지 않는다. 호출 시
// unknown tool(-32602)로 응답한다.
//
// 2026-10-19: firewall alias 도구 6개를 더했다. DESIGN.md §11의 2단계 전환
// (fw_block_src를 alias `tereo_block` 기반 실차단으로 옮기는 경로)과 TTL 갱신이
// alias를 필요로 한다.
//
// 2026-10-21 (이슈 #529): ops_list_firewall_rules(읽기 전용)를 더했다. tereo의
// TTL sweep이 고아 차단 탐지를 한다. OPNsense에 남은 tereo 규칙(description
// `tereo-ttl=<unix> inc=<action_id>`)을 control.db의 ttl_timers 대장과 대조한다.
// 쓰기 도구는 더하지 않았다 — 탐지에 필요한 것은 list뿐이다.
var AllowedTools = []string{
	// OPNsense (ops_*)
	"ops_add_firewall_rule",
	"ops_delete_firewall_rule",
	"ops_apply_changes",
	// 읽기 전용 (이슈 #529): 고아 차단 탐지가 규칙 목록과 TTL 대장을 대조한다.
	"ops_list_firewall_rules",
	"ops_add_dns_override",
	"ops_delete_dns_override",
	"ops_firmware_status",
	// OPNsense firewall alias — tereo의 alias 차단 경로·TTL 갱신 대상 (2026-10-19)
	"ops_list_firewall_aliases",
	"ops_get_firewall_alias",
	"ops_add_firewall_alias",
	"ops_update_firewall_alias",
	"ops_delete_firewall_alias",
	"ops_toggle_firewall_alias",
	// K8s (k8s_*)
	"k8s_scale_deployment",
	"k8s_deployment_status",
	"k8s_delete_pod",
	"k8s_cordon_node",
	"k8s_uncordon_node",
	// Proxmox (pve_*)
	"pve_vm_stop",
	"pve_vm_start",
	"pve_vm_reboot",
	"pve_vm_snapshot_create",
	// pve_vm_snapshot_delete: nagar-mcp 미등록. tereo enum 정합성 문제(보고).
}

// IsAllowed reports whether name is in the tereo whitelist.
func IsAllowed(name string) bool {
	for _, n := range AllowedTools {
		if n == name {
			return true
		}
	}
	return false
}

// registeredToolNames lists the tools currently registered on s. The SDK does not
// export an enumeration, so the list is read through an in-memory session round trip.
func registeredToolNames(s *mcp.Server) ([]string, error) {
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	if _, err := s.Connect(context.Background(), serverTransport); err != nil {
		return nil, fmt.Errorf("whitelist: connect: %w", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "whitelist-reader", Version: "0.1.0"}, nil)
	session, err := client.Connect(context.Background(), clientTransport)
	if err != nil {
		return nil, fmt.Errorf("whitelist: client connect: %w", err)
	}
	defer session.Close()

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("whitelist: list tools: %w", err)
	}

	names := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
	}
	return names, nil
}

// ApplyWhitelist removes every registered tool that is not in AllowedTools.
// It returns the names it removed. The whitelist is fail-closed: a tool outside
// the set is not reachable through this server instance.
func ApplyWhitelist(s *mcp.Server) ([]string, error) {
	registered, err := registeredToolNames(s)
	if err != nil {
		return nil, err
	}

	removed := make([]string, 0, len(registered))
	for _, name := range registered {
		if !IsAllowed(name) {
			removed = append(removed, name)
		}
	}

	if len(removed) > 0 {
		s.RemoveTools(removed...)
		slog.Info("whitelist applied", "allowed", len(AllowedTools), "removed", len(removed))
	}

	return removed, nil
}
