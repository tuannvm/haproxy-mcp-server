Okay, here is a plan outlining the steps to create the `haproxy-mcp-server` Go project, leveraging the `haproxytech/client-native` SDK and mirroring the structure of your `mcp-trino` and `kafka-mcp-server` projects.

This plan is designed to be followed sequentially by a coding agent like Cursor.

## Plan: Create HAProxy MCP Server

**Goal:** Develop a Go-based MCP server (`haproxy-mcp-server`) that allows interaction with HAProxy's runtime API using the `haproxytech/client-native` library, following the patterns established in `tuannvm/mcp-trino` and `tuannvm/kafka-mcp-server`.

**Repository:** `tuannvm/haproxy-mcp-server` (already created)
**Branch:** `main`

**Important Notes:**
1. This MCP server will strictly interact with HAProxy through the Runtime API only, using TCP4 connections.
2. The project will exclusively use Go's standard library `log/slog` for logging throughout the codebase, not third-party logging libraries.

---

### Step 1: Project Initialization and Basic Structure

1.  **Clone the empty repository:**
    ```bash
    git clone git@github.com:tuannvm/haproxy-mcp-server.git
    cd haproxy-mcp-server
    ```
2.  **Initialize Go module:**
    ```bash
    go mod init github.com/tuannvm/haproxy-mcp-server
    ```
3.  **Add Core Dependencies:**
    ```bash
    go get github.com/haproxytech/client-native/v6@latest # Use v6 based on example
    go get github.com/mark3labs/mcp-go@latest
    go get github.com/spf13/viper@latest # For configuration
    ```
4.  **Create Basic Directory Structure:**
    ```bash
    mkdir -p cmd/server internal/haproxy internal/mcp/tools internal/config
    ```
5.  **Create Initial `main.go`:**
    *   Create `cmd/server/main.go`.
    *   Add a basic `main` function that prints a startup message (e.g., "Starting HAProxy MCP Server...").
    *   Use Go's standard library `log/slog` for logging.
6.  **Create `.gitignore`:**
    *   Add a standard Go `.gitignore` file (e.g., covering binaries, `.idea`, `.vscode`, `*.env`).

### Step 2: Implement HAProxy Client Wrapper

**Important Note: This MCP server will strictly interact with HAProxy through the Runtime API only.**

1.  **Create `internal/haproxy/client.go`:**
    *   Define a struct (e.g., `HAProxyClient`) to hold the initialized `client_native.Client`.
    *   Implement a `NewHAProxyClient` function.
        *   This function should take configuration parameters for connecting to the HAProxy Runtime API over TCP4.
        *   For TCP4 connections, rather than using a Unix socket path, we'll use the IP address and port of the HAProxy instance.
        *   Use the example code from the `haproxytech/client-native` README to initialize the `runtime` client with TCP4 connectivity.
        *   Configure the Socket option to use a TCP4 connection string (e.g., "tcp4:192.168.1.10:9999").
        *   Combine necessary components using `client_native.New` with appropriate options.
        *   Handle potential initialization errors and connection issues.
        *   Return an instance of the `HAProxyClient` struct or an error.
    *   Add methods to the `HAProxyClient` struct that wrap the core `client-native` runtime functionalities needed for the MCP tools (e.g., `GetBackends`, `GetRuntimeInfo`, `EnableServer`, `DisableServer`, `ReloadHAProxy`). These methods will be called by the tool implementations later.

### Step 3: Define MCP Tools

1.  **Identify Core HAProxy Operations:** Based on `client-native` capabilities, select initial operations to expose as tools. Good candidates include:
    *   `list_backends`: List configured backends.
    *   `get_backend`: Get details of a specific backend.
    *   `list_servers`: List servers within a backend.
    *   `get_server`: Get details of a specific server.
    *   `enable_server`: Enable a server in a backend.
    *   `disable_server`: Disable a server in a backend.
    *   `show_info`: Get runtime information (like `show info`).
    *   `reload_haproxy`: Trigger a configuration reload.
    *   `get_stats`: Get runtime statistics (like `show stat`).
2.  **Create Tool Definition Files:** In `internal/mcp/tools/`:
    *   For each tool (e.g., `list_backends.go`, `enable_server.go`), define:
        *   Input struct (e.g., `EnableServerInput`) with JSON tags.
        *   Output struct (e.g., `EnableServerOutput` or a simple success message struct) with JSON tags.
        *   An implementation function (e.g., `EnableServerFunc`) that:
            *   Takes the `HAProxyClient` (from Step 2) and the tool input as arguments.
            *   Calls the appropriate wrapper method on the `HAProxyClient`.
            *   Processes the result and returns the output struct or an error.
3.  **Create Tool Registration Logic:**
    *   In `internal/mcp/tools/register.go` (or similar), create a function (e.g., `RegisterTools`) that takes an `mcp.Server` and the `HAProxyClient` as input.
    *   Inside this function, for each defined tool:
        *   Create an `mcp.Tool` definition, specifying the `Name`, `Description`, `InputSchema` (derived from the input struct), and `OutputSchema` (derived from the output struct).
        *   Use `mcp.RegisterTool` to register the tool, providing the implementation function created above.

### Step 4: Implement MCP Server Logic

1.  **Update `cmd/server/main.go`:**
    *   Initialize configuration (Step 5).
    *   Initialize logging using Go's standard library `log/slog`.
    *   Initialize the `HAProxyClient` (Step 2) using the configuration.
    *   Initialize the `mcp.Server` using `mcp.NewServer`.
    *   Register the tools using the function from Step 3 (`RegisterTools`).
    *   Implement transport handling based on configuration (`MCP_TRANSPORT` env var):
        *   If `stdio`, use `server.Run()` or `server.RunWithContext()`.
        *   If `http`, configure and run an HTTP server (using `net/http`) with the SSE handler (`mcp.HTTPHandlerSSE`) listening on the configured port (`MCP_PORT`). Mirror the logic in `mcp-trino`.
    *   Include proper error handling and shutdown logic.

