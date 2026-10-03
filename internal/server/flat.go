package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// The SDK keeps its wire structs and helpers in an internal package, so this
// module cannot import jsonrpc2 (Go refuses internal packages from another
// module). The public jsonrpc aliases only expose type names, and jsonrpc.ID has
// no exported field, so decoding a request through jsonrpc.Request loses the
// request id. The wire shape is therefore restated here.
type wireRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type wireError struct {
	Code    int64           `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type wireResponse struct {
	JSONRPC string     `json:"jsonrpc"`
	ID      any        `json:"id,omitempty"`
	Result  any        `json:"result,omitempty"`
	Error   *wireError `json:"error,omitempty"`
}

// JSON-RPC error codes. These match the SDK's internal jsonrpc2 values.
const (
	errParse          int64 = -32700
	errInvalidRequest int64 = -32600
	errMethodNotFound int64 = -32601
	errInvalidParams  int64 = -32602
	errInternal       int64 = -32603
)

// Protocol versions. The SDK keeps latestProtocolVersion unexported.
const latestProtocolVersion = "2025-06-18"

var supportedProtocolVersions = []string{
	latestProtocolVersion,
	"2025-03-26",
	"2024-11-05",
}

// serverCapabilities is unexported in the SDK, so the initialize response is
// assembled from these constants for a tools-only server.
type flatCapabilities struct {
	Tools       map[string]any `json:"tools,omitempty"`
	Resources   map[string]any `json:"resources,omitempty"`
	Prompts     map[string]any `json:"prompts,omitempty"`
	Logging     map[string]any `json:"logging,omitempty"`
	Completions map[string]any `json:"completions,omitempty"`
}

type flatInitResult struct {
	ProtocolVersion string              `json:"protocolVersion"`
	Capabilities    flatCapabilities    `json:"capabilities"`
	Instructions    string              `json:"instructions,omitempty"`
	ServerInfo      *mcp.Implementation `json:"serverInfo"`
}

// FlatHandler answers one JSON-RPC request with one JSON response. tereo sends a
// plain POST without an Accept header. The SDK streamable handler requires both
// application/json and text/event-stream and rejects that request with 400, so
// this path exists. No session header is required here.
type FlatHandler struct {
	server      *mcp.Server
	impl        *mcp.Implementation
	mu          sync.Mutex
	initialized bool
}

// NewFlatHandler builds the flat path handler. initialized is tracked by this
// handler because each POST is one request, not a session.
func NewFlatHandler(s *mcp.Server, impl *mcp.Implementation) *FlatHandler {
	return &FlatHandler{server: s, impl: impl}
}

// Initialized reports whether an initialize request has been seen.
func (h *FlatHandler) Initialized() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.initialized
}

// Reset clears the initialize state. Tests use this to test refusal before
// initialize without restarting the server.
func (h *FlatHandler) Reset() {
	h.mu.Lock()
	h.initialized = false
	h.mu.Unlock()
}

func (h *FlatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, nil, errInvalidRequest, "flat path accepts POST only")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, nil, errParse, "failed to read request body")
		return
	}

	var req wireRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, nil, errParse, "invalid JSON: "+err.Error())
		return
	}
	if req.JSONRPC != "2.0" {
		writeJSONError(w, http.StatusBadRequest, nil, errInvalidRequest, fmt.Sprintf("invalid jsonrpc version %q", req.JSONRPC))
		return
	}
	if req.Method == "" {
		writeJSONError(w, http.StatusBadRequest, nil, errInvalidRequest, "request has no method")
		return
	}

	result, rpcErr := h.dispatch(r.Context(), &req)

	// A notification carries no id, so it gets no response body.
	if req.ID == nil {
		if rpcErr != nil {
			slog.Warn("flat notification error", "method", req.Method, "error", rpcErr.Message)
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := wireResponse{JSONRPC: "2.0", ID: req.ID}
	if rpcErr != nil {
		resp.Error = rpcErr
	} else {
		resp.Result = result
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("flat encode failed", "error", err)
	}
}

// dispatch runs one method. initialize and ping are allowed before initialize.
// Every other method is refused before initialize, as the spec requires.
func (h *FlatHandler) dispatch(ctx context.Context, req *wireRequest) (any, *wireError) {
	if req.Method == "initialize" {
		return h.handleInitialize(ctx, req)
	}
	if req.Method == "ping" {
		return map[string]any{}, nil
	}

	if !h.Initialized() {
		return nil, &wireError{
			Code:    errInvalidRequest,
			Message: fmt.Sprintf("method %q is invalid during session initialization", req.Method),
		}
	}

	switch req.Method {
	case "tools/list", "tools/call":
		return h.roundTrip(ctx, req)
	default:
		return nil, &wireError{Code: errMethodNotFound, Message: fmt.Sprintf("method %q not found", req.Method)}
	}
}

func (h *FlatHandler) handleInitialize(ctx context.Context, req *wireRequest) (any, *wireError) {
	version := latestProtocolVersion
	if len(req.Params) > 0 {
		var params struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, &wireError{Code: errInvalidParams, Message: "invalid initialize params: " + err.Error()}
		}
		for _, v := range supportedProtocolVersions {
			if params.ProtocolVersion == v {
				version = v
				break
			}
		}
	}

	caps := flatCapabilities{
		Logging:     map[string]any{},
		Completions: map[string]any{},
	}

	if h.hasTools() {
		caps.Tools = map[string]any{"listChanged": true}
	}

	h.mu.Lock()
	h.initialized = true
	h.mu.Unlock()

	return &flatInitResult{
		ProtocolVersion: version,
		Capabilities:    caps,
		Instructions:    h.instructions(),
		ServerInfo:      h.impl,
	}, nil
}

func (h *FlatHandler) instructions() string { return "" }

// hasTools reports the tool capability through an in-memory round trip. The SDK
// tool set is unexported, so the count cannot be read directly.
func (h *FlatHandler) hasTools() bool {
	names, err := registeredToolNames(h.server)
	return err == nil && len(names) > 0
}

// roundTrip runs tools/list or tools/call through an SDK session so registered
// tool handlers execute normally. Client.Connect performs the SDK's own
// initialize inside this call; the initialize state that the spec requires is
// enforced by this handler (see dispatch).
func (h *FlatHandler) roundTrip(ctx context.Context, req *wireRequest) (any, *wireError) {
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ss, err := h.server.Connect(ctx, serverTransport)
	if err != nil {
		return nil, &wireError{Code: errInternal, Message: "session connect failed: " + err.Error()}
	}
	defer ss.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "flat-client", Version: "0.1.0"}, nil)
	cs, err := client.Connect(ctx, clientTransport)
	if err != nil {
		return nil, &wireError{Code: errInternal, Message: "client connect failed: " + err.Error()}
	}
	defer cs.Close()

	switch req.Method {
	case "tools/list":
		var params struct {
			Cursor string `json:"cursor,omitempty"`
		}
		if len(req.Params) > 0 {
			if err := json.Unmarshal(req.Params, &params); err != nil {
				return nil, &wireError{Code: errInvalidParams, Message: "invalid tools/list params: " + err.Error()}
			}
		}
		result, err := cs.ListTools(ctx, &mcp.ListToolsParams{Cursor: params.Cursor})
		if err != nil {
			return nil, &wireError{Code: errInternal, Message: err.Error()}
		}
		return result, nil

	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments,omitempty"`
		}
		if len(req.Params) > 0 {
			if err := json.Unmarshal(req.Params, &params); err != nil {
				return nil, &wireError{Code: errInvalidParams, Message: "invalid tools/call params: " + err.Error()}
			}
		}
		if !IsAllowed(params.Name) {
			return nil, &wireError{Code: errInvalidParams, Message: fmt.Sprintf("unknown tool %q", params.Name)}
		}
		result, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      params.Name,
			Arguments: jsonToAny(params.Arguments),
		})
		if err != nil {
			return nil, &wireError{Code: errInternal, Message: err.Error()}
		}
		return result, nil

	default:
		return nil, &wireError{Code: errMethodNotFound, Message: fmt.Sprintf("method %q not found", req.Method)}
	}
}

func jsonToAny(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil
	}
	return v
}

func writeJSONError(w http.ResponseWriter, status int, id any, code int64, message string) {
	resp := wireResponse{JSONRPC: "2.0", ID: id, Error: &wireError{Code: code, Message: message}}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("flat error encode failed", "error", err)
	}
}

// AcceptsEventStream reports whether an Accept header asks for the streamable
// transport. tereo sends no Accept header, so it takes the flat path.
func AcceptsEventStream(accept string) bool {
	for _, part := range strings.Split(accept, ",") {
		if strings.HasPrefix(strings.TrimSpace(part), "text/event-stream") {
			return true
		}
	}
	return false
}
