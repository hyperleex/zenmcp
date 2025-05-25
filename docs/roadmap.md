# Roadmap: FastMCP-Go - A Go Library for Rapid MCP Development

## Introduction

This document outlines the plan for creating FastMCP-Go, a Go language library designed to simplify the development of applications using the Model Context Protocol (MCP). MCP is a standard that enables AI applications to connect with various data sources and tools, facilitating seamless integration and interaction. The goal of FastMCP-Go is to provide a Go-idiomatic equivalent to the ease-of-use features found in Python's FastMCP library.

## Understanding "FastMCP" (Python Context)

FastMCP in Python (e.g., `mcp.server.fastmcp.FastMCP`) is a component of the official Python MCP SDK. It aims to make MCP server development straightforward and efficient.

Its key features include:

*   **Ease of Use via Decorators:** Simplifies the registration of tools, resources, and prompt handlers.
*   **Type Hints for Schema Generation:** Leverages Python type hints to automatically generate MCP schema definitions.
*   **Lifespan Management:** Provides hooks for `startup` and `shutdown` events.
*   **Context Object:** Offers a convenient context object (`MCPContext`) for accessing MCP utilities and request-specific information.
*   **Transport Abstraction:** Decouples the server logic from the underlying transport mechanisms (e.g., Stdio, HTTP).
*   **ASGI Integration:** Works seamlessly with ASGI servers for HTTP-based transports.
*   **CLI Tooling:** Often accompanied by CLI tools for development and testing.
*   **Authentication Support:** Includes mechanisms for securing MCP endpoints.

The primary objective of FastMCP-Go is to deliver these benefits within the Go ecosystem, adhering to Go's idiomatic practices and leveraging its strengths.

## Core MCP Concepts for Go Implementation

The FastMCP-Go library must provide robust support for the following essential MCP features:

*   **MCP Lifecycle Management:**
    *   `initialize` method: To set up the MCP application, returning `ServerInfo`.
    *   `initialized` notification: Sent by the client after it has processed `ServerInfo`.
    *   `shutdown` method: To gracefully terminate the MCP application.
*   **MCP Tools:**
    *   Definition: Mechanisms to define available tools.
    *   Invocation: Handling `tool/invoke` requests.
    *   Listing: Handling `tool/list` requests.
    *   Annotations: Support for tool-specific annotations.
*   **MCP Resources:**
    *   Definition: Mechanisms to define available resources.
    *   Reading: Handling `resource/read` requests.
    *   Listing: Handling `resource/list` requests.
    *   Templates: Support for `resource/template` definitions and instantiation.
    *   Optional Subscriptions: Handling `resource/list_changed_subscribe` and sending `resource/list_changed` notifications.
*   **MCP Prompts:**
    *   Definition: Mechanisms to define available prompts.
    *   Retrieval: Handling `prompt/retrieve` requests.
    *   Listing: Handling `prompt/list` requests.
*   **Standard MCP Transports:**
    *   Stdio: For communication over standard input/output.
    *   Streamable HTTP: For HTTP-based communication, typically using Server-Sent Events (SSE) for streaming, including session management.
*   **Core Protocol Requirements:**
    *   JSON-RPC 2.0: Adherence to the JSON-RPC 2.0 specification for message formatting and request/response handling.
    *   Message Serialization/Deserialization: Efficient and correct handling of MCP message types.
    *   Error Handling: Standardized MCP error responses.
*   **Utility Features:**
    *   Logging: Integrated logging capabilities.
    *   Progress Tracking: Support for `$/progress` notifications.
    *   Cancellation: Support for `$/cancelRequest` notifications.
    *   Completion: (If applicable to specific use cases, though less common in basic MCP server/client).
*   **Optional: Client-Side Features:**
    *   Roots: Managing root objects for client-side interactions.
    *   Sampling: Client-side support for data sampling if needed.

## Proposed Library Structure for FastMCP-Go

The proposed directory and package structure for FastMCP-Go is as follows:

```
fastmcp-go/
├── server/               # Server-side logic, "FastMCP" style registration
│   └── fastmcp.go        # Main server component
├── client/               # Client-side library (Optional)
│   └── client.go         # Client implementation
├── transport/            # Transport interfaces and implementations
│   ├── stdio.go
│   └── http.go
├── protocol/             # Go structs for MCP message types (requests, responses, notifications)
│   └── types.go
├── schema/               # Go type to JSON schema generation utilities
│   └── generator.go
├── internal/             # Internal utility packages
│   └── jsonrpc/          # JSON-RPC 2.0 handling specifics
└── examples/             # Example MCP applications using FastMCP-Go
    ├── basic_stdio_server/
    └── basic_http_server/
```

**Explanation of Key Packages:**

*   **`protocol`**: Contains Go struct definitions for all MCP message types (requests, responses, notifications, and their parameters/results). This ensures type safety when working with MCP messages.
*   **`server`**: Implements the server-side logic. This package will provide the "FastMCP" style API for easily registering tools, resources, and prompt handlers, similar to Python's decorator-based approach but using Go idioms (e.g., functional registration, struct embedding).
*   **`transport`**: Defines interfaces for different communication transports (e.g., `TransportConnection`) and provides concrete implementations for Stdio (`StdioTransport`) and Streamable HTTP (`HTTPTransport`).
*   **`schema`**: Provides utilities for generating JSON schema definitions from Go types. This is crucial for the `initialize` method's `ServerInfo` response, which includes schemas for tools, resources, and prompts.