### Step 5: Configuration Management

1.  **Create `internal/config/config.go`:**
    *   Define a `Config` struct holding all necessary configuration parameters:
        *   `HAProxyHost` - Host of the HAProxy instance
        *   `HAProxyPort` - Port for the HAProxy Runtime API
        *   `HAProxyRuntimeMode` - "tcp4" or "unix"
        *   `HAProxyRuntimeSocket` - Used only when HAProxyRuntimeMode is "unix"
        *   `MCPTransport` (stdio/http)
        *   `MCPPort` (for http)
        *   `LogLevel`
    *   Implement a `LoadConfig` function:
        *   Use `viper` to read configuration from environment variables.
        *   Set sensible defaults (e.g., "127.0.0.1" for host, 9999 for port, "tcp4" for mode, "stdio" for transport).
        *   Define environment variable names (e.g., `HAPROXY_HOST`, `HAPROXY_PORT`, `HAPROXY_RUNTIME_MODE`, `MCP_TRANSPORT`).
        *   Return the populated `Config` struct or an error.

### Step 6: Build and Packaging (Makefile & Dockerfile)

1.  **Create `Makefile`:**
    *   Mimic the `Makefile` from `mcp-trino` or `kafka-mcp-server`.
    *   Include targets: `build`, `test`, `lint`, `clean`, `docker-build`.
    *   Ensure `build` creates the binary in a `./bin/` directory.
    *   Configure `golangci-lint` (add `.golangci.yml` similar to reference projects).
2.  **Create `Dockerfile`:**
    *   Mimic the `Dockerfile` from `mcp-trino` or `kafka-mcp-server`.
    *   Use a multi-stage build.
    *   Copy the binary into a minimal base image (like `gcr.io/distroless/static-debian11` or `alpine`).
    *   Set the appropriate entrypoint.
    *   Declare expected environment variables using `ENV` (for documentation purposes, they will be overridden at runtime).

### Step 7: Testing

1.  **Write Unit Tests:**
    *   For functions in `internal/haproxy`, `internal/mcp/tools`, and `internal/config`.
    *   Mock the `client-native` interfaces where necessary to test tool logic without a live HAProxy.
    *   Place tests alongside the code (e.g., `client_test.go`).
2.  **(Optional) Integration Tests:**
    *   If desired, create integration tests that require a running HAProxy instance (perhaps managed via Docker Compose).
    *   Use build tags (e.g., `//go:build integration`) or `-short` flag conventions to separate these from unit tests.

### Step 8: Documentation (README.md)

1.  **Create `README.md`:**
    *   Follow the structure and style of `mcp-trino` and `kafka-mcp-server`.
    *   **Title:** HAProxy MCP Server in Go.
    *   **Badges:** Add relevant badges (CI, Go version, License, Go Report Card, Docker Image, Release).
    *   **Overview:** Briefly describe the project and its purpose (interacting with HAProxy via MCP).
    *   **Features:** List key features (MCP server, HAProxy config/runtime interaction, Docker, transport modes, compatibility).
    *   **Installation:** Provide instructions for Homebrew (requires creating a tap later), manual download, and building from source.
    *   **Downloads:** Add a table for binary downloads (placeholders for release assets).
    *   **MCP Integration:**
        *   Provide JSON examples for Cursor, Claude Desktop, Windsurf, ChatWise, and Docker.
        *   **Crucially, update the `env` section in these examples to use the HAProxy-specific environment variables defined in Step 5** (e.g., `HAPROXY_CONFIG_FILE`, `HAPROXY_RUNTIME_SOCKET`).
    *   **Available MCP Tools:**
        *   For *each tool* defined in Step 3:
            *   Add a heading (e.g., `### list_backends`).
            *   Provide a clear description.
            *   Include a sample prompt.
            *   Show example input JSON.
            *   Show example output JSON.
    *   **(Optional) MCP Resources:** If implementing resource endpoints (like `kafka-mcp://`), document them here.
    *   **(Optional) MCP Prompts:** If implementing prompt endpoints, document them here.
    *   **Configuration:** Add a table listing all environment variables (from Step 5), their descriptions, and default values.
    *   **Contributing:** Add a brief section on contributing.
    *   **License:** State the license (e.g., MIT).
    *   **(Optional) CI/CD:** Briefly describe the CI/CD process.

### Step 9: CI/CD Setup (GitHub Actions)

1.  **Create `.github/workflows/build.yml`:**
    *   Copy and adapt the workflow from `mcp-trino` or `kafka-mcp-server`.
    *   Ensure it includes steps for:
        *   Checkout code.
        *   Setup Go.
        *   Run linters (`golangci-lint`).
        *   Run tests (`go test ./...`).
        *   Run vulnerability checks (`govulncheck`, `trivy`).
        *   Build the binary.
        *   (Optional) Build Docker image.
2.  **(Optional) Create Release Workflow:**
    *   Create a separate workflow (e.g., `release.yml`) triggered on tag creation.
    *   Use `goreleaser` (add `.goreleaser.yml` config) to:
        *   Build binaries for multiple platforms.
        *   Create a GitHub Release.
        *   Generate SBOM/provenance attestations.
        *   Build and push Docker images to GHCR.
        *   (Optional) Update the Homebrew tap.
