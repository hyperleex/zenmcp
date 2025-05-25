# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ZenMCP is a minimalist Go framework for building Model Context Protocol servers. The project emphasizes zero external dependencies, single binary deployment, and high performance with Go's philosophy of simplicity.

## Development Commands

Use Task runner for all development tasks:

- `task lint` - Run golangci-lint on all packages
- `task test` - Run all tests
- `task race` - Run tests with race condition detection

Standard Go commands also work:
- `go test ./...` - Run all tests
- `go test -race ./...` - Run tests with race detection
- `go vet ./...` - Run Go vet

## Architecture

The codebase follows a layered architecture with clear separation of concerns:

```
protocol/         # JSON-RPC & MCP messages, codec (foundation layer)
transport/        # HTTP and stdio transports with abstraction
registry/         # Tool/Resource/Prompt descriptors, JSON-Schema generation
runtime/          # Router, Context, validation, progress (core engine)
mcp/              # Public API (Server, Client, helpers)
middleware/       # Cross-cutting concerns (log, recover, metrics, rate, toolsanity)
auth/             # OAuth 2.1 implementation (optional)
discover/         # mdns + Unix-socket local registry (optional)
health/           # Health check endpoints and graceful shutdown
```

**Import rule**: Upper layers may import lower layers, never the opposite.

## Key Architectural Decisions

- **A-01**: Streamable HTTP as default transport (matches MCP spec §4.2)
- **A-02**: Stdio transport included (required for local plugin model)
- **A-03**: Generics + zero reflection in hot path (sub-µs routing)
- **A-04**: Middleware chain: recover→log→auth→rate→metrics→router
- **A-05**: Optional modules via `ServerOption` (keeps core lightweight)

## Development Workflow

The project follows a staged development approach. Check `docs/PROJECT.md` for the current active stage. When implementing:

1. Start with interface stubs
2. Fill in logic implementation
3. Add unit tests
4. Run `task lint test race` before committing

## Quality Requirements

- Unit coverage target ≥ 90% for protocol & runtime packages
- Router benchmark ≤ 1 µs/op (P99)
- All code must pass `go vet`, `staticcheck`, and `golangci-lint`
- No new external dependencies without maintainer approval

i used taskfile instead of makefile.
When finish current stage change docs/PROJECT.md stage.
Write what already maked and what next.
