package runtime

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hyperleex/zenmcp/protocol"
	"github.com/hyperleex/zenmcp/registry"
)

type mockToolHandler struct {
	called bool
}

func (h *mockToolHandler) Call(ctx interface{}, args json.RawMessage) (*protocol.ToolCallResult, error) {
	h.called = true
	return &protocol.ToolCallResult{
		Content: []protocol.Content{{Type: "text", Text: "mock result"}},
	}, nil
}

func TestRouter_Route_Initialize(t *testing.T) {
	reg := registry.New()
	router := NewRouter(reg)
	
	ctx := NewContext(context.Background(), protocol.NewRequestID("test"))
	
	initReq := protocol.InitializeRequest{
		ProtocolVersion: "2024-11-05",
		Capabilities:    protocol.ClientCapabilities{},
		ClientInfo: protocol.ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}
	
	params, err := json.Marshal(initReq)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	
	result, err := router.Route(ctx, protocol.MethodInitialize, params)
	if err != nil {
		t.Fatalf("Route error: %v", err)
	}
	
	initResult, ok := result.(*protocol.InitializeResult)
	if !ok {
		t.Fatalf("Expected InitializeResult, got %T", result)
	}
	
	if initResult.ProtocolVersion != "2024-11-05" {
		t.Errorf("Expected protocol version 2024-11-05, got %s", initResult.ProtocolVersion)
	}
	
	if initResult.ServerInfo.Name != "zenmcp-server" {
		t.Errorf("Expected server name zenmcp-server, got %s", initResult.ServerInfo.Name)
	}
}

func TestRouter_Route_ToolsList(t *testing.T) {
	reg := registry.New()
	handler := &mockToolHandler{}
	reg.RegisterTool("test_tool", "Test tool", handler, nil)
	
	router := NewRouter(reg)
	ctx := NewContext(context.Background(), protocol.NewRequestID("test"))
	
	result, err := router.Route(ctx, protocol.MethodToolsList, nil)
	if err != nil {
		t.Fatalf("Route error: %v", err)
	}
	
	toolsResult, ok := result.(*protocol.ToolListResult)
	if !ok {
		t.Fatalf("Expected ToolListResult, got %T", result)
	}
	
	if len(toolsResult.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(toolsResult.Tools))
	}
	
	if toolsResult.Tools[0].Name != "test_tool" {
		t.Errorf("Expected tool name test_tool, got %s", toolsResult.Tools[0].Name)
	}
}

func TestRouter_Route_ToolsCall(t *testing.T) {
	reg := registry.New()
	handler := &mockToolHandler{}
	reg.RegisterTool("test_tool", "Test tool", handler, nil)
	
	router := NewRouter(reg)
	ctx := NewContext(context.Background(), protocol.NewRequestID("test"))
	
	callReq := protocol.ToolCallRequest{
		Name:      "test_tool",
		Arguments: json.RawMessage(`{}`),
	}
	
	params, err := json.Marshal(callReq)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	
	result, err := router.Route(ctx, protocol.MethodToolsCall, params)
	if err != nil {
		t.Fatalf("Route error: %v", err)
	}
	
	callResult, ok := result.(*protocol.ToolCallResult)
	if !ok {
		t.Fatalf("Expected ToolCallResult, got %T", result)
	}
	
	if len(callResult.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(callResult.Content))
	}
	
	if callResult.Content[0].Text != "mock result" {
		t.Errorf("Expected content 'mock result', got %s", callResult.Content[0].Text)
	}
	
	if !handler.called {
		t.Error("Expected handler to be called")
	}
}

func TestRouter_Route_MethodNotFound(t *testing.T) {
	reg := registry.New()
	router := NewRouter(reg)
	ctx := NewContext(context.Background(), protocol.NewRequestID("test"))
	
	_, err := router.Route(ctx, "unknown_method", nil)
	if err == nil {
		t.Fatal("Expected error for unknown method")
	}
	
	mcpErr, ok := err.(*protocol.Error)
	if !ok {
		t.Fatalf("Expected protocol.Error, got %T", err)
	}
	
	if mcpErr.Code != protocol.MethodNotFound {
		t.Errorf("Expected code %d, got %d", protocol.MethodNotFound, mcpErr.Code)
	}
}

func TestRouter_Route_ToolNotFound(t *testing.T) {
	reg := registry.New()
	router := NewRouter(reg)
	ctx := NewContext(context.Background(), protocol.NewRequestID("test"))
	
	callReq := protocol.ToolCallRequest{
		Name:      "unknown_tool",
		Arguments: json.RawMessage(`{}`),
	}
	
	params, err := json.Marshal(callReq)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	
	_, err = router.Route(ctx, protocol.MethodToolsCall, params)
	if err == nil {
		t.Fatal("Expected error for unknown tool")
	}
	
	mcpErr, ok := err.(*protocol.Error)
	if !ok {
		t.Fatalf("Expected protocol.Error, got %T", err)
	}
	
	if mcpErr.Code != protocol.MethodNotFound {
		t.Errorf("Expected code %d, got %d", protocol.MethodNotFound, mcpErr.Code)
	}
}