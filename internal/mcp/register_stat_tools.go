package mcp

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"

    "github.com/tuannvm/haproxy-mcp-server/internal/haproxy"
)

func registerStatTools(s *server.MCPServer, client *haproxy.HAProxyClient) {
    slog.Info("Registering HAProxy statistics & process info tools...")

    // show_stat tool
    showStat := mcp.NewTool("show_stat",
        mcp.WithDescription("Shows HAProxy statistics table (show stat command)"),
        mcp.WithString("filter", mcp.Description("Optional filter for proxy or server names")),
    )
    s.AddTool(showStat, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        filter := getString(req, "filter")
        slog.InfoContext(ctx, "Executing show_stat", "filter", filter)
        return callJSON(ctx, "get statistics", "stats", func() (interface{}, error) {
            return client.ShowStat(filter)
        })
    })

    // show_info tool
    showInfo := mcp.NewTool("show_info",
        mcp.WithDescription("Shows HAProxy runtime information (version, uptime, limits, mode)"),
    )
    s.AddTool(showInfo, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        slog.InfoContext(ctx, "Executing show_info")
        return callJSON(ctx, "get runtime info", "info", func() (interface{}, error) {
            return client.GetRuntimeInfo()
        })
    })

    // debug_counters tool
    debugCounters := mcp.NewTool("debug_counters",
        mcp.WithDescription("Shows HAProxy internal counters (allocations, events)"),
    )
    s.AddTool(debugCounters, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        slog.InfoContext(ctx, "Executing debug_counters")
        return callJSON(ctx, "get debug counters", "counters", func() (interface{}, error) {
            return client.DebugCounters()
        })
    })

    // clear_counters_all tool
    clearAll := mcp.NewTool("clear_counters_all",
        mcp.WithDescription("Reset all HAProxy statistics counters"),
    )
    s.AddTool(clearAll, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        slog.InfoContext(ctx, "Executing clear_counters_all")
        return callExec(ctx, "clear counters", func() (string, error) {
            if err := client.ClearCountersAll(); err != nil {
                return "", err
            }
            return "All statistics counters have been reset successfully", nil
        })
    })

    // dump_stats_file tool
    dumpStats := mcp.NewTool("dump_stats_file",
        mcp.WithDescription("Dump HAProxy stats to a file"),
        mcp.WithString("filepath", mcp.Required(), mcp.Description("Path where stats file should be saved")),
    )
    s.AddTool(dumpStats, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        path := getString(req, "filepath")
        slog.InfoContext(ctx, "Executing dump_stats_file", "filepath", path)
        return callExec(ctx, "dump stats to file", func() (string, error) {
            out, err := client.DumpStatsFile(path)
            if err != nil {
                return "", err
            }
            return fmt.Sprintf("Statistics dumped successfully to %s", out), nil
        })
    })

    slog.Info("Statistic & process info tools registered")
}