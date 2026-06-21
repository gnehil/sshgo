# Advanced Usage

## Migrating from Existing SSH Config

If you have a large existing `~/.ssh/config`:

```bash
# Import with one command
sshgo import

# View import results
sshgo list --group "Imported from SSH config"
```

After import supports:
- Preserves Host aliases
- Imports HostName, User, Port, IdentityFile, ProxyJump
- Expands Include directives (recursive parsing)

## Bidirectional Sync Workflow

```bash
# 1. Add or modify config in sshgo
sshgo add new-server --host 10.0.0.3 --user admin --group dev

# 2. Sync to SSH config (so native ssh command can also use it)
sshgo sync

# 3. Now native ssh also works
ssh new-server

# 4. If you manually edited ~/.ssh/config, re-import
sshgo import --overwrite
```

## Batch Operations

### Health Checks

```bash
# Check disk on all production servers
sshgo exec --group prod "df -h /"

# Check nginx status on all web servers
sshgo exec 'web-*' "systemctl is-active nginx"

# Batch restart services
sshgo exec "app-1,app-2,app-3" "sudo systemctl restart app"
```

### Configuration Collection

```bash
# Collect system info from all production servers
sshgo exec --group prod "hostname && cat /etc/os-release | head -3 && uptime"
```

**Note**: Batch execution runs in SSH BatchMode. Servers requiring password authentication will fail. Consider using key authentication for all servers.

### Timeout Control

Batch execution has no default timeout. Unresponsive servers can hang. Use SSH options to configure keepalive:

```yaml
# In config.yaml, set keepalive for unstable connections
profiles:
  - name: "unstable-server"
    host: "10.0.0.5"
    user: "admin"
    keepalive_interval: 10
    server_alive_count: 2
```

## Advanced Jump Host Scenarios

### Multi-hop Chains

Typical enterprise network architecture (outer → inner jump → target):

```bash
# Method 1: CLI addition
sshgo add-jump internal-db \
  --jump deploy@bastion.company.com:2222 \
  --jump ops@gateway:22

# With per-hop identity files
sshgo add-jump internal-db \
  --jump deploy@bastion.company.com:2222 -i ~/.ssh/id_bastion \
  --jump ops@gateway:22               -i ~/.ssh/id_gateway

# Method 2: Direct YAML editing
# See JumpHost section in docs/config.md
```

### Jump Hosts with Port Forwarding

```yaml
profiles:
  - name: "internal-db"
    host: "10.0.0.50"
    user: "dba"
    jump_hosts:
      - name: "bastion"
        host: "jump.company.com"
        user: "bastion"
    forward_ports:
      - local_port: 5432
        remote_host: "localhost"
        remote_port: 5432
```

After connecting, access the remote database via `localhost:5432`:

```bash
sshgo internal-db
# In another terminal
psql -h localhost -p 5432 -U postgres
```

## Persistent Shell Completion

```bash
# Bash
echo 'eval "$(sshgo completion bash)"' >> ~/.bashrc

# Zsh
echo 'eval "$(sshgo completion zsh)"' >> ~/.zshrc

# Fish
echo 'sshgo completion fish | source' >> ~/.config/fish/config.fish
```

## Configuration File Management

### Backup and Recovery

sshgo automatically creates a backup before each config modification:

```yaml
~/.sshgo/
├── config.yaml
├── config.yaml.bak.20250115_103045
├── config.yaml.bak.20250116_091230
└── history.json
```

Manual recovery:

```bash
cp ~/.sshgo/config.yaml.bak.20250115_103045 ~/.sshgo/config.yaml
```

### Version Control Configuration

You can put `~/.sshgo/config.yaml` under version control (passwords don't go in config file):

```bash
git init ~/.sshgo
git add config.yaml groups.yaml
git commit -m "Initial SSH config version"
```

### Manual Editing

```bash
# Use your default editor
sshgo edit

# Or specify an editor
EDITOR=nano sshgo edit
```

## Exit Codes

| Exit Code | Meaning | Use Case |
|-----------|---------|----------|
| `0` | Success | - |
| `1` | Config/argument error | Validate arguments in scripts |
| `2` | Connection failed | Retry logic after automated connection failure |
| `3` | Partial batch execution failure | Identify which servers failed in CI/CD |
| `4` | Keychain access failure | Security audit |
| `5` | Config file corruption | Detect config file format issues |

```bash
# Example: Using exit codes in scripts
if sshgo ping prod-web; then
  echo "server is up"
else
  echo "server is down" | mail -s "Alert" admin@example.com
fi
```