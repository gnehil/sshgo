# Command Reference

## Connection Management

### `sshgo add <name>`

Add a new SSH connection profile.

| Flag | Description | Required |
|------|-------------|----------|
| `<name>` | Connection alias (lowercase letters, numbers, hyphens; cannot start with hyphen) | Yes |
| `--host` | Server IP or hostname | Yes |
| `--port`, `-p` | SSH port (default: 22) | No |
| `--user`, `-u` | SSH username | Yes |
| `--group` | Group name | No |
| `--identity-file`, `-i` | Path to identity (private key) file | No |

```bash
sshgo add my-server --host 192.168.1.10 --user deploy -p 2222 --group prod

# With key-based authentication
sshgo add prod --host 10.0.0.1 --user deploy -i ~/.ssh/id_ed25519_prod
```

### `sshgo show <name>`

Show detailed connection profile information (including jump hosts, port forwarding, etc.).

```bash
sshgo show my-server
```

### `sshgo edit <name>`

Open `~/.sshgo/config.yaml` for manual editing.

```bash
# Edit specific connection (opens full config file)
sshgo edit my-server

# Edit the full configuration file
sshgo edit
```

### `sshgo delete <name>`

Delete the specified connection profile (aliases: `rm`, `del`).

```bash
sshgo delete my-server
sshgo rm my-server
```

---

## Listing

### `sshgo list`

List all connection profiles.

| Flag | Description |
|------|-------------|
| `--group` | Filter by group |
| `--sort` | Sort order: `name` |
| `--format`, `-f` | Output format: `table` (default), `json` |

```bash
sshgo list                  # List all
sshgo list --group prod     # Show only prod group
sshgo list --sort name      # Sort by name
sshgo list --format json    # JSON output
sshgo ls                    # Alias
```

---

## Connecting

### `sshgo connect <name>`

Connect to a saved SSH session. Actually executes `ssh` and replaces the current process.

| Flag | Description |
|------|-------------|
| `<name>` | Connection alias |
| `--recent` | Choose from connection history |

```bash
sshgo connect my-server
sshgo my-server           # Quick mode (equivalent)
sshgo connect --recent    # Choose from history
```

---

## Groups

### `sshgo group list`

List all groups with the number of connections in each.

### `sshgo group add <name>`

Add a new group.

| Flag | Description |
|------|-------------|
| `--description` | Group description |

```bash
sshgo group add prod --description "Production servers"
```

### `sshgo group delete <name>`

Delete a group (does not affect connections in that group).

```bash
sshgo group delete prod
```

---

## Jump Hosts

### `sshgo add-jump <name>`

Add jump host(s) to a connection (supports multi-hop chains).

| Flag | Description |
|------|-------------|
| `<name>` | Target connection alias |
| `--jump` | Jump host address (can be used multiple times) |
| `--identity-file` | Identity file for the Nth `--jump` (position-matched, repeatable) |

Jump host address format: `[user@]host[:port]`, user defaults to root.

```bash
# Single hop
sshgo add-jump db-server --jump bastion

# Multi-hop chain
sshgo add-jump db-server --jump bastion --jump gateway

# With user and port
sshgo add-jump db-server --jump deploy@bastion:2222

# With per-hop identity files (position-matched to --jump)
sshgo add-jump db-server \
  --jump admin@bastion1:22 -i ~/.ssh/id_bastion1 \
  --jump ops@bastion2:22   -i ~/.ssh/id_bastion2
```

---

## Port Forwarding

### `sshgo forward add <name>`

Add port forwarding rule to a connection.

| Flag | Description |
|------|-------------|
| `<name>` | Connection alias |
| `--local`, `-L` | Local port forward (format: `local_port:remote_host:remote_port` or `local:remote`) |

```bash
# Short form (remote_host defaults to localhost)
sshgo forward add my-server -L 8080:80

# Full form
sshgo forward add my-server -L 8080:db-host:5432
```

### `sshgo forward list`

List all port forwarding rules.

```bash
sshgo forward list
```

---

## Batch Execution

### `sshgo exec <pattern> <command>`

Execute commands on multiple servers in parallel.

| Flag | Description |
|------|-------------|
| `<pattern>` | Match pattern: wildcard, comma-separated list |
| `<command>` | Command to execute |
| `--group` | Execute on all servers in specified group |

```bash
# Wildcard matching
sshgo exec 'web-*' "uptime"

# Comma-separated list
sshgo exec "web-1,web-2,db-1" "df -h"

# Execute by group
sshgo exec --group prod "systemctl status nginx"
```

Output format:

```
=== web-1 ===
 10:30:45 up 42 days, 2 users

=== web-2 ===
 10:30:45 up 15 days, 1 users

Total: 2 succeeded, 0 failed
```

**Note**: Batch execution uses SSH default authentication (key, Agent). Servers requiring password authentication may not work interactively.

---

## Import/Sync

### `sshgo import`

Import connections from OpenSSH config file.

| Flag | Description |
|------|-------------|
| `--file`, `-f` | Source file path (default: `~/.ssh/config`) |
| `--overwrite`, `-w` | Overwrite existing profiles with same name |

```bash
sshgo import
sshgo import --file /path/to/custom-ssh-config
sshgo import --overwrite
```

Supported directives: Host, HostName, User, Port, IdentityFile, ProxyJump, Include.

### `sshgo sync`

Export sshgo configuration to OpenSSH compatible format, writing to `~/.ssh/config`.

| Flag | Description |
|------|-------------|
| `--dry-run` | Preview output without writing |

```bash
sshgo sync --dry-run    # Preview
sshgo sync              # Write to ~/.ssh/config
```

**Behavior**:
- Automatically backs up original `~/.ssh/config` to `~/.ssh/config.sshgo.bak`
- Preserves non-sshgo content in `~/.ssh/config`
- Replaces sshgo-generated blocks (marked with `# Generated by sshgo`)

---

## Miscellaneous

### `sshgo ping <name>`

Test SSH connectivity to target server.

```bash
sshgo ping web-server
```

Output: `✓ user@host:port - Connected in 1.234s` or `✗ user@host:port - Connection failed (...)`

Timeout: 5 seconds (BatchMode non-interactive).

### `sshgo history`

View connection history.

```bash
sshgo history
```

### `sshgo completion [bash|zsh|fish]`

Generate shell completion scripts.

```bash
# Bash
source <(sshgo completion bash)

# Zsh
sshgo completion zsh > "${fpath[1]}/_sshgo"

# Fish
sshgo completion fish | source
```