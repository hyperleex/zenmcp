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
	"github.com/hyperleex/zenmcp/runtime"
	"github.com/hyperleex/zenmcp/transport/stdio"
)

// GreetingArgs defines the input structure for the greeting tool
type GreetingArgs struct {
	Name string `json:"name" description:"The name to greet"`
	Age  *int   `json:"age,omitempty" description:"Optional age"`
}

// AddArgs defines the input structure for the add tool
type AddArgs struct {
	A int `json:"a" description:"First number"`
	B int `json:"b" description:"Second number"`
}

func main() {
	// Create server with stdio transport
	transport := stdio.New()
	server := mcp.NewServer(transport)

	// Register tools using the improved type-safe API
	err := mcp.RegisterToolFunc(server, "greet", "Generate a personalized greeting", 
		func(ctx *runtime.Context, args GreetingArgs) (*protocol.ToolCallResult, error) {
			greeting := fmt.Sprintf("Hello, %s!", args.Name)
			if args.Age != nil {
				greeting += fmt.Sprintf(" You are %d years old.", *args.Age)
			}
			
			return &protocol.ToolCallResult{
				Content: []protocol.Content{{
					Type: "text",
					Text: greeting,
				}},
			}, nil
		})
	if err != nil {
		log.Fatalf("Failed to register greet tool: %v", err)
	}

	err = mcp.RegisterToolFunc(server, "add", "Add two numbers together",
		func(ctx *runtime.Context, args AddArgs) (*protocol.ToolCallResult, error) {
			result := args.A + args.B
			
			return &protocol.ToolCallResult{
				Content: []protocol.Content{{
					Type: "text", 
					Text: fmt.Sprintf("%d + %d = %d", args.A, args.B, result),
				}},
			}, nil
		})
	if err != nil {
		log.Fatalf("Failed to register add tool: %v", err)
	}

	// Register a resource using the improved API
	mcp.RegisterResourceFunc(server, "config://settings", "settings", "Application settings", "application/json",
		func(ctx *runtime.Context, uri string) ([]byte, string, error) {
			data := `{"version": "1.0.0", "debug": true}`
			return []byte(data), "application/json", nil
		})

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server...")
		cancel()
	}()

	log.Println("Starting ZenMCP server with improved type-safe API...")
	if err := server.Serve(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped")
}