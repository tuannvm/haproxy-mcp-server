package mcp

import (
    "context"
    "encoding/json"
    "fmt"
    "log/slog"

    "github.com/mark3labs/mcp-go/mcp"
)

// callJSON handles executing a client call, error logging, and JSON marshalling
func callJSON(ctx context.Context, action, mapKey string, fn func() (interface{}, error)) (*mcp.CallToolResult, error) {
    v, err := fn()
    if err != nil {
        slog.ErrorContext(ctx, "Failed to "+action, "error", err)
        return mcp.NewToolResultError(fmt.Sprintf("Failed to %s: %v", action, err)), nil
    }
    out, err := json.Marshal(map[string]interface{}{mapKey: v})
    if err != nil {
        slog.ErrorContext(ctx, "Failed to marshal "+mapKey+" output", "error", err)
        return mcp.NewToolResultError("Internal server error: failed to marshal results"), nil
    }
    return mcp.NewToolResultText(string(out)), nil
}

// callExec handles executing a client action, error logging, and returning plain text
func callExec(ctx context.Context, action string, fn func() (string, error)) (*mcp.CallToolResult, error) {
    s, err := fn()
    if err != nil {
        slog.ErrorContext(ctx, "Failed to "+action, "error", err)
        return mcp.NewToolResultError(fmt.Sprintf("Failed to %s: %v", action, err)), nil
    }
    return mcp.NewToolResultText(s), nil
}

// getString extracts a string argument from the request
func getString(req mcp.CallToolRequest, key string) string {
    if v, ok := req.Params.Arguments[key].(string); ok {
        return v
    }
    return ""
}

// getInt extracts a numeric argument (as int) from the request
func getInt(req mcp.CallToolRequest, key string) int {
    if f, ok := req.Params.Arguments[key].(float64); ok {
        return int(f)
    }
    return 0
}