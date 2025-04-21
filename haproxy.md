# HAProxy Configuration Guide

This guide provides detailed instructions for configuring HAProxy to expose both the Runtime API and the Statistics page, which are required for the HAProxy MCP Server to function properly.

## Table of Contents
- [Runtime API Configuration](#runtime-api-configuration)
  - [TCP Socket Mode](#tcp-socket-mode)
  - [Unix Socket Mode](#unix-socket-mode)
- [Statistics Page Configuration](#statistics-page-configuration)
- [Security Considerations](#security-considerations)
- [Combined Configuration Example](#combined-configuration-example)
- [Troubleshooting](#troubleshooting)

## Runtime API Configuration

HAProxy's Runtime API allows for dynamic configuration changes and monitoring without restarting the service. The HAProxy MCP Server requires access to this API to function properly.

### TCP Socket Mode

To expose the Runtime API over a TCP socket, add the following to your `haproxy.cfg`:

```
global
    # Other global settings...
    
    # Runtime API configuration
    stats socket ipv4@0.0.0.0:9999 level admin
    # OR for more secure setup, bind to localhost only
    # stats socket ipv4@127.0.0.1:9999 level admin
```

For HAProxy 2.0 and later, you can also use:

```
global
    # Other global settings...
    
    # Runtime API with HTTP wrapper
    stats socket ipv4@0.0.0.0:9999 level admin expose-fd listeners
    # Enable prometheus-exporter on the stats socket
    stats socket ipv4@0.0.0.0:9999 level admin expose-fd listeners
```

### Unix Socket Mode

For Unix socket mode, which provides better security as it's file-system based:

```
global
    # Other global settings...
    
    # Runtime API configuration using Unix socket
    stats socket /var/run/haproxy/admin.sock mode 660 level admin
    stats timeout 30s
```

Ensure that the directory exists and has proper permissions:

```bash
mkdir -p /var/run/haproxy
chown haproxy:haproxy /var/run/haproxy
chmod 755 /var/run/haproxy
```

## Statistics Page Configuration

HAProxy's Statistics page provides a web-based dashboard for monitoring. To enable it, add:

```
frontend stats
    bind *:8404
    stats enable
    stats uri /stats
    stats refresh 10s
    stats show-legends
    stats show-node
    # Optional: Enable admin features if needed
    # stats admin if LOCALHOST
```

For a more secure setup, restrict access:

```
frontend stats
    bind 127.0.0.1:8404
    stats enable
    stats uri /stats
    stats refresh 10s
    stats auth admin:YourSecurePassword
    stats hide-version
```

## Security Considerations

When exposing the Runtime API and Statistics page, consider these security practices:

1. **Authentication**: Always use authentication for production environments
    ```
    stats socket ipv4@127.0.0.1:9999 level admin user admin password YourSecurePassword
    ```

2. **Binding**: Bind services to localhost or internal IPs only when possible

3. **Firewall Rules**: Use firewall rules to restrict access to the Runtime API and Stats ports

4. **TLS/SSL**: For statistics page, consider using HTTPS:
    ```
    frontend stats
        bind *:8404 ssl crt /path/to/cert.pem
        stats enable
        stats uri /stats
    ```

5. **Access Control**: Limit who can access admin functions:
    ```
    acl internal_networks src 10.0.0.0/8 192.168.0.0/16
    stats admin if internal_networks
    ```

## Combined Configuration Example

Here's a complete example that includes both Runtime API and Statistics page:

```
global
    log /dev/log local0
    log /dev/log local1 notice
    chroot /var/lib/haproxy
    stats socket /var/run/haproxy/admin.sock mode 660 level admin
    stats socket ipv4@127.0.0.1:9999 level admin
    stats timeout 30s
    user haproxy
    group haproxy
    daemon

defaults
    log     global
    mode    http
    option  httplog
    option  dontlognull
    timeout connect 5000
    timeout client  50000
    timeout server  50000

frontend stats
    bind *:8404
    stats enable
    stats uri /stats
    stats refresh 10s
    stats auth admin:YourSecurePassword
    stats hide-version

# Your other frontend/backend configurations...
```

## Troubleshooting

If you experience issues connecting to the Runtime API or Statistics page:

1. **Check permissions**:
   - For Unix sockets: `ls -la /var/run/haproxy/admin.sock`
   - Ensure the user running the MCP server has access

2. **Verify the socket is listening**:
   - For TCP mode: `netstat -an | grep 9999`
   - For Stats page: `netstat -an | grep 8404`

3. **Test connections directly**:
   ```bash
   # TCP socket
   echo "show info" | socat tcp-connect:127.0.0.1:9999 stdio
   
   # Unix socket
   echo "show info" | socat unix-connect:/var/run/haproxy/admin.sock stdio
   
   # Stats page (should return HTML)
   curl -s http://localhost:8404/stats
   ```

4. **Check HAProxy logs**:
   ```bash
   tail -f /var/log/haproxy.log
   ```

5. **Restart HAProxy after config changes**:
   ```bash
   systemctl restart haproxy
   # or
   service haproxy restart
   ```

For more detailed information, refer to the [official HAProxy documentation](https://www.haproxy.org/download/2.6/doc/management.txt).
