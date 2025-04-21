In HAProxy Runtime API, the `level` parameter controls the access permissions granted to clients connecting to the stats socket. There are three main access levels available:

| Level    | Description                                                                                   | Permissions                                                                                  |
|----------|-----------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------|
| **admin**    | Full administrative access to all commands and operations.                                  | Can execute all commands, including configuration changes, enabling/disabling servers, clearing tables, etc. |
| **operator** | Intermediate access with limited control.                                                  | Can perform standard operational commands like resetting counters, viewing stats, but cannot make configuration changes. |
| **user**     | Read-only access.                                                                           | Can only view information and query status, no ability to change configuration or state.     |

### Summary of Access Levels

- **admin**: Full control, including modifying configuration at runtime.
- **operator**: Operational commands allowed, but no configuration changes.
- **user**: View-only access, no changes allowed.

This allows you to restrict what different users or processes can do when connecting to the HAProxy runtime socket, improving security and control over your HAProxy instance.
