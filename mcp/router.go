package mcp

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/j4ng5y/mcpgate/server"
)

// Router handles request routing to appropriate upstream servers
type Router struct {
	manager *server.Manager
}

// NewRouter creates a new request router
func NewRouter(mgr *server.Manager) *Router {
	return &Router{
		manager: mgr,
	}
}

// Route handles a JSON-RPC request and returns a response
func (r *Router) Route(ctx context.Context, req *Request) *Response {
	// Validate request
	if req.JSONRPC != "2.0" {
		return &Response{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    InvalidRequest,
				Message: "Invalid JSON-RPC version",
			},
		}
	}

	// Handle gateway-level methods
	switch req.Method {
	case "gateway/list_servers":
		return r.handleListServers(ctx, req)
	case "gateway/get_server":
		return r.handleGetServer(ctx, req)
	case "gateway/server_status":
		return r.handleServerStatus(ctx, req)
	case "gateway/capabilities":
		return r.handleCapabilities(ctx, req)
	}

	// Route to upstream server based on method or explicit server specification
	return r.routeToServer(ctx, req)
}

// handleListServers returns a list of all registered servers
func (r *Router) handleListServers(ctx context.Context, req *Request) *Response {
	servers := r.manager.ListServers()
	result := make([]map[string]interface{}, 0, len(servers))

	for _, srv := range servers {
		result = append(result, map[string]interface{}{
			"name":         srv.Name,
			"connected":    srv.IsConnected(),
			"initialized":  srv.IsInitialized(),
			"transport":    srv.Config.Transport,
			"capabilities": srv.Capabilities,
		})
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleGetServer returns details about a specific server
func (r *Router) handleGetServer(ctx context.Context, req *Request) *Response {
	var params struct {
		Name string `json:"name"`
	}

	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &JSONRPCError{
					Code:    InvalidParams,
					Message: "Invalid parameters",
				},
			}
		}
	}

	srv, err := r.manager.GetServer(params.Name)
	if err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32000,
				Message: "Server not found",
			},
		}
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"name":         srv.Name,
			"connected":    srv.IsConnected(),
			"initialized":  srv.IsInitialized(),
			"transport":    srv.Config.Transport,
			"capabilities": srv.Capabilities,
			"metadata":     srv.Metadata,
		},
	}
}

// handleServerStatus returns current status of a server
func (r *Router) handleServerStatus(ctx context.Context, req *Request) *Response {
	var params struct {
		Name string `json:"name"`
	}

	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &JSONRPCError{
					Code:    InvalidParams,
					Message: "Invalid parameters",
				},
			}
		}
	}

	srv, err := r.manager.GetServer(params.Name)
	if err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32000,
				Message: "Server not found",
			},
		}
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"connected":   srv.IsConnected(),
			"initialized": srv.IsInitialized(),
			"last_used":   srv.GetLastUsed(),
		},
	}
}

// handleCapabilities returns capabilities of a server or all servers
func (r *Router) handleCapabilities(ctx context.Context, req *Request) *Response {
	var params struct {
		Name string `json:"name,omitempty"`
	}

	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &JSONRPCError{
					Code:    InvalidParams,
					Message: "Invalid parameters",
				},
			}
		}
	}

	if params.Name != "" {
		srv, err := r.manager.GetServer(params.Name)
		if err != nil {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &JSONRPCError{
					Code:    -32000,
					Message: "Server not found",
				},
			}
		}

		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"name":         srv.Name,
				"capabilities": srv.Capabilities,
			},
		}
	}

	// Return capabilities from all servers
	result := make(map[string][]string)
	for _, srv := range r.manager.ListServers() {
		result[srv.Name] = srv.Capabilities
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// routeToServer routes a request to the appropriate upstream server
func (r *Router) routeToServer(ctx context.Context, req *Request) *Response {
	// Try to determine target server
	// First check for explicit server specification in params
	targetServer := r.findTargetServer(ctx, req)
	if targetServer == nil {
		// If no target, try routing based on method
		// For now, try all servers with the capability
		servers := r.manager.ListServers()
		if len(servers) == 0 {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &JSONRPCError{
					Code:    -32000,
					Message: "No servers available",
				},
			}
		}
		// Use first available server
		targetServer = servers[0]
	}

	// Send request to target server
	log.Printf("Routing request %v to server %s", req.ID, targetServer.Name)

	// Convert request to map for sending
	reqMap := map[string]interface{}{
		"jsonrpc": req.JSONRPC,
		"method":  req.Method,
	}
	if req.ID != nil {
		reqMap["id"] = req.ID
	}
	if len(req.Params) > 0 {
		var params interface{}
		if err := json.Unmarshal(req.Params, &params); err == nil {
			reqMap["params"] = params
		}
	}

	respData, err := targetServer.SendRequest(ctx, reqMap)
	if err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    InternalError,
				Message: err.Error(),
			},
		}
	}

	// Parse the response
	var response Response
	if err := json.Unmarshal(respData, &response); err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    ParseError,
				Message: "Failed to parse upstream response",
			},
		}
	}

	return &response
}

// findTargetServer determines which server should handle the request
func (r *Router) findTargetServer(ctx context.Context, req *Request) *server.ManagedServer {
	// Check for explicit server in params
	if req.Params != nil {
		var params map[string]interface{}
		if err := json.Unmarshal(req.Params, &params); err == nil {
			if serverName, ok := params["_server"].(string); ok {
				srv, err := r.manager.GetServer(serverName)
				if err == nil {
					return srv
				}
			}
		}
	}

	// Try to route based on method name
	// e.g., "tools/list" -> find server with tools capability
	capability := r.extractCapability(req.Method)
	if capability != "" {
		servers := r.manager.ListServersByCapability(capability)
		if len(servers) > 0 {
			return servers[0]
		}
	}

	return nil
}

// extractCapability extracts capability from method name
func (r *Router) extractCapability(method string) string {
	// Map methods to capabilities
	switch {
	case strings.HasPrefix(method, "tools/"):
		return "tools"
	case strings.HasPrefix(method, "resources/"):
		return "resources"
	case strings.HasPrefix(method, "prompts/"):
		return "prompts"
	}
	return ""
}
