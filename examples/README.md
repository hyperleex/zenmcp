# ZenMCP Examples

This directory contains examples and tutorials for using the ZenMCP framework.

## Quick Start

1. **Build the binaries:**
   ```bash
   make build
   ```

2. **Start the demo server:**
   ```bash
   ./bin/zenmcp-server -addr=:8080
   ```

3. **Test with the CLI client:**
   ```bash
   ./bin/zenmcp-client -cmd=list-tools
   ```

## Examples

### [Basic Server](./basic-server/)
A minimal MCP server with one tool and one resource.

### [HTTP Client](./http-client/)  
Example of creating an HTTP client to connect to an MCP server.

### [Stdio Server](./stdio-server/)
Example of a server that communicates via stdin/stdout for local plugins.

### [Advanced Tools](./advanced-tools/)
Examples of more complex tools with validation, progress notifications, and error handling.

## Demo Commands

The Makefile includes several demo commands:

```bash
# Demo the echo tool
make demo-echo

# Demo the math tool
make demo-math  

# Demo resources
make demo-resources

# Quick start guide
make quickstart
```

## Testing

Run the integration tests to verify everything works:

```bash
make integration
```