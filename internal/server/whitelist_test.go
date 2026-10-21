package server

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Test_AllowedTools_set_tereoenum23중미등록항목제외항확인 verifies the whitelist size and
// the missing tereo enum entry (pve_vm_snapshot_delete).
func Test_AllowedTools_set_tereoenum23중미등록항목제외항확인(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{"허용셋 크기", 22},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(AllowedTools) != tt.want {
				t.Errorf("AllowedTools = %d, want %d", len(AllowedTools), tt.want)
			}
			if IsAllowed("pve_vm_snapshot_delete") {
				t.Errorf("pve_vm_snapshot_delete must not be allowed (nagar-mcp 미등록)")
			}
		})
	}

	for _, name := range AllowedTools {
		if !IsAllowed(name) {
			t.Errorf("IsAllowed(%q) = false, want true", name)
		}
	}

	// ops_list_firewall_rules(읽기 전용)는 허용셋에 있어야 한다 (이슈 #529).
	if !IsAllowed("ops_list_firewall_rules") {
		t.Errorf("IsAllowed(%q) = false, want true (고아 차단 탐지, #529)", "ops_list_firewall_rules")
	}

	// alias 도구 6개는 허용셋에 있어야 한다 (2026-10-19 확장).
	for _, name := range []string{
		"ops_list_firewall_aliases", "ops_get_firewall_alias", "ops_add_firewall_alias",
		"ops_update_firewall_alias", "ops_delete_firewall_alias", "ops_toggle_firewall_alias",
	} {
		if !IsAllowed(name) {
			t.Errorf("IsAllowed(%q) = false, want true (alias 도구 확장)", name)
		}
	}

	// 미허용 도구 확인: nagar-mcp에 등록돼 있어도 tereo는 쓸 수 없다.
	for _, name := range []string{"pve_list_vms", "ops_get_firewall_rule", "k8s_pod_logs"} {
		if IsAllowed(name) {
			t.Errorf("IsAllowed(%q) = true, want false", name)
		}
	}
}

// Test_ApplyWhitelist_미허용도구제거 verifies that non-whitelisted tools are removed.
func Test_ApplyWhitelist_미허용도구제거(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)

	allowed := "ops_firmware_status"
	removed := "ops_get_firewall_rule" // nagar-mcp에 등록되지만 tereo whitelist 밖

	addTestTool(s, allowed)
	addTestTool(s, removed)

	got, err := ApplyWhitelist(s)
	if err != nil {
		t.Fatalf("ApplyWhitelist() error = %v", err)
	}

	if len(got) != 1 || got[0] != removed {
		t.Errorf("ApplyWhitelist() removed = %v, want [%s]", got, removed)
	}

	names, err := registeredToolNames(s)
	if err != nil {
		t.Fatalf("registeredToolNames() error = %v", err)
	}
	if len(names) != 1 || names[0] != allowed {
		t.Errorf("remaining tools = %v, want [%s]", names, allowed)
	}
}

// Test_AddTool_whitelist경계_허용셋내도구는남는다 verifies a whitelisted tool survives.
func Test_AddTool_whitelist경계_허용셋내도구는남는다(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)
	addTestTool(s, "pve_vm_snapshot_create")

	removed, err := ApplyWhitelist(s)
	if err != nil {
		t.Fatalf("ApplyWhitelist() error = %v", err)
	}
	if len(removed) != 0 {
		t.Errorf("ApplyWhitelist() removed = %v, want none", removed)
	}

	names, err := registeredToolNames(s)
	if err != nil {
		t.Fatalf("registeredToolNames() error = %v", err)
	}
	if len(names) != 1 || names[0] != "pve_vm_snapshot_create" {
		t.Errorf("remaining tools = %v, want [pve_vm_snapshot_create]", names)
	}
}

// Test_AddTool_빈도구셋_제거대상없음 is the empty-input edge case.
func Test_AddTool_빈도구셋_제거대상없음(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)

	removed, err := ApplyWhitelist(s)
	if err != nil {
		t.Fatalf("ApplyWhitelist() error = %v", err)
	}
	if len(removed) != 0 {
		t.Errorf("ApplyWhitelist() removed = %v, want none", removed)
	}

	names, err := registeredToolNames(s)
	if err != nil {
		t.Fatalf("registeredToolNames() error = %v", err)
	}
	if len(names) != 0 {
		t.Errorf("remaining tools = %v, want none", names)
	}
}

// addTestTool registers a tool that returns empty content. A tool needs an input
// schema, so map[string]any is used.
func addTestTool(s *mcp.Server, name string) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        name,
		Description: "test tool",
	}, func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "{}"}}}, nil
	})
}
