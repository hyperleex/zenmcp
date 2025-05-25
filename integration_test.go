package zenmcp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/hyperleex/zenmcp/mcp"
	"github.com/hyperleex/zenmcp/protocol"
	"github.com/hyperleex/zenmcp/registry"
	"github.com/hyperleex/zenmcp/runtime"
	zhttp "github.com/hyperleex/zenmcp/transport/http"
)

// TestMVPIntegration tests the complete MVP functionality end-to-end
func TestMVPIntegration(t *testing.T) {
	// Create a test server with tools, resources, and prompts
	server := mcp.NewServer(
		mcp.WithHTTPTransport(":0"), // Use random port
	)

	// Add a simple echo tool
	server.RegisterTool("echo", registry.ToolDescriptor{
		Name:        "echo",
		Description: "Echo back the input message",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "Message to echo back",
				},
			},
			"required": []string{"message"},
		},
	}, func(ctx *runtime.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
		message, ok := args["message"].(string)
		if !ok {
			return nil, fmt.Errorf("message must be a string")
		}
		return &protocol.ToolResult{
			Content: []protocol.Content{{
				Type: "text",
				Text: fmt.Sprintf("Echo: %s", message),
			}},
		}, nil
	})

	// Add a math tool
	server.RegisterTool("add", registry.ToolDescriptor{
		Name:        "add",
		Description: "Add two numbers",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"a": map[string]interface{}{
					"type":        "number",
					"description": "First number",
				},
				"b": map[string]interface{}{
					"type":        "number",
					"description": "Second number",
				},
			},
			"required": []string{"a", "b"},
		},
	}, func(ctx *runtime.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
		a, okA := args["a"].(float64)
		b, okB := args["b"].(float64)
		if !okA || !okB {
			return nil, fmt.Errorf("both a and b must be numbers")
		}
		result := a + b
		return &protocol.ToolResult{
			Content: []protocol.Content{{
				Type: "text",
				Text: fmt.Sprintf("Result: %.2f", result),
			}},
		}, nil
	})

	// Add a test resource
	server.RegisterResource("test/greeting", registry.ResourceDescriptor{
		URI:         "test://greeting",
		Name:        "greeting",
		Description: "A simple greeting resource",
		MimeType:    "text/plain",
	}, func(ctx *runtime.Context) (*protocol.ResourceResult, error) {
		return &protocol.ResourceResult{
			Contents: []protocol.Content{{
				Type: "text",
				Text: "Hello from ZenMCP!",
			}},
		}, nil
	})

	// Start server in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(ctx)
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	// Get the actual port the server is listening on
	addr := server.Addr()
	if addr == "" {
		t.Fatal("Server address is empty")
	}

	baseURL := "http://" + addr

	// Test 1: Initialize handshake
	t.Run("Initialize", func(t *testing.T) {
		initReq := protocol.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      "1",
			Method:  "initialize",
			Params: map[string]interface{}{
				"protocolVersion": "2025-03-26",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"clientInfo": map[string]interface{}{
					"name":    "zenmcp-test-client",
					"version": "1.0.0",
				},
			},
		}

		resp := makeJSONRPCRequest(t, baseURL, initReq)
		
		// Verify response structure
		if resp.Error != nil {
			t.Fatalf("Initialize failed: %v", resp.Error)
		}

		result, ok := resp.Result.(map[string]interface{})
		if !ok {
			t.Fatal("Initialize result is not an object")
		}

		// Check protocol version
		if result["protocolVersion"] != "2025-03-26" {
			t.Errorf("Expected protocol version 2025-03-26, got %v", result["protocolVersion"])
		}

		// Check server capabilities
		capabilities, ok := result["capabilities"].(map[string]interface{})
		if !ok {
			t.Fatal("Server capabilities not found")
		}

		if _, hasTools := capabilities["tools"]; !hasTools {
			t.Error("Server should advertise tools capability")
		}
	})

	// Test 2: List tools
	t.Run("ListTools", func(t *testing.T) {
		listReq := protocol.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      "2",
			Method:  "tools/list",
			Params:  map[string]interface{}{},
		}

		resp := makeJSONRPCRequest(t, baseURL, listReq)
		
		if resp.Error != nil {
			t.Fatalf("ListTools failed: %v", resp.Error)
		}

		result, ok := resp.Result.(map[string]interface{})
		if !ok {
			t.Fatal("ListTools result is not an object")
		}

		tools, ok := result["tools"].([]interface{})
		if !ok {
			t.Fatal("Tools list not found")
		}

		if len(tools) != 2 {
			t.Errorf("Expected 2 tools, got %d", len(tools))
		}

		// Verify tool names
		toolNames := make([]string, len(tools))
		for i, tool := range tools {
			toolObj, ok := tool.(map[string]interface{})
			if !ok {
				t.Fatal("Tool is not an object")
			}
			toolNames[i] = toolObj["name"].(string)
		}

		expectedTools := []string{"echo", "add"}
		for _, expected := range expectedTools {
			found := false
			for _, actual := range toolNames {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected tool %s not found", expected)
			}
		}
	})

	// Test 3: Call echo tool
	t.Run("CallEchoTool", func(t *testing.T) {
		callReq := protocol.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      "3",
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "echo",
				"arguments": map[string]interface{}{
					"message": "Hello, ZenMCP!",
				},
			},
		}

		resp := makeJSONRPCRequest(t, baseURL, callReq)
		
		if resp.Error != nil {
			t.Fatalf("CallEchoTool failed: %v", resp.Error)
		}

		result, ok := resp.Result.(map[string]interface{})
		if !ok {
			t.Fatal("CallEchoTool result is not an object")
		}

		content, ok := result["content"].([]interface{})
		if !ok {
			t.Fatal("Tool result content not found")
		}

		if len(content) != 1 {
			t.Errorf("Expected 1 content item, got %d", len(content))
		}

		contentItem, ok := content[0].(map[string]interface{})
		if !ok {
			t.Fatal("Content item is not an object")
		}

		text, ok := contentItem["text"].(string)
		if !ok {
			t.Fatal("Content text not found")
		}

		expected := "Echo: Hello, ZenMCP!"
		if text != expected {
			t.Errorf("Expected %q, got %q", expected, text)
		}
	})

	// Test 4: Call math tool
	t.Run("CallMathTool", func(t *testing.T) {
		callReq := protocol.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      "4",
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "add",
				"arguments": map[string]interface{}{
					"a": 15.5,
					"b": 24.3,
				},
			},
		}

		resp := makeJSONRPCRequest(t, baseURL, callReq)
		
		if resp.Error != nil {
			t.Fatalf("CallMathTool failed: %v", resp.Error)
		}

		result, ok := resp.Result.(map[string]interface{})
		if !ok {
			t.Fatal("CallMathTool result is not an object")
		}

		content, ok := result["content"].([]interface{})
		if !ok {
			t.Fatal("Tool result content not found")
		}

		contentItem := content[0].(map[string]interface{})
		text := contentItem["text"].(string)

		expected := "Result: 39.80"
		if text != expected {
			t.Errorf("Expected %q, got %q", expected, text)
		}
	})

	// Test 5: List resources
	t.Run("ListResources", func(t *testing.T) {
		listReq := protocol.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      "5",
			Method:  "resources/list",
			Params:  map[string]interface{}{},
		}

		resp := makeJSONRPCRequest(t, baseURL, listReq)
		
		if resp.Error != nil {
			t.Fatalf("ListResources failed: %v", resp.Error)
		}

		result, ok := resp.Result.(map[string]interface{})
		if !ok {
			t.Fatal("ListResources result is not an object")
		}

		resources, ok := result["resources"].([]interface{})
		if !ok {
			t.Fatal("Resources list not found")
		}

		if len(resources) != 1 {
			t.Errorf("Expected 1 resource, got %d", len(resources))
		}

		resource := resources[0].(map[string]interface{})
		if resource["name"] != "greeting" {
			t.Errorf("Expected resource name 'greeting', got %v", resource["name"])
		}
	})

	// Test 6: Read resource
	t.Run("ReadResource", func(t *testing.T) {
		readReq := protocol.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      "6",
			Method:  "resources/read",
			Params: map[string]interface{}{
				"uri": "test://greeting",
			},
		}

		resp := makeJSONRPCRequest(t, baseURL, readReq)
		
		if resp.Error != nil {
			t.Fatalf("ReadResource failed: %v", resp.Error)
		}

		result, ok := resp.Result.(map[string]interface{})
		if !ok {
			t.Fatal("ReadResource result is not an object")
		}

		contents, ok := result["contents"].([]interface{})
		if !ok {
			t.Fatal("Resource contents not found")
		}

		content := contents[0].(map[string]interface{})
		text := content["text"].(string)

		expected := "Hello from ZenMCP!"
		if text != expected {
			t.Errorf("Expected %q, got %q", expected, text)
		}
	})

	// Cleanup
	cancel()
	
	// Wait for server to shutdown
	select {
	case err := <-errCh:
		if err != nil && !strings.Contains(err.Error(), "context canceled") {
			t.Errorf("Server error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Server did not shutdown within timeout")
	}
}

