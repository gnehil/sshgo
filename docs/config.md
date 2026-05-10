# Configuration Format

sshgo's configuration file is located at `~/.sshgo/config.yaml` in YAML format.

## Directory Structure

```yaml
profiles:    # Connection profile list
  - ...

groups:      # Group definitions
  - ...
```

## Connection Profile

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Connection alias. Only lowercase letters, numbers, hyphens allowed; cannot start with hyphen |
| `host` | string | Yes | - | Server IP or hostname |
| `port` | int | No | 22 | SSH port (1-65535) |
| `user` | string | Yes | - | SSH username |
| `group` | string | No | - | Group name (must be pre-defined in groups) |
| `identity_file` | string | No | - | SSH private key path (supports `~` expansion) |
| `jump_hosts` | array | No | - | Jump host list (see below) |
| `forward_ports` | array | No | - | Port forwarding list (see below) |
| `keepalive_interval` | int | No | - | ServerAliveInterval in seconds |
| `server_alive_count` | int | No | - | ServerAliveCountMax |

### Example

```yaml
profiles:
  - name: "web-prod"
    host: "10.0.0.1"
    port: 2222
    user: "deploy"
    group: "prod"
    identity_file: "~/.ssh/prod_key"
    keepalive_interval: 30
    server_alive_count: 3
```

## Jump Host

Each jump host is a nested object:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | No | - | Jump host display name |
| `host` | string | Yes | - | Jump host address |
| `port` | int | No | 22 | Jump host port |
| `user` | string | Yes | - | Jump host username |
| `identity_file` | string | No | - | Jump host key path |

### Multi-hop Example

```yaml
profiles:
  - name: "internal-db"
    host: "10.0.0.50"
    user: "dba"
    jump_hosts:
      - name: "bastion"
        host: "jump.company.com"
        user: "bastion-user"
        port: 22
      - name: "gateway"
        host: "10.0.1.1"
        user: "gateway-user"
```

The actual SSH command executed is equivalent to:

```bash
ssh -i ~/.ssh/prod_key -J bastion-user@jump.company.com,gateway-user@10.0.1.1 dba@10.0.0.50
```

## Port Forwarding

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `local_port` | int | Yes | Local listening port |
| `remote_host` | string | Yes | Remote target host |
| `remote_port` | int | Yes | Remote target port |

### Example

```yaml
profiles:
  - name: "db-server"
    host: "10.0.0.5"
    user: "dba"
    forward_ports:
      - local_port: 5432
        remote_host: "localhost"
        remote_port: 5432
      - local_port: 8080
        remote_host: "web-app"
        remote_port: 80
```

## Group

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Group name (unique identifier) |
| `description` | string | No | Group description |

### Example

```yaml
groups:
  - name: "prod"
    description: "Production servers"
  - name: "staging"
    description: "Staging servers"
  - name: "dev"
    description: "Development servers"
```

**Note**: Profile's `group` field must reference an already-defined group, otherwise validation will fail.

## Connection History

Connection history is stored in `~/.sshgo/history.json`:

```json
[
  {
    "name": "web-prod",
    "last_connected": "2025-01-15T10:30:00Z",
    "connect_count": 42
  },
  {
    "name": "db-server",
    "last_connected": "2025-01-14T09:15:00Z",
    "connect_count": 7
  }
]
```

Sorted by last connection time in descending order (most recent first).

## Security Notes

- **Passwords are not stored in the config file**, only in OS native keychain (macOS Keychain / Windows Credential Manager / Linux Secret Service)
- Config file permissions are `0600` (owner read/write only)
- Automatic backup created before each config modification (`config.yaml.bak.<timestamp>`)
- `~` in paths is automatically expanded to user home directory