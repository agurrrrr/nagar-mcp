package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// newTestFlatHandler builds a deterministic handler with one whitelisted tool.
// Every test calls Reset() so the initialize state does not leak between cases.
func newTestFlatHandler() *FlatHandler {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)
	addTestTool(s, "ops_firmware_status")
	return NewFlatHandler(s, &mcp.Implementation{Name: "test", Version: "0.1.0"})
}

// postFlat sends one JSON-RPC request through the flat path and returns the
// decoded response. It mirrors what tereo sends: one POST, no Accept header.
func postFlat(h *FlatHandler, accept, body string) (*httptest.ResponseRecorder, *wireResponse) {
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var resp wireResponse
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			return rec, nil
		}
	}
	return rec, &resp
}

// Test_ServeHTTP_flat왕복_initialize후toolscall200 verifies the two-step round trip:
// initialize, then tools/call on a whitelisted tool. Both answer 200 with no error.
func Test_ServeHTTP_flat왕복_initialize후toolscall200(t *testing.T) {
	h := newTestFlatHandler()
	h.Reset()

	tests := []struct {
		name string
		body string
	}{
		{
			"initialize 왕복",
			`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{}}}`,
		},
		{
			"tools/call 왕복 (허용 도구)",
			`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"ops_firmware_status","arguments":{}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, resp := postFlat(h, "", tt.body)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			if resp == nil {
				t.Fatalf("response body is not JSON-RPC")
			}
			if resp.Error != nil {
				t.Fatalf("error = %v, want none", resp.Error)
			}
			if resp.Result == nil {
				t.Fatalf("result is empty, want a result")
			}
		})
	}

	if !h.Initialized() {
		t.Errorf("Initialized() = false after initialize, want true")
	}
}

// Test_ServeHTTP_initialize전거부_toolscall와toolslist_32600 verifies that methods
// other than initialize and ping are refused before initialize.
func Test_ServeHTTP_initialize전거부_toolscall와toolslist_32600(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{"tools/call 거부", "tools/call"},
		{"tools/list 거부", "tools/list"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestFlatHandler()
			h.Reset()

			body := `{"jsonrpc":"2.0","id":1,"method":"` + tt.method + `","params":{}}`
			rec, resp := postFlat(h, "", body)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if resp == nil || resp.Error == nil {
				t.Fatalf("error is missing, want -32600")
			}
			if resp.Error.Code != errInvalidRequest {
				t.Errorf("error code = %d, want %d", resp.Error.Code, errInvalidRequest)
			}
			if h.Initialized() {
				t.Errorf("Initialized() = true after a refused method, want false")
			}
		})
	}
}

// Test_ServeHTTP_ping_initialize전허용 verifies that ping is allowed before initialize.
func Test_ServeHTTP_ping_initialize전허용(t *testing.T) {
	h := newTestFlatHandler()
	h.Reset()

	rec, resp := postFlat(h, "", `{"jsonrpc":"2.0","id":1,"method":"ping","params":{}}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if resp == nil || resp.Error != nil {
		t.Fatalf("error = %v, want none for ping", resp.Error)
	}
	if h.Initialized() {
		t.Errorf("Initialized() = true after ping, want false")
	}
}

// Test_ServeHTTP_whitelist밖도구_32602unknowntool verifies that a tool outside the
// whitelist answers with -32602. pve_vm_snapshot_delete is the tereo enum entry
// that nagar-mcp does not register.
func Test_ServeHTTP_whitelist밖도구_32602unknowntool(t *testing.T) {
	h := newTestFlatHandler()
	h.Reset()

	// initialize first, so dispatch reaches the whitelist check.
	_, resp := postFlat(h, "", `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{}}}`)
	if resp != nil && resp.Error != nil {
		t.Fatalf("initialize failed: %v", resp.Error)
	}

	rec, resp := postFlat(h, "", `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"pve_vm_snapshot_delete","arguments":{}}}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if resp == nil || resp.Error == nil {
		t.Fatalf("error is missing, want -32602 unknown tool")
	}
	if resp.Error.Code != errInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, errInvalidParams)
	}
	if !strings.Contains(resp.Error.Message, "unknown tool") {
		t.Errorf("error message = %q, want it to mention unknown tool", resp.Error.Message)
	}

	// A registered but non-whitelisted tool must answer the same way.
	rec, resp = postFlat(h, "", `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"pve_list_vms","arguments":{}}}`)
	if resp == nil || resp.Error == nil || resp.Error.Code != errInvalidParams {
		t.Fatalf("error = %v, want -32602 for a non-whitelisted tool", resp.Error)
	}
}

// Test_ServeHTTP_notification_ID없으면_본문없음204 verifies that a request without an
// id gets no response body.
func Test_ServeHTTP_notification_ID없으면_본문없음204(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"notification ping", `{"jsonrpc":"2.0","method":"ping"}`},
		{"notification initialized", `{"jsonrpc":"2.0","method":"notifications/initialized"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestFlatHandler()
			h.Reset()

			rec, _ := postFlat(h, "", tt.body)

			if rec.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want 204", rec.Code)
			}
			if rec.Body.Len() != 0 {
				t.Errorf("body = %q, want empty for a notification", rec.Body.String())
			}
		})
	}
}