## Development Steps/Phases for FastMCP-Go

The development of FastMCP-Go will proceed in the following phases:

1.  **Phase 0: Foundation & Core Protocol**
    *   Initialize Go module (`go mod init`).
    *   Define basic Go structs for core MCP message types (requests, responses, errors) in the `protocol` package based on the MCP specification.
    *   Implement JSON-RPC 2.0 request/response parsing and serialization utilities in `internal/jsonrpc`.
    *   Set up basic project structure and linting (`.golangci.yml`).

2.  **Phase 1: Stdio Transport & Basic Server Lifecycle**
    *   Implement the Stdio transport layer in the `transport` package.
    *   Develop the initial server structure in the `server` package.
    *   Implement server-side handling for `initialize` (returning static `ServerInfo` initially), `initialized`, `ping`, and `shutdown` methods.
    *   Create a basic example of a stdio-based server.

3.  **Phase 2: Server - Tools & Resources Implementation (Core Functionality)**
    *   Design and implement the "FastMCP" style registration mechanism for tools and resources within the `server` package (e.g., `server.RegisterTool()`, `server.RegisterResource()`).
    *   Implement the first version of schema generation (`schema/generator.go`) from Go function signatures/structs to JSON Schema for tools and resources.
    *   Integrate schema generation into the `initialize` response.
    *   Implement server-side handling for `tool/list`, `tool/invoke`, `resource/list`, and `resource/read`.
    *   Develop a server context object (v1) for passing request-specific data and utilities to handlers.

4.  **Phase 3: Streamable HTTP Transport**
    *   Implement the Streamable HTTP transport layer in the `transport` package, likely using Server-Sent Events (SSE) for asynchronous message streaming.
    *   Address HTTP-specific concerns like request multiplexing and session management.
    *   Create a basic example of an HTTP-based server.

5.  **Phase 4: Server - Prompts & Advanced Resource Features**
    *   Implement registration and handling for MCP Prompts (`prompt/list`, `prompt/retrieve`).
    *   Add support for advanced resource features:
        *   `resource/template` (defining and instantiating resource templates).
        *   `resource/list_changed_subscribe` and `resource/list_changed` notifications.
    *   Update schema generation to support prompts and these advanced resource features.

6.  **Phase 5: Client-Side Library (Optional but Recommended)**
    *   Design and implement a basic client library in the `client` package.
    *   Provide helper functions for common client tasks: connecting to a server, invoking tools, reading resources, and retrieving prompts.
    *   This phase can be deferred or developed in parallel if resources allow.

7.  **Phase 6: Utilities & Advanced MCP Features**
    *   Implement support for `$/progress` notifications (server-to-client).
    *   Implement support for `$/cancelRequest` notifications (client-to-server).
    *   Implement `$/completion` if deemed broadly useful.
    *   Integrate robust logging mechanisms.
    *   Explore and implement authentication/authorization hooks if required for common use cases.

8.  **Phase 7: Documentation, Testing, and Refinement**
    *   Write comprehensive unit and integration tests for all packages and features.
    *   Generate Godoc documentation for the public API.
    *   Create tutorials and usage examples.
    *   Refine the API based on usage and feedback, ensuring it is idiomatic and developer-friendly.
    *   Performance testing and optimization.

## Leveraging Go-Specific Idioms and Libraries

FastMCP-Go will be designed to feel natural for Go developers by leveraging:

*   **Concurrency:**
    *   Goroutines for handling concurrent requests and managing individual client connections.
    *   Goroutines for streaming data in SSE.
    *   Channels for internal communication between different parts of the library (e.g., transport layer and server logic, progress updates).
*   **Strong Static Typing:**
    *   Protocol messages (`protocol` package) will be strongly-typed Go structs, providing compile-time safety.
    *   Handler functions registered with the server will have type-safe signatures.
*   **Interfaces:**
    *   Define clear interfaces for abstractions like `TransportConnection`, `ToolHandler`, `ResourceHandler`, promoting modularity and testability.
*   **Standard Library:**
    *   `encoding/json` for JSON serialization/deserialization.
    *   `net/http` for the HTTP transport.
    *   `os/exec` (if needed for specific tool implementations, though generally tools are Go functions).
    *   `reflect` package for runtime type introspection, crucial for the `schema` generator.
    *   `context` package for request lifecycle management, cancellation, and passing deadlines.
*   **Struct Tags:**
    *   Utilize struct tags (e.g., `mcp:""`) on Go types used for tools, resources, and prompts to provide metadata for schema generation, similar to how `json:""` tags are used for JSON marshaling. This can help customize names, descriptions, and validation rules.
*   **Error Handling:**
    *   Employ idiomatic Go error handling patterns (e.g., returning `error` as the last value, using custom error types).
*   **Build System:**
    *   Standard Go modules for dependency management.
    *   `go test` for running unit and integration tests.

## Conclusion

The development of FastMCP-Go aims to provide the Go community with a powerful yet easy-to-use library for building MCP-compliant applications. By mirroring the developer experience of Python's FastMCP while fully embracing Go's strengths—such as concurrency, static typing, and performance—FastMCP-Go will empower developers to create robust, efficient, and maintainable MCP integrations. This roadmap provides a structured approach to achieving this goal, focusing on iterative development and core MCP compliance.
