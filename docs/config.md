# 配置文件格式

sshgo 的配置文件位于 `~/.sshgo/config.yaml`，采用 YAML 格式。

## 目录结构

```yaml
profiles:    # 连接配置列表
  - ...

groups:      # 分组定义列表
  - ...
```

## 连接配置 (Profile)

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `name` | string | 是 | - | 连接别名。仅允许小写字母、数字、连字符，不能以连字符开头 |
| `host` | string | 是 | - | 服务器 IP 或域名 |
| `port` | int | 否 | 22 | SSH 端口号（1-65535） |
| `user` | string | 是 | - | SSH 用户名 |
| `group` | string | 否 | - | 所属分组名称（需在 groups 中预定义） |
| `identity_file` | string | 否 | - | SSH 私钥路径（支持 `~` 展开） |
| `jump_hosts` | array | 否 | - | 跳板机列表（见下方） |
| `forward_ports` | array | 否 | - | 端口转发列表（见下方） |
| `keepalive_interval` | int | 否 | - | ServerAliveInterval（秒） |
| `server_alive_count` | int | 否 | - | ServerAliveCountMax 次数 |

### 示例

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

## 跳板机 (JumpHost)

每个跳板机是一个嵌套对象：

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `name` | string | 否 | - | 跳板机显示名称 |
| `host` | string | 是 | - | 跳板机地址 |
| `port` | int | 否 | 22 | 跳板机端口 |
| `user` | string | 是 | - | 跳板机用户名 |
| `identity_file` | string | 否 | - | 跳板机密钥路径 |

### 多层跳板示例

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

实际执行的 SSH 命令等价于：

```bash
ssh -i ~/.ssh/prod_key -J bastion-user@jump.company.com,gateway-user@10.0.1.1 dba@10.0.0.50
```

## 端口转发 (ForwardPort)

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `local_port` | int | 是 | 本地监听端口 |
| `remote_host` | string | 是 | 远程目标主机 |
| `remote_port` | int | 是 | 远程目标端口 |

### 示例

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

## 分组 (Group)

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 分组名称（唯一标识） |
| `description` | string | 否 | 分组描述 |

### 示例

```yaml
groups:
  - name: "prod"
    description: "Production servers"
  - name: "staging"
    description: "Staging servers"
  - name: "dev"
    description: "Development servers"
```

**注意**：Profile 的 `group` 字段必须引用已定义的分组，否则验证会报错。

## 连接历史

连接历史存储在 `~/.sshgo/history.json`，格式：

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

按最后连接时间倒序排列（最近的在前）。

## 安全须知

- **密码不存储在配置文件中**，仅存于 OS 原生密钥链（macOS Keychain / Windows Credential Manager / Linux Secret Service）
- 配置文件权限为 `0600`（仅所有者可读写）
- 每次修改配置前自动创建备份（`config.yaml.bak.<timestamp>`）
- 路径中的 `~` 自动展开为用户主目录
