### Supported HAProxy Runtime API Methods

Below is the full list of Runtime API commands our MCP tools will invoke, grouped by feature area. Each corresponds exactly to a supported method in HAProxy’s Runtime API.

#### 1. Statistics & Process Info
- show stat — retrieve the full statistics table (bytes, sessions, errors, etc.)
- show info — display version, uptime, limits, process mode
- debug counters — show internal HAProxy counters (allocations, events)
- clear counters all — reset all statistics counters to zero
- dump stats-file — write current statistics to a file for post‑reload population

#### 2. Topology Discovery
- show frontend — list all frontends with bind addresses, modes, and states
- show backend — list all backends and their configuration snippets
- show servers state — display per‑server state, current sessions, weights; filterable by backend
- show map — list entries in a specified HAProxy map file
- show table — list entries in a specified stick table

#### 3. Dynamic Server Management
- add server — dynamically register a new server in a backend
- del server — remove a dynamic server from a backend
- enable server — take a server out of maintenance mode (`enable server <backend>/<name>`)
- disable server — put a server into maintenance mode (`disable server <backend>/<name>`)
- set weight — change a server’s load‑balancing weight (`set weight <backend>/<server> <value>`)
- set maxconn server — set max connections per server (`set maxconn server <backend>/<name> <value>`)
- set maxconn frontend — set max connections per frontend (`set maxconn frontend <name> <value>`)

#### 4. Session Control
- show sess — list all active sessions; optional backend filter
- shutdown session — terminate a specific client session by ID
- shutdown sessions server — terminate all sessions on a given server (`shutdown sessions server <backend>/<name>`)

#### 5. Maps & ACLs
- add map — add an entry to a map file
- del map — delete a single entry from a map file
- set map — update the value of an existing map entry
- clear map — delete all entries from a map file
- commit map — commit a prepared map‐file transaction
- add acl — add a value to an ACL list
- del acl — remove a value from an ACL list
- clear acl — delete all entries from an ACL list
- commit acl — commit a prepared ACL transaction

#### 6. SSL Management
- show ssl cert — list loaded SSL certificates
- dump ssl cert — output a certificate in PEM format (`dump ssl cert <file> <index>`)
- add ssl ca-file — begin a transaction to add certificates to a CA file
- del ssl ca-file — delete a CA file from memory
- commit ssl ca-file — commit the CA‑file update transaction

#### 7. Health Checks & Agents
- enable health — enable active health checks on a server (`enable health <backend>/<name>`)
- disable health — disable active health checks on a server (`disable health <backend>/<name>`)
- enable agent — resume agent‑based health probes (`enable agent <backend>/<name>`)
- disable agent — stop agent‑based health probes (`disable agent <backend>/<name>`)

#### 8. Miscellaneous
- show errors — list protocol‑violation errors logged since startup or since a given time
- echo — return a string (connectivity and sequencing test)
- help — list all available Runtime API commands
