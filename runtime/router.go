package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/hyperleex/zenmcp/protocol"
	"github.com/hyperleex/zenmcp/registry"
)

type Router struct {
	registry *registry.Registry
	handlers map[string]RequestHandler
}

type RequestHandler func(ctx *Context, params json.RawMessage) (interface{}, error)

func NewRouter(reg *registry.Registry) *Router {
	r := &Router{
		registry: reg,
		handlers: make(map[string]RequestHandler),
	}
	
	r.registerCoreHandlers()
	return r
}

func (r *Router) registerCoreHandlers() {
	r.handlers[protocol.MethodInitialize] = r.handleInitialize
	r.handlers[protocol.MethodToolsList] = r.handleToolsList
	r.handlers[protocol.MethodToolsCall] = r.handleToolsCall
}

func (r *Router) Route(ctx *Context, method string, params json.RawMessage) (interface{}, error) {
	handler, exists := r.handlers[method]
	if !exists {
		return nil, protocol.NewError(protocol.MethodNotFound, "method not found", nil)
	}
	
	return handler(ctx, params)
}

func (r *Router) handleInitialize(ctx *Context, params json.RawMessage) (interface{}, error) {
	var req protocol.InitializeRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, protocol.NewError(protocol.InvalidParams, "invalid parameters", err.Error())
	}
	
	return &protocol.InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: protocol.ServerCapabilities{
			Tools: &protocol.ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: protocol.ServerInfo{
			Name:    "zenmcp-server",
			Version: "0.1.0",
		},
	}, nil
}

func (r *Router) handleToolsList(ctx *Context, params json.RawMessage) (interface{}, error) {
	tools := r.registry.ListTools()
	return &protocol.ToolListResult{
		Tools: tools,
	}, nil
}

func (r *Router) handleToolsCall(ctx *Context, params json.RawMessage) (interface{}, error) {
	var req protocol.ToolCallRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, protocol.NewError(protocol.InvalidParams, "invalid parameters", err.Error())
	}
	
	tool, exists := r.registry.GetTool(req.Name)
	if !exists {
		return nil, protocol.NewError(protocol.MethodNotFound, fmt.Sprintf("tool %s not found", req.Name), nil)
	}
	
	return tool.Handler.Call(ctx, req.Arguments)
}