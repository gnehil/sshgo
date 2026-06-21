# Quick Start

## 1. Add Groups (Optional)

While groups are not required, it's recommended to create some for organization:

```bash
sshgo group add prod --description "Production servers"
sshgo group add dev --description "Development servers"
sshgo group add staging --description "Staging servers"
```

View all groups:

```bash
sshgo group list
```

## 2. Add Connections

```bash
# Basic connection
sshgo add web-prod --host 10.0.0.1 --user deploy --group prod

# Custom port
sshgo add web-prod --host 10.0.0.1 --user deploy -p 2222 --group prod

# Specify identity file
sshgo add db-prod --host 10.0.0.2 --user dbadmin \
  --identity-file ~/.ssh/id_rsa_prod --group prod

# Note: sshgo requires the identity file to be readable only by you
# (e.g. mode 0o600). Files with group/other access are rejected to match
# OpenSSH's policy and avoid "Permissions 0644 ... are too open" later.

# Configure keepalive
sshgo add db-prod --host 10.0.0.2 --user dbadmin --group prod
# Keepalive parameters can be added later via editing config.yaml:
# keepalive_interval: 30
# server_alive_count: 3
```

## 3. Connect to Server

```bash
# Quick connect
sshgo web-prod

# Equivalent to
sshgo connect web-prod

# Choose from history
sshgo connect --recent
```

## 4. List and Manage

```bash
# List all connections in table format
sshgo list

# Filter by group
sshgo list --group prod

# JSON output
sshgo list --format json

# Sort by name
sshgo list --sort name

# Show single profile details
sshgo show web-prod
```

## 5. Import Existing SSH Config

If you already have `~/.ssh/config`, import it with one command:

```bash
# Import from default ~/.ssh/config
sshgo import

# Specify custom file
sshgo import --file /path/to/ssh-config

# Overwrite existing profiles with same name
sshgo import --overwrite
```

Imported profiles are automatically tagged with the `Imported from SSH config` group.

## Next Steps

- Read [Command Reference](commands.md) for all available commands
- Read [Configuration Format](config.md) to understand the YAML structure
- Read [Advanced Usage](advanced.md) for jump hosts, port forwarding, and batch execution