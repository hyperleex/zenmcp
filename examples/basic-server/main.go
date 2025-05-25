package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hyperleex/zenmcp/mcp"
	"github.com/hyperleex/zenmcp/protocol"
	"github.com/hyperleex/zenmcp/registry"
	"github.com/hyperleex/zenmcp/runtime"
)

func main() {
	// Create a new MCP server with HTTP transport
	server := mcp.NewServer(
		mcp.WithHTTPTransport(":8080"),
	)

	// Register a simple greeting tool
	server.RegisterTool("greet", registry.ToolDescriptor{
		Name:        "greet",
		Description: "Generate a personalized greeting",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the person to greet",
				},
				"language": map[string]interface{}{
					"type":        "string",
					"description": "Language for the greeting (en, es, fr)",
					"enum":        []string{"en", "es", "fr"},
					"default":     "en",
				},
			},
			"required": []string{"name"},
		},
	}, func(ctx *runtime.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
		name, ok := args["name"].(string)
		if !ok {
			return nil, fmt.Errorf("name must be a string")
		}

		language, ok := args["language"].(string)
		if !ok {
			language = "en" // default
		}

		var greeting string
		switch language {
		case "es":
			greeting = fmt.Sprintf("¡Hola, %s!", name)
		case "fr":
			greeting = fmt.Sprintf("Bonjour, %s!", name)
		default:
			greeting = fmt.Sprintf("Hello, %s!", name)
		}

		return &protocol.ToolResult{
			Content: []protocol.Content{{
				Type: "text",
				Text: greeting,
			}},
		}, nil
	})

	// Register a simple info resource
	server.RegisterResource("server/info", registry.ResourceDescriptor{
		URI:         "zenmcp://server/info",
		Name:        "server-info",
		Description: "Information about this ZenMCP server",
		MimeType:    "application/json",
	}, func(ctx *runtime.Context) (*protocol.ResourceResult, error) {
		info := map[string]interface{}{
			"name":        "Basic ZenMCP Server",
			"version":     "1.0.0",
			"description": "A minimal example of a ZenMCP server",
			"features":    []string{"tools", "resources"},
			"transport":   "http",
		}

		// Convert to JSON string for text content
		jsonStr := `{
  "name": "Basic ZenMCP Server",
  "version": "1.0.0",
  "description": "A minimal example of a ZenMCP server",
  "features": ["tools", "resources"],
  "transport": "http"
}`

		return &protocol.ResourceResult{
			Contents: []protocol.Content{{
				Type: "text",
				Text: jsonStr,
			}},
		}, nil
	})

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, stopping server...")
		cancel()
	}()

	// Start the server
	log.Println("Starting Basic ZenMCP Server on :8080")
	log.Println("Try: curl -X POST http://localhost:8080 -H 'Content-Type: application/json' -d '{\"jsonrpc\":\"2.0\",\"id\":\"1\",\"method\":\"tools/list\",\"params\":{}}'")

	if err := server.Serve(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped")
}