// Test_ServeHTTP_POST아니면_405 verifies that the flat path accepts POST only.
func Test_ServeHTTP_POST아니면_405(t *testing.T) {
	h := newTestFlatHandler()
	h.Reset()

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}

	var resp wireResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response body is not JSON-RPC: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != errInvalidRequest {
		t.Errorf("error = %v, want -32600", resp.Error)
	}
}

// Test_BearerMiddleware_실패401_WWWAuthenticateBearer verifies the auth path. Token
// empty means no authentication (the loopback-only case). ConstantTimeCompare is
// not observable here, so the test checks the outcome only.
func Test_BearerMiddleware_실패401_WWWAuthenticateBearer(t *testing.T) {
	h := newTestFlatHandler()

	tests := []struct {
		name       string
		token      string
		authHeader string
		wantStatus int
	}{
		{"토큰 설정 + 헤더 없음 401", "test-token", "", http.StatusUnauthorized},
		{"토큰 설정 + 헤더 불일치 401", "test-token", "Bearer wrong", http.StatusUnauthorized},
		{"토큰 설정 + 헤더 일치 통과", "test-token", "Bearer test-token", http.StatusOK},
		{"토큰 빈 값 통과 (loopback 봉인)", "", "", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := BearerMiddleware(tt.token, h)

			req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(
				`{"jsonrpc":"2.0","id":1,"method":"ping","params":{}}`))
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusUnauthorized {
				if got := rec.Header().Get("WWW-Authenticate"); got != "Bearer" {
					t.Errorf("WWW-Authenticate = %q, want Bearer", got)
				}
				var resp wireResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("error body is not JSON-RPC: %v", err)
				}
				if resp.Error == nil || resp.Error.Code != errInvalidRequest {
					t.Errorf("error = %v, want -32600 unauthorized", resp.Error)
				}
			} else {
				if got := rec.Header().Get("WWW-Authenticate"); got != "" {
					t.Errorf("WWW-Authenticate = %q, want none on success", got)
				}
			}
		})
	}
}

// Test_AcceptsEventStream_accept분기_flat과streamable verifies the mux branch: no
// Accept header takes the flat path, text/event-stream takes the streamable path.
func Test_AcceptsEventStream_accept분기_flat과streamable(t *testing.T) {
	tests := []struct {
		name   string
		accept string
		want   bool
	}{
		{"Accept 없음 flat", "", false},
		{"application/json만 flat", "application/json", false},
		{"text/event-stream streamable", "text/event-stream", true},
		{"둘 다 streamable", "application/json, text/event-stream", true},
		{"공백·파라미터 포함 streamable", " application/json; q=1.0, text/event-stream", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AcceptsEventStream(tt.accept); got != tt.want {
				t.Errorf("AcceptsEventStream(%q) = %v, want %v", tt.accept, got, tt.want)
			}
		})
	}
}
