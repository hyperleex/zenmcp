package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hyperleex/zenmcp/protocol"
	"github.com/hyperleex/zenmcp/registry"
)

type Context struct {
	context.Context
	requestID     *protocol.RequestID
	progressToken *protocol.ProgressToken
	mu            sync.RWMutex
	cancelled     bool
	progress      float64
	total         *float64
}

func NewContext(ctx context.Context, requestID *protocol.RequestID) *Context {
	return &Context{
		Context:   ctx,
		requestID: requestID,
	}
}

func (c *Context) WithProgressToken(token *protocol.ProgressToken) *Context {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.progressToken = token
	return c
}

func (c *Context) RequestID() *protocol.RequestID {
	return c.requestID
}

func (c *Context) ProgressToken() *protocol.ProgressToken {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.progressToken
}

func (c *Context) SetProgress(progress float64, total *float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.progress = progress
	c.total = total
}

func (c *Context) Progress() (float64, *float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.progress, c.total
}

func (c *Context) Cancel() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancelled = true
}

func (c *Context) IsCancelled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cancelled
}

// Type-safe handler interfaces

// ToolHandler is the modern type-safe interface for tool handlers
type ToolHandler[T any] interface {
	Call(ctx *Context, args T) (*protocol.ToolCallResult, error)
}

// ToolFunc is a function type that implements ToolHandler
type ToolFunc[T any] func(ctx *Context, args T) (*protocol.ToolCallResult, error)

func (f ToolFunc[T]) Call(ctx *Context, args T) (*protocol.ToolCallResult, error) {
	return f(ctx, args)
}

// ResourceHandler is the modern type-safe interface for resource handlers
type ResourceHandler interface {
	Read(ctx *Context, uri string) ([]byte, string, error)
}

// ResourceFunc is a function type that implements ResourceHandler
type ResourceFunc func(ctx *Context, uri string) ([]byte, string, error)

func (f ResourceFunc) Read(ctx *Context, uri string) ([]byte, string, error) {
	return f(ctx, uri)
}

// PromptHandler is the modern type-safe interface for prompt handlers
type PromptHandler[T any] interface {
	Get(ctx *Context, args T) (*registry.PromptResult, error)
}

// PromptFunc is a function type that implements PromptHandler
type PromptFunc[T any] func(ctx *Context, args T) (*registry.PromptResult, error)

func (f PromptFunc[T]) Get(ctx *Context, args T) (*registry.PromptResult, error) {
	return f(ctx, args)
}

// Wrapper types to bridge typed handlers to legacy interface
type typedToolWrapper[T any] struct {
	handler ToolHandler[T]
}

func (w *typedToolWrapper[T]) Call(ctx interface{}, args json.RawMessage) (*protocol.ToolCallResult, error) {
	runtimeCtx, ok := ctx.(*Context)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}
	
	var typedArgs T
	if err := json.Unmarshal(args, &typedArgs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}
	
	return w.handler.Call(runtimeCtx, typedArgs)
}

type typedResourceWrapper struct {
	handler ResourceHandler
}

func (w *typedResourceWrapper) Read(ctx interface{}, uri string) ([]byte, string, error) {
	runtimeCtx, ok := ctx.(*Context)
	if !ok {
		return nil, "", fmt.Errorf("invalid context type")
	}
	
	return w.handler.Read(runtimeCtx, uri)
}

// RegisterToolTyped registers a type-safe tool handler
func RegisterToolTyped[T any](reg *registry.Registry, name, description string, handler ToolHandler[T]) error {
	var zero T
	
	// Wrap the typed handler to match legacy interface
	legacyHandler := &typedToolWrapper[T]{handler: handler}
	
	return reg.RegisterTool(name, description, legacyHandler, zero)
}

// RegisterResourceTyped registers a type-safe resource handler  
func RegisterResourceTyped(reg *registry.Registry, uri, name, description, mimeType string, handler ResourceHandler) {
	// Wrap to match legacy interface
	legacyHandler := &typedResourceWrapper{handler: handler}
	
	reg.RegisterResource(uri, name, description, mimeType, legacyHandler)
}

