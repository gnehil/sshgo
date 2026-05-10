# sshgo

[中文](docs/README_zh.md) · [English](README.md)

> A pure CLI SSH session management tool with capabilities rivaling Tabby.

## Features

- **Quick Connect** — `sshgo <alias>` connects instantly
- **Groups** — Organize connections by project/environment
- **Jump Host Support** — Single and multi-hop jump host chains
- **Port Forwarding** — Built-in LocalForward management
- **Batch Execution** — Run commands on multiple servers in parallel
- **Import/Export** — Seamless import from and export to `~/.ssh/config`
- **Connection History** — Auto-logged, supports `--recent` for quick reconnect
- **Secure Credentials** — Passwords stored in OS native keychain
- **Shell Completion** — Supports Bash / Zsh / Fish

## Installation

```bash
# Clone the repository
git clone https://github.com/gnehil/sshgo.git
cd sshgo

# Build
go build -o sshgo .

# Install to PATH (optional)
sudo mv sshgo /usr/local/bin/
```

### Shell Completion

```bash
# Bash
source <(./sshgo completion bash)
# Or install permanently
./sshgo completion bash > /etc/bash_completion.d/sshgo

# Zsh
./sshgo completion zsh > "${fpath[1]}/_sshgo"

# Fish
./sshgo completion fish | source
```

## Quick Start

```bash
# Add your first connection
sshgo add web-server --host 192.168.1.10 --user deploy -p 2222

# Quick connect
sshgo web-server
sshgo connect web-server

# List all connections
sshgo list
sshgo list --format json

# Delete a connection
sshgo delete web-server
```

## Commands

| Command | Description |
|---------|-------------|
| `sshgo add <name>` | Add a connection profile |
| `sshgo list` | List all connections |
| `sshgo show <name>` | Show profile details |
| `sshgo edit <name>` | Edit profile |
| `sshgo delete <name>` | Delete a profile |
| `sshgo connect <name>` | Connect to a server |
| `sshgo connect --recent` | Choose from connection history |
| `sshgo group list/add/delete` | Manage groups |
| `sshgo add-jump <name>` | Configure jump hosts |
| `sshgo forward add/list` | Manage port forwarding |
| `sshgo exec <pattern>` | Batch command execution |
| `sshgo import` | Import from SSH config |
| `sshgo sync` | Sync to SSH config |
| `sshgo ping <name>` | Test connectivity |
| `sshgo history` | View connection history |
| `sshgo completion` | Generate shell completions |

## Configuration

Location: `~/.sshgo/config.yaml`

```yaml
profiles:
  - name: "my-server"
    host: "192.168.1.10"
    port: 22
    user: "deploy"
    group: "prod"
    identity_file: "~/.ssh/id_rsa_prod"
    jump_hosts:
      - name: "bastion"
        host: "jump.example.com"
        user: "jumpuser"
        port: 22
    forward_ports:
      - local_port: 8080
        remote_host: "localhost"
        remote_port: 80
    keepalive_interval: 30
    server_alive_count: 3

groups:
  - name: "prod"
    description: "Production servers"
```

## Documentation

- [Quick Start](docs/quickstart.md)
- [Command Reference](docs/commands.md)
- [Configuration Format](docs/config.md)
- [Advanced Usage](docs/advanced.md)

## License

MIT