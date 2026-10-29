package opnsense

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agurrrrr/nagar-mcp/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// addRuleCapture는 firewall/filter/addRule 요청 본문의 rule 객체를 잡는 가짜 OPNsense다.
func addRuleCapture(t *testing.T) (*Client, func() map[string]any) {
	t.Helper()
	var rule map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/firewall/filter/addRule" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var body struct {
			Rule map[string]any `json:"rule"`
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("decode body %s: %v", raw, err)
		}
		rule = body.Rule
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":"saved","uuid":"rule-uuid-1"}`))
	}))
	t.Cleanup(srv.Close)

	client, err := NewClient(config.OPNsenseConfig{URL: srv.URL, APIKey: "k", APISecret: "s"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client, func() map[string]any { return rule }
}

func callAddRule(t *testing.T, client *Client, in AddFirewallRuleInput) *mcp.CallToolResultFor[AddFirewallRuleOutput] {
	t.Helper()
	res, err := createAddFirewallRuleHandler(client)(context.Background(), nil,
		&mcp.CallToolParamsFor[AddFirewallRuleInput]{Name: "ops_add_firewall_rule", Arguments: in})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool error: %+v", res.Content)
	}
	return res
}

// Test_AddFirewallRule_quick_sequence_전달 verifies that quick and sequence reach the
// addRule rule object as given (이슈 tereo #670).
func Test_AddFirewallRule_quick_sequence_전달(t *testing.T) {
	client, captured := addRuleCapture(t)

	res := callAddRule(t, client, AddFirewallRuleInput{
		Action:      "block",
		Interface:   "wan,opt1",
		Source:      "203.0.113.7",
		Destination: "any",
		Protocol:    "any",
		Description: "tereo-ttl=1 inc=act-1",
		Quick:       "1",
		Sequence:    "1",
	})

	rule := captured()
	tests := []struct {
		key  string
		want string
	}{
		{"quick", "1"},
		{"sequence", "1"},
		{"interface", "wan,opt1"},
		{"action", "block"},
		{"source_net", "203.0.113.7"},
		{"destination_net", "any"},
		{"protocol", "any"},
		{"description", "tereo-ttl=1 inc=act-1"},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got, ok := rule[tt.key]; !ok || got != tt.want {
				t.Errorf("rule[%q] = %v (present=%v), want %q", tt.key, got, ok, tt.want)
			}
		})
	}
	if res.StructuredContent.UUID != "rule-uuid-1" || res.StructuredContent.Result != "saved" {
		t.Errorf("output = %+v, want saved/rule-uuid-1", res.StructuredContent)
	}
}

// Test_AddFirewallRule_빈quick_sequence는본문에서빠짐 verifies that empty quick and
// sequence are not sent, so OPNsense keeps its defaults (quick=1, sequence=last+100).
// 기존 필드는 빈 값이어도 그대로 실린다(기존 호출자와 같은 본문).
func Test_AddFirewallRule_빈quick_sequence는본문에서빠짐(t *testing.T) {
	client, captured := addRuleCapture(t)

	callAddRule(t, client, AddFirewallRuleInput{
		Action: "block",
		Source: "203.0.113.7",
	})

	rule := captured()
	for _, key := range []string{"quick", "sequence"} {
		if v, ok := rule[key]; ok {
			t.Errorf("rule[%q] = %v, want key absent", key, v)
		}
	}
	defaults := map[string]string{"enabled": "1", "direction": "in", "ipprotocol": "inet"}
	for key, want := range defaults {
		if rule[key] != want {
			t.Errorf("rule[%q] = %v, want %q", key, rule[key], want)
		}
	}
	for _, key := range []string{"interface", "protocol", "destination_net", "destination_port", "description"} {
		if v, ok := rule[key]; !ok || v != "" {
			t.Errorf("rule[%q] = %v (present=%v), want empty string kept", key, v, ok)
		}
	}
}

// Test_addFirewallRuleBody_quick0유지 verifies that quick="0" is a value, not empty.
func Test_addFirewallRuleBody_quick0유지(t *testing.T) {
	body := addFirewallRuleBody(AddFirewallRuleInput{Action: "pass", Quick: "0"})
	rule, ok := body["rule"].(map[string]any)
	if !ok {
		t.Fatalf("body[rule] type = %T", body["rule"])
	}
	if rule["quick"] != "0" {
		t.Errorf("rule[quick] = %v, want \"0\"", rule["quick"])
	}
	if _, ok := rule["sequence"]; ok {
		t.Errorf("rule[sequence] present, want absent")
	}
}