// Helper function to make JSON-RPC HTTP requests
func makeJSONRPCRequest(t *testing.T, baseURL string, req protocol.JSONRPCRequest) *protocol.JSONRPCResponse {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", baseURL, strings.NewReader(string(reqBytes)))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("Failed to make HTTP request: %v", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		t.Fatalf("HTTP request failed with status %d", httpResp.StatusCode)
	}

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var resp protocol.JSONRPCResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	return &resp
}

// TestStdioMVP tests the stdio transport MVP functionality
func TestStdioMVP(t *testing.T) {
	// This test would require starting a subprocess and communicating via stdio
	// For MVP, we'll just verify that the demo server can be built and run
	t.Run("BuildDemoServer", func(t *testing.T) {
		cmd := exec.Command("go", "build", "-o", "/tmp/zenmcp-demo", "./cmd/server")
		cmd.Dir = "/Users/lee/dev/zenmcp"
		
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to build demo server: %v\nOutput: %s", err, output)
		}
	})

	t.Run("RunDemoServerHelp", func(t *testing.T) {
		cmd := exec.Command("/tmp/zenmcp-demo", "-help")
		
		output, err := cmd.CombinedOutput()
		if err != nil {
			// -help typically exits with code 1, which is normal
			if !strings.Contains(string(output), "Usage:") {
				t.Fatalf("Demo server help output unexpected: %v\nOutput: %s", err, output)
			}
		}

		if !strings.Contains(string(output), "stdio") {
			t.Error("Demo server should support stdio mode")
		}
	})
}