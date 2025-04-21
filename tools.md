Below is the revised project plan, with **Step 3** (“Define MCP Tools”) updated to include only commands actually supported by the HAProxy Runtime API. All other steps remain as before.

## Revised Plan: Create HAProxy MCP Server

### Step 1: Project Initialization and Structure  
_(unchanged — see previous plan)_

### Step 2: HAProxy Client Wrapper  
_(unchanged — see previous plan)_

### Step 3 (Revised): Define MCP Tools Based on Runtime API  
For each tool below, the implementation will invoke the corresponding Runtime API command over the stats socket via `client-native`’s Runtime client.

#### 3.1 Statistics & Process Info  
- **show_stat**  
  • Runtime API: `show stat`  
  • Input: optional filter (proxy or server names)  
  • Output: full stats table (in JSON) including bytes, sessions, errors  
- **show_info**  
  • Runtime API: `show info`  
  • Input: none  
  • Output: version, uptime, process limits, mode  
- **debug_counters**  
  • Runtime API: `debug counters`  
  • Input: none  
  • Output: internal counters (allocations, events)  
- **clear_counters_all**  
  • Runtime API: `clear counters all`  
  • Input: none  
  • Output: confirmation that all stats have been reset  
- **dump_stats_file**  
  • Runtime API: `dump stats-file`  
  • Input: filepath  
  • Output: confirmation + path of dump file  

#### 3.2 Topology Discovery  
- **show_frontend**  
  • Runtime API: `show frontend`  
  • Input: none  
  • Output: list of frontends with bind address, mode, and state  
- **show_backend**  
  • Runtime API: `show backend`  
  • Input: none  
  • Output: list of backends and their configuration snippets  
- **show_servers_state**  
  • Runtime API: `show servers state`  
  • Input: optional backend name  
  • Output: per-server state, current sessions, weight  
- **show_map**  
  • Runtime API: `show map`  
  • Input: map filename  
  • Output: all key→value entries  
- **show_table**  
  • Runtime API: `show table`  
  • Input: table name  
  • Output: stick‑table entries  

#### 3.3 Dynamic Pool Management  
- **add_server**  
  • Runtime API: `add server <backend> <name> <addr>`  
  • Input: backend, server name, address, optional port, weight  
  • Output: success/failure  
- **del_server**  
  • Runtime API: `del server <backend> <name>`  
  • Input: backend, server name  
  • Output: success/failure  
- **enable_server** / **disable_server**  
  • Runtime API: `enable server <backend>/<name>` / `disable server <backend>/<name>`  
  • Input: backend, server name  
  • Output: confirmation  
- **set_weight**  
  • Runtime API: `set weight <backend>/<server> <weight>`  
  • Input: backend, server, new weight  
  • Output: old vs. new weight  
- **set_maxconn_server**  
  • Runtime API: `set maxconn server <backend>/<name> <maxconn>`  
  • Input: backend, server, maxconn value  
  • Output: confirmation  
- **set_maxconn_frontend**  
  • Runtime API: `set maxconn frontend <frontend> <maxconn>`  
  • Input: frontend, maxconn value  
  • Output: confirmation  

#### 3.4 Session Control  
- **show_sess**  
  • Runtime API: `show sess`  
  • Input: none or backend filter  
  • Output: list of active sessions  
- **shutdown_session**  
  • Runtime API: `shutdown session <session‑id>`  
  • Input: session ID  
  • Output: confirmation  
- **shutdown_sessions_server**  
  • Runtime API: `shutdown sessions server <backend>/<name>`  
  • Input: backend, server  
  • Output: confirmation  

#### 3.5 Maps & ACLs  
- **add_map** / **del_map**  
  • Runtime API: `add map <file> <key> <value>` / `del map <file> <key>`  
  • Input: map file, key, optional value  
  • Output: confirmation  
- **set_map**  
  • Runtime API: `set map <file> <key> <value>`  
  • Input: map file, key, new value  
  • Output: confirmation  
- **clear_map** / **commit_map**  
  • Runtime API: `clear map <file>` / `commit map <file>`  
  • Input: map file  
  • Output: confirmation  
- **add_acl**, **del_acl**, **clear_acl**, **commit_acl**  
  • Runtime API: analogous commands for ACL files  

#### 3.6 Health Checks & Agents  
- **enable_health** / **disable_health**  
  • Runtime API: `enable health <backend>/<server>` / `disable health <backend>/<server>`  
  • Input: backend, server  
  • Output: confirmation  
- **enable_agent** / **disable_agent**  
  • Runtime API: `enable agent <backend>/<server>` / `disable agent <backend>/<server>`  
  • Input: backend, server  
  • Output: confirmation  

#### 3.8 Miscellaneous  
- **show_errors**  
  • Runtime API: `show errors [since <seconds>]`  
  • Output: protocol violation errors  
- **echo**  
  • Runtime API: `echo <string>`  
  • Output: echoed string (useful for connectivity tests)  
- **help**  
  • Runtime API: `help`  
  • Output: list of all Runtime API commands  

> Each of these tools should be implemented in `internal/mcp/` with clearly defined input/output structs, registered in the MCP server, and exercised via `client-native`’s Runtime client under the hood.  

### Steps 4–10  
_Remain unchanged — see previous plan._

This revised **Step 3** ensures our MCP server surface matches exactly what HAProxy’s Runtime API provides, avoiding unsupported or hypothetical operations.
