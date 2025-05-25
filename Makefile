# ZenMCP MVP Makefile
.PHONY: help build test clean demo client integration lint bench install

# Default target
help: ## Show this help message
	@echo "ZenMCP MVP Build Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Build targets
build: ## Build all binaries
	@echo "Building ZenMCP binaries..."
	@go build -o bin/zenmcp-server ./cmd/server
	@go build -o bin/zenmcp-client ./cmd/client
	@echo "Built: bin/zenmcp-server, bin/zenmcp-client"

demo: ## Build demo server only
	@echo "Building demo server..."
	@go build -o bin/zenmcp-server ./cmd/server
	@echo "Built: bin/zenmcp-server"

client: ## Build CLI client only
	@echo "Building CLI client..."
	@go build -o bin/zenmcp-client ./cmd/client
	@echo "Built: bin/zenmcp-client"

# Test targets
test: ## Run unit tests
	@echo "Running unit tests..."
	@go test -v ./...

integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -run "TestMVPIntegration|TestStdioMVP" .

race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	@go test -race ./...

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Quality targets
lint: ## Run linting (requires golangci-lint)
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, running basic checks..."; \
		go vet ./...; \
		go fmt ./...; \
	fi

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

# Utility targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean ./...

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Installation targets
install: ## Install binaries to $GOPATH/bin
	@echo "Installing binaries..."
	@go install ./cmd/server
	@go install ./cmd/client

# Development targets
dev-server: demo ## Run demo server in development mode
	@echo "Starting demo server on :8080..."
	@./bin/zenmcp-server -addr=:8080

dev-stdio: demo ## Run demo server in stdio mode
	@echo "Starting demo server in stdio mode..."
	@./bin/zenmcp-server -stdio

# Demo scenarios
demo-echo: build ## Demo: Echo tool via HTTP
	@echo "Demo: Echo tool"
	@echo "Starting server in background..."
	@./bin/zenmcp-server -addr=:8081 & \
	SERVER_PID=$$!; \
	sleep 2; \
	echo "Calling echo tool..."; \
	./bin/zenmcp-client -addr=localhost:8081 -cmd=call-tool -tool=echo -args='{"message":"Hello from ZenMCP MVP!"}'; \
	echo "Stopping server..."; \
	kill $$SERVER_PID || true

demo-math: build ## Demo: Math tool via HTTP  
	@echo "Demo: Math tool"
	@echo "Starting server in background..."
	@./bin/zenmcp-server -addr=:8082 & \
	SERVER_PID=$$!; \
	sleep 2; \
	echo "Calling add tool..."; \
	./bin/zenmcp-client -addr=localhost:8082 -cmd=call-tool -tool=add -args='{"a":42,"b":8}'; \
	echo "Stopping server..."; \
	kill $$SERVER_PID || true

demo-resources: build ## Demo: List and read resources
	@echo "Demo: Resources"
	@echo "Starting server in background..."
	@./bin/zenmcp-server -addr=:8083 & \
	SERVER_PID=$$!; \
	sleep 2; \
	echo "Listing resources..."; \
	./bin/zenmcp-client -addr=localhost:8083 -cmd=list-resources; \
	echo "Reading greeting resource..."; \
	./bin/zenmcp-client -addr=localhost:8083 -cmd=read-resource -resource=test://greeting; \
	echo "Stopping server..."; \
	kill $$SERVER_PID || true

# All-in-one targets
mvp: build test ## Build and test MVP
	@echo "MVP build and test complete!"

ci: deps lint test race integration ## Full CI pipeline
	@echo "CI pipeline complete!"

# Release targets
dist: clean ## Build distribution binaries
	@echo "Building distribution binaries..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build -o dist/zenmcp-server-linux-amd64 ./cmd/server
	@GOOS=linux GOARCH=amd64 go build -o dist/zenmcp-client-linux-amd64 ./cmd/client
	@GOOS=darwin GOARCH=amd64 go build -o dist/zenmcp-server-darwin-amd64 ./cmd/server  
	@GOOS=darwin GOARCH=amd64 go build -o dist/zenmcp-client-darwin-amd64 ./cmd/client
	@GOOS=darwin GOARCH=arm64 go build -o dist/zenmcp-server-darwin-arm64 ./cmd/server
	@GOOS=darwin GOARCH=arm64 go build -o dist/zenmcp-client-darwin-arm64 ./cmd/client
	@GOOS=windows GOARCH=amd64 go build -o dist/zenmcp-server-windows-amd64.exe ./cmd/server
	@GOOS=windows GOARCH=amd64 go build -o dist/zenmcp-client-windows-amd64.exe ./cmd/client
	@echo "Distribution binaries built in dist/"

# Info targets
version: ## Show version info
	@echo "ZenMCP MVP"
	@echo "Go version: $$(go version)"
	@echo "Git commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Build date: $$(date)"

status: ## Show project status
	@echo "Project Status:"
	@echo "- Protocol: JSON-RPC 2.0 + MCP 2025-03-26"
	@echo "- Transports: HTTP (SSE), stdio"
	@echo "- Features: Tools, Resources, Prompts"
	@echo "- Go modules: $$(go list -m)"
	@echo "- Test coverage: $$(go test -cover ./... 2>/dev/null | grep -o '[0-9]*\.[0-9]*%' | tail -1 || echo 'unknown')"

# Quick start target
quickstart: build ## Quick start guide
	@echo ""
	@echo "🚀 ZenMCP MVP Quick Start"
	@echo "========================"
	@echo ""
	@echo "1. Start the demo server:"
	@echo "   make dev-server"
	@echo ""
	@echo "2. In another terminal, test with client:"
	@echo "   ./bin/zenmcp-client -cmd=list-tools"
	@echo "   ./bin/zenmcp-client -cmd=call-tool -tool=echo -args='{\"message\":\"Hello\"}'"
	@echo ""
	@echo "3. Or run quick demos:"
	@echo "   make demo-echo"
	@echo "   make demo-math"
	@echo "   make demo-resources"
	@echo ""
	@echo "4. Run tests:"
	@echo "   make test"
	@echo "   make integration"
	@echo ""

# Create bin directory
bin:
	@mkdir -p bin