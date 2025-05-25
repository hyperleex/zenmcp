package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/hyperleex/zenmcp/registry"
	"github.com/hyperleex/zenmcp/runtime"
	"github.com/hyperleex/zenmcp/transport/stdio"
)

func TestNewServer(t *testing.T) {
	transport := stdio.New()
	defer transport.Close()

	server := NewServer(transport)
	if server == nil {
		t.Fatal("expected server to be created")
	}

	if server.transport != transport {
		t.Error("transport not set correctly")
	}
	if server.registry == nil {
		t.Error("registry not initialized")
	}
	if server.router == nil {
		t.Error("router not initialized")
	}
}

func TestServerWithOptions(t *testing.T) {
	transport := stdio.New()
	defer transport.Close()

	logger := &testLogger{}
	server := NewServer(transport, WithLogger(logger))

	if server.options.Logger != logger {
		t.Error("logger option not applied")
	}
}

func TestServerRegisterTool(t *testing.T) {
	transport := stdio.New()
	defer transport.Close()

	server := NewServer(transport)

	handler := &testToolHandler{}
	err := server.RegisterTool("test_tool", "A test tool", handler, map[string]interface{}{})
	if err != nil {
		t.Fatalf("failed to register tool: %v", err)
	}
}

func TestServerRegisterResource(t *testing.T) {
	transport := stdio.New()
	defer transport.Close()

	server := NewServer(transport)

	handler := &testResourceHandler{}
	server.RegisterResource("test://resource", "test_resource", "A test resource", "text/plain", handler)
}

func TestServerRegisterPrompt(t *testing.T) {
	transport := stdio.New()
	defer transport.Close()

	server := NewServer(transport)

	handler := &testPromptHandler{}
	args := []registry.Argument{{Name: "input", Description: "Test input", Required: true}}
	server.RegisterPrompt("test_prompt", "A test prompt", args, handler)
}

func TestServerClose(t *testing.T) {
	transport := stdio.New()
	server := NewServer(transport)

	err := server.Close()
	if err != nil {
		t.Errorf("close failed: %v", err)
	}
}

type testLogger struct {
	messages []string
}

func (l *testLogger) Printf(format string, v ...interface{}) {
	// Store messages for testing
}

type testToolHandler struct{}

func (h *testToolHandler) Call(ctx *runtime.Context, args map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{"result": "success"}, nil
}

type testResourceHandler struct{}

func (h *testResourceHandler) Read(ctx *runtime.Context, uri string) ([]byte, error) {
	return []byte("test content"), nil
}

type testPromptHandler struct{}

func (h *testPromptHandler) GetPrompt(ctx *runtime.Context, args map[string]interface{}) (string, error) {
	return "test prompt result", nil
}