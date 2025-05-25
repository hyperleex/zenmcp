package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperleex/zenmcp/protocol"
	"github.com/hyperleex/zenmcp/registry"
	"github.com/hyperleex/zenmcp/runtime"
	"github.com/hyperleex/zenmcp/transport"
)

type Server struct {
	transport transport.Transport
	registry  *registry.Registry
	router    *runtime.Router
	options   ServerOptions
}

type ServerOptions struct {
	Logger Logger
}

type Logger interface {
	Printf(format string, v ...interface{})
}

type defaultLogger struct{}

func (d defaultLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func NewServer(transport transport.Transport, opts ...ServerOption) *Server {
	options := ServerOptions{
		Logger: defaultLogger{},
	}
	
	for _, opt := range opts {
		opt(&options)
	}
	
	reg := registry.New()
	router := runtime.NewRouter(reg)
	
	return &Server{
		transport: transport,
		registry:  reg,
		router:    router,
		options:   options,
	}
}

type ServerOption func(*ServerOptions)

func WithLogger(logger Logger) ServerOption {
	return func(opts *ServerOptions) {
		opts.Logger = logger
	}
}

func (s *Server) RegisterTool(name, description string, handler registry.LegacyToolHandler, inputType interface{}) error {
	return s.registry.RegisterTool(name, description, handler, inputType)
}

// RegisterToolTyped registers a type-safe tool handler
func RegisterToolTyped[T any](s *Server, name, description string, handler runtime.ToolHandler[T]) error {
	return runtime.RegisterToolTyped(s.registry, name, description, handler)
}

// RegisterToolFunc is a convenience method for registering function-based tool handlers
func RegisterToolFunc[T any](s *Server, name, description string, handler runtime.ToolFunc[T]) error {
	return runtime.RegisterToolTyped(s.registry, name, description, handler)
}

func (s *Server) RegisterResource(uri, name, description, mimeType string, handler registry.ResourceHandler) {
	s.registry.RegisterResource(uri, name, description, mimeType, handler)
}

func (s *Server) RegisterPrompt(name, description string, args []registry.Argument, handler registry.LegacyPromptHandler) {
	s.registry.RegisterPrompt(name, description, args, handler)
}

// RegisterResourceFunc is a convenience method for registering function-based resource handlers
func RegisterResourceFunc(s *Server, uri, name, description, mimeType string, handler runtime.ResourceFunc) {
	runtime.RegisterResourceTyped(s.registry, uri, name, description, mimeType, handler)
}

func (s *Server) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		conn, err := s.transport.Accept(ctx)
		if err != nil {
			return fmt.Errorf("failed to accept connection: %w", err)
		}
		
		go s.handleConnection(ctx, conn)
	}
}

func (s *Server) handleConnection(ctx context.Context, conn transport.Connection) {
	defer conn.Close()
	
	codec := conn.Codec()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-conn.Context().Done():
			return
		default:
		}
		
		var msg json.RawMessage
		if err := codec.Decode(&msg); err != nil {
			if err.Error() == "EOF" {
				// EOF is expected when client disconnects
				return
			}
			s.options.Logger.Printf("decode error: %v", err)
			return
		}
		
		if err := s.processMessage(ctx, codec, msg); err != nil {
			s.options.Logger.Printf("process message error: %v", err)
		}
	}
}

func (s *Server) processMessage(ctx context.Context, codec protocol.Codec, msg json.RawMessage) error {
	var base struct {
		JSONRPC string             `json:"jsonrpc"`
		ID      *protocol.RequestID `json:"id,omitempty"`
		Method  string             `json:"method,omitempty"`
	}
	
	if err := json.Unmarshal(msg, &base); err != nil {
		return s.sendError(codec, nil, protocol.ParseError, "parse error", err.Error())
	}
	
	if base.ID == nil {
		return nil
	}
	
	runtimeCtx := runtime.NewContext(ctx, base.ID)
	
	var req protocol.Request
	if err := json.Unmarshal(msg, &req); err != nil {
		return s.sendError(codec, base.ID, protocol.InvalidRequest, "invalid request", err.Error())
	}
	
	result, err := s.router.Route(runtimeCtx, req.Method, req.Params)
	if err != nil {
		if mcpErr, ok := err.(*protocol.Error); ok {
			return s.sendError(codec, base.ID, mcpErr.Code, mcpErr.Message, mcpErr.Data)
		}
		return s.sendError(codec, base.ID, protocol.InternalError, "internal error", err.Error())
	}
	
	response := &protocol.Response{
		JSONRPC: protocol.JSONRPCVersion,
		ID:      base.ID,
		Result:  result,
	}
	
	return codec.Encode(response)
}

func (s *Server) sendError(codec protocol.Codec, id *protocol.RequestID, code int, message string, data interface{}) error {
	response := &protocol.Response{
		JSONRPC: protocol.JSONRPCVersion,
		ID:      id,
		Error: &protocol.Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	return codec.Encode(response)
}

func (s *Server) Close() error {
	return s.transport.Close()
}