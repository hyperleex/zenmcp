# zenmcp — Project Overview & Specification Traceability

> **Purpose**: single authoritative artifact that tells a human developer **and** an LLM *what* we are building, *how* it is laid out, *why* the architecture looks this way, and *which* requirements from the official **Model Context Protocol (MCP 2025‑03‑26)** are satisfied.
>
> Keep this file in the repo root — link to it from every design discussion so context is never lost.

---

## 1 Elevator Pitch

*Build once, expose everywhere.*
`zenmcp` is a **Go SDK** that turns plain Go functions into MCP‑compliant endpoints. It ships two built‑in transports (Streamable HTTP, stdio) and optional production modules (OAuth 2.1, discovery, metrics, rate‑limit). The SDK mirrors the ergonomics of FastMCP but with Go generics, zero reflection on hot path, and idiomatic middleware.

---

## 2 Filesystem Layout (high‑level)

```
📦 zenmcp/
├─ protocol/         # JSON‑RPC & MCP messages, codec
├─ transport/
│  ├─ http/          # Streamable HTTP (POST+SSE)
│  ├─ stdio/         # Length‑prefixed stdin/stdout
│  └─ interfaces.go  # Transport abstraction
├─ registry/         # Tool/Resource/Prompt descriptors, JSON‑Schema gen
├─ runtime/          # Router, Context, validation, progress
├─ mcp/              # Public API (Server, Client, helpers)
├─ middleware/       # log, recover, metrics, rate, toolsanity
├─ auth/             # OAuth 2.1 resource‑server & client creds provider
├─ discover/         # mdns + Unix‑socket local registry (opt‑in)
├─ health/           # /livez /readyz endpoints, graceful shutdown helpers
├─ cmd/              # reference binaries (demo server, CLI)
├─ examples/         # copy‑paste tutorials
└─ internal/         # interop & fuzz tests
```

> **Rule of thumb**: upper layers may import lower; never the opposite.

---

## 3 Architectural Decisions (A‑series)

| ID       | Decision                                                | Rationale                                                 |
| -------- | ------------------------------------------------------- | --------------------------------------------------------- |
| **A‑01** | Streamable HTTP as default transport                    | Matches MCP §4.2; easier to proxy than WebSocket.         |
| **A‑02** | Stdio transport included                                | Required for local plugin model (spec §4.3).              |
| **A‑03** | Generics + zero reflection in hot path                  | Sub‑µs routing, compile‑time type safety.                 |
| **A‑04** | Middleware chain `recover→log→auth→rate→metrics→router` | Separation of x‑cutting concerns, configurable order.     |
| **A‑05** | Optional modules via `ServerOption`                     | Core remains dependency‑light; prod features opt‑in.      |
| **A‑06** | OAuth 2.1 as OPTIONAL                                   | Spec marks security as optional; still provided for prod. |
| **A‑07** | Discovery with mdns + UDS                               | Multiple MCP servers on same host w/o port juggling.      |

---

## 4 Functional Requirements ↔ Spec Traceability

### 4.1 MUST

| Spec § | Requirement                            | Package(s)            | Status                |
| ------ | -------------------------------------- | --------------------- | --------------------- |
| 3.1    | JSON‑RPC 2.0 message format            | `protocol`            |  plannded|
| 4.2    | Streamable HTTP transport              | `transport/http`      | planned             |
| 5.2    | Capability exchange (`serverFeatures`) | `runtime`, `registry` | planned               |
| 5.3    | Tool execution `tool/run`              | `runtime`             | planned               |
| 6.1    | Progress notifications                 | `runtime.Context`     | planned               |
| 6.2    | Cancellation                           | `runtime.Context`     | planned               |

### 4.2 OPTIONAL (SHOULD / MAY)

| Capability               | Spec §     | Package(s)                | Notes    |
| ------------------------ | ---------- | ------------------------- | -------- |
| OAuth 2.1                | 7          | `auth`, `middleware/auth` | Stage 8  |
| Logging min‑level        | 8.1        | `middleware/log`          | Stage 7  |
| Rate limits              | 8.2        | `middleware/rate`         | Stage 7  |
| Discovery                | 9.3        | `discover`                | Stage 9  |
| Tool Best Practices lint | docs/tools | `middleware/toolsanity`   | Stage 10 |

---

## 5 Roadmap Snapshot
 Current focus: **Stage 2**.

---

## 6 Development Workflow for Humans & LLMs

1. **Read this file**.
2. Ask: *“Which stage is active?”* (check Roadmap).
   LLM should limit scope to that stage.
3. Break stage into atomic tasks (≤ 200 LOC each).
4. For each task: design → implement → add unit test → run `task lint test race`.
5. Open PR referencing A‑series decision or update decision if changed.

> Tip for LLM: start with interface stubs, then fill logic, then write tests.

---

## 7 Quality Gates

* `go vet`, `staticcheck`, `golangci-lint` ✓
* Unit coverage target ≥ 90 % for protocol & runtime.
* Router benchmark ≤ 1 µs/op (P99).
* Interop: fastmcp ⇄ zenmcp end‑to‑end suite.

---

## 8 Contribution Guidelines (abridged)

* No new external dependency without maintainer approval.
* Public API changes → open `design/NN-proposal.md`.
* Each PR must include tests and docs.

---
