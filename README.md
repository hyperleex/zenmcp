# ZenMCP - Minimalist Go Framework for MCP Servers

A lightweight, high-performance Go framework for building Model Context Protocol servers.
Designed with Go's philosophy of simplicity in mind - write less code, deploy faster,
scale effortlessly.

✨ Zero external dependencies
🚀 Single binary deployment
⚡ High performance with low memory footprint
🔧 Simple, intuitive API
🌍 Cross-platform support
📡 HTTP and stdio transports
🛠️ Tools, Resources, and Prompts support

## MVP Status

The ZenMCP MVP is **ready for use**! The framework includes:

### ✅ Core Features
- **JSON-RPC 2.0** message handling with MCP 2025-03-26 protocol
- **HTTP Transport** with Server-Sent Events for streaming
- **Stdio Transport** for local plugin communication
- **Tools** - Execute functions with typed parameters
- **Resources** - Serve static or dynamic content
- **Context** - Request cancellation and progress tracking
- **Registry** - JSON Schema generation for tool parameters

### ✅ Transports
- **HTTP** - RESTful API with streaming support
- **Stdio** - Length-prefixed stdin/stdout for local plugins

### ✅ Tooling
- **Demo Server** (`cmd/server`) - Full-featured example server
- **CLI Client** (`cmd/client`) - Interactive testing client
- **Integration Tests** - End-to-end verification
- **Examples** - Copy-paste tutorials and patterns

### 🔄 Planned Features (Post-MVP)
- OAuth 2.1 authentication
- Rate limiting middleware
- Service discovery (mDNS + Unix sockets)
- Health check endpoints
- Metrics and logging middleware

## Quick Start

### 1. Install

```bash
go get github.com/hyperleex/zenmcp
```

### 2. Build and Run Demo

```bash
# Build binaries
make build

# Start demo server
./bin/zenmcp-server -addr=:8080

# Test with CLI client (in another terminal)
./bin/zenmcp-client -cmd=list-tools
./bin/zenmcp-client -cmd=call-tool -tool=echo -args='{"message":"Hello!"}'
```

### 3. Quick Demo Commands

```bash
# Demo the echo tool
make demo-echo

# Demo the math tool
make demo-math

# Demo resources
make demo-resources
```

## Basic Server Example

```go
package main

import (
    "context"
    "fmt"
    "github.com/hyperleex/zenmcp/mcp"
    "github.com/hyperleex/zenmcp/protocol"
    "github.com/hyperleex/zenmcp/registry"
    "github.com/hyperleex/zenmcp/runtime"
)

func main() {
    // Create server with HTTP transport
    server := mcp.NewServer(mcp.WithHTTPTransport(":8080"))

    // Register a tool
    server.RegisterTool("greet", registry.ToolDescriptor{
        Name: "greet",
        Description: "Generate a greeting",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "name": map[string]interface{}{
                    "type": "string",
                    "description": "Name to greet",
                },
            },
            "required": []string{"name"},
        },
    }, func(ctx *runtime.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
        name := args["name"].(string)
        return &protocol.ToolResult{
            Content: []protocol.Content{{
                Type: "text",
                Text: fmt.Sprintf("Hello, %s!", name),
            }},
        }, nil
    })

    // Start server
    server.Serve(context.Background())
}
```

## Testing

```bash
# Run all tests
make test

# Run integration tests
make integration

# Run with race detection
make race

# Run benchmarks
make bench
```

## Development

```bash
# Format code
make fmt

# Run linting
make lint

# Full CI pipeline
make ci
```

## Examples

See the [`examples/`](./examples/) directory for:

- **Basic Server** - Minimal MCP server setup
- **HTTP Client** - Connecting to servers programmatically
- **Advanced Tools** - Complex tools with validation
- **Stdio Server** - Local plugin communication

## Architecture

ZenMCP follows a layered architecture:

```
mcp/              # Public API (Server, Client)
├─ protocol/      # JSON-RPC & MCP message types
├─ transport/     # HTTP and stdio transports
├─ runtime/       # Router, Context, validation
├─ registry/      # Tool/Resource descriptors
└─ middleware/    # Cross-cutting concerns (planned)
```

**Import rule**: Upper layers import lower layers, never the opposite.

## Performance

- **Sub-microsecond routing** with generics and zero reflection
- **Minimal memory allocation** in hot paths
- **Efficient connection handling** with Go's concurrent model
- **Streamable responses** for large datasets

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

1. Read [`docs/PROJECT.md`](./docs/PROJECT.md) for architecture details
2. Check current development stage and roadmap
3. Follow the development workflow outlined in project docs
4. All PRs must include tests and pass CI

## Support

- 📚 [Documentation](./docs/)
- 🐛 [Issues](https://github.com/hyperleex/zenmcp/issues)
- 💬 [Discussions](https://github.com/hyperleex/zenmcp/discussions)
