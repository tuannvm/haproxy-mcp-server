# HAProxy MCP Server Tools

This document describes the MCP tools supported by the HAProxy MCP Server, which allow AI assistants to interact with HAProxy's Runtime API through the Model Context Protocol (MCP).

Each tool maps directly to HAProxy Runtime API commands and is implemented using the `client-native` library's Runtime client.

## 1. Statistics & Process Info

### show_stat
Retrieves the full statistics table for HAProxy.
- **Runtime API**: `show stat`
- **Input**: Optional filter (proxy or server names)
- **Output**: Full stats table including bytes, sessions, errors

### show_info
Displays HAProxy version, uptime, and process information.
- **Runtime API**: `show info`
- **Input**: None
- **Output**: Version, uptime, process limits, mode

### debug_counters
Shows internal HAProxy counters.
- **Runtime API**: `debug counters`
- **Input**: None
- **Output**: Internal counters (allocations, events)

### clear_counters_all
Resets all statistics counters.
- **Runtime API**: `clear counters all`
- **Input**: None
- **Output**: Confirmation that all stats have been reset

### dump_stats_file
Writes current statistics to a file.
- **Runtime API**: `dump stats-file`
- **Input**: Filepath
- **Output**: Confirmation + path of dump file

## 2. Topology Discovery

### show_frontend
Lists all frontends with their configurations.
- **Runtime API**: `show frontend`
- **Input**: None
- **Output**: List of frontends with bind address, mode, and state

### show_backend
Lists all backends and their configurations.
- **Runtime API**: `show backend`
- **Input**: None
- **Output**: List of backends and their configuration snippets

### show_servers_state
Displays per-server state and statistics.
- **Runtime API**: `show servers state`
- **Input**: Optional backend name
- **Output**: Per-server state, current sessions, weight

### show_map
Shows entries in a map file.
- **Runtime API**: `show map`
- **Input**: Map filename
- **Output**: All key→value entries

### show_table
Shows entries in a stick table.
- **Runtime API**: `show table`
- **Input**: Table name
- **Output**: Stick‑table entries

## 3. Dynamic Pool Management

### add_server
Dynamically registers a new server in a backend.
- **Runtime API**: `add server <backend> <name> <addr>`
- **Input**: Backend, server name, address, optional port, weight
- **Output**: Success/failure confirmation

### del_server
Removes a dynamic server from a backend.
- **Runtime API**: `del server <backend> <name>`
- **Input**: Backend, server name
- **Output**: Success/failure confirmation

### enable_server
Takes a server out of maintenance mode.
- **Runtime API**: `enable server <backend>/<name>`
- **Input**: Backend, server name
- **Output**: Confirmation

### disable_server
Puts a server into maintenance mode.
- **Runtime API**: `disable server <backend>/<name>`
- **Input**: Backend, server name
- **Output**: Confirmation

### set_weight
Changes a server's load-balancing weight.
- **Runtime API**: `set weight <backend>/<server> <weight>`
- **Input**: Backend, server, new weight
- **Output**: Old vs. new weight

### set_maxconn_server
Sets the maximum number of connections for a server.
- **Runtime API**: `set maxconn server <backend>/<name> <maxconn>`
- **Input**: Backend, server, maxconn value
- **Output**: Confirmation

### set_maxconn_frontend
Sets the maximum number of connections for a frontend.
- **Runtime API**: `set maxconn frontend <frontend> <maxconn>`
- **Input**: Frontend, maxconn value
- **Output**: Confirmation

## 4. Session Control

### show_sess
Lists all active sessions.
- **Runtime API**: `show sess`
- **Input**: None or backend filter
- **Output**: List of active sessions

### shutdown_session
Terminates a specific client session by ID.
- **Runtime API**: `shutdown session <session‑id>`
- **Input**: Session ID
- **Output**: Confirmation

### shutdown_sessions_server
Terminates all sessions on a given server.
- **Runtime API**: `shutdown sessions server <backend>/<name>`
- **Input**: Backend, server
- **Output**: Confirmation

## 5. Maps & ACLs

### add_map
Adds an entry to a map file.
- **Runtime API**: `add map <file> <key> <value>`
- **Input**: Map file, key, value
- **Output**: Confirmation

### del_map
Deletes a single entry from a map file.
- **Runtime API**: `del map <file> <key>`
- **Input**: Map file, key
- **Output**: Confirmation

### set_map
Updates the value of an existing map entry.
- **Runtime API**: `set map <file> <key> <value>`
- **Input**: Map file, key, new value
- **Output**: Confirmation

### clear_map
Deletes all entries from a map file.
- **Runtime API**: `clear map <file>`
- **Input**: Map file
- **Output**: Confirmation

### commit_map
Commits a prepared map‐file transaction.
- **Runtime API**: `commit map <file>`
- **Input**: Map file
- **Output**: Confirmation

### add_acl
Adds a value to an ACL list.
- **Runtime API**: `add acl <file> <key>`
- **Input**: ACL file, key
- **Output**: Confirmation

### del_acl
Removes a value from an ACL list.
- **Runtime API**: `del acl <file> <key>`
- **Input**: ACL file, key
- **Output**: Confirmation

### clear_acl
Deletes all entries from an ACL list.
- **Runtime API**: `clear acl <file>`
- **Input**: ACL file
- **Output**: Confirmation

### commit_acl
Commits a prepared ACL transaction.
- **Runtime API**: `commit acl <file>`
- **Input**: ACL file
- **Output**: Confirmation

## 6. Health Checks & Agents

### enable_health
Enables active health checks on a server.
- **Runtime API**: `enable health <backend>/<server>`
- **Input**: Backend, server
- **Output**: Confirmation

### disable_health
Disables active health checks on a server.
- **Runtime API**: `disable health <backend>/<server>`
- **Input**: Backend, server
- **Output**: Confirmation

### enable_agent
Resumes agent-based health probes.
- **Runtime API**: `enable agent <backend>/<server>`
- **Input**: Backend, server
- **Output**: Confirmation

### disable_agent
Stops agent-based health probes.
- **Runtime API**: `disable agent <backend>/<server>`
- **Input**: Backend, server
- **Output**: Confirmation

## 7. Miscellaneous

### show_errors
Lists protocol violation errors.
- **Runtime API**: `show errors [since <seconds>]`
- **Input**: Optional time filter in seconds
- **Output**: Protocol violation errors

### echo
Returns a string (connectivity test).
- **Runtime API**: `echo <string>`
- **Input**: String to echo
- **Output**: Echoed string

### help
Shows available Runtime API commands.
- **Runtime API**: `help`
- **Input**: None
- **Output**: List of all Runtime API commands
