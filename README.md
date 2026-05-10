# sshgo

> 纯 CLI 的 SSH 会话管理工具，媲美 Tabby 的 SSH 管理能力。

## 特性

- **快速连接** — `sshgo <别名>` 一键连接
- **分组管理** — 按项目/环境组织连接
- **跳板机支持** — 支持单层及多层跳板机链
- **端口转发** — 内置 LocalForward 管理
- **批量执行** — 对多台服务器并行执行命令
- **导入导出** — 无缝导入 `~/.ssh/config`，同步导出兼容 SSH
- **连接历史** — 自动记录，支持 `--recent` 快速重连
- **安全凭证** — 密码存储于 OS 原生密钥链
- **Shell 补全** — 支持 Bash / Zsh / Fish

## 安装

```bash
# 克隆仓库
git clone https://github.com/sshgo/sshgo.git
cd sshgo

# 编译
go build -o sshgo .

# 安装到 PATH (可选)
sudo mv sshgo /usr/local/bin/
```

### Shell 补全

```bash
# Bash
source <(./sshgo completion bash)
# 或永久安装
./sshgo completion bash > /etc/bash_completion.d/sshgo

# Zsh
./sshgo completion zsh > "${fpath[1]}/_sshgo"

# Fish
./sshgo completion fish | source
```

## 快速开始

```bash
# 添加第一个连接
sshgo add web-server --host 192.168.1.10 --user deploy -p 2222

# 快速连接
sshgo web-server
sshgo connect web-server

# 查看所有连接
sshgo list
sshgo list --format json

# 删除连接
sshgo delete web-server
```

## 命令参考

| 命令 | 说明 |
|------|------|
| `sshgo add <name>` | 添加连接配置 |
| `sshgo list` | 列出所有连接 |
| `sshgo show <name>` | 显示配置详情 |
| `sshgo edit <name>` | 编辑配置 |
| `sshgo delete <name>` | 删除配置 |
| `sshgo connect <name>` | 连接目标服务器 |
| `sshgo connect --recent` | 从历史中选择连接 |
| `sshgo group list/add/delete` | 分组管理 |
| `sshgo add-jump <name>` | 配置跳板机 |
| `sshgo forward add/list` | 端口转发管理 |
| `sshgo exec <pattern>` | 批量命令执行 |
| `sshgo import` | 从 SSH config 导入 |
| `sshgo sync` | 同步到 SSH config |
| `sshgo ping <name>` | 连接性测试 |
| `sshgo history` | 查看连接历史 |
| `sshgo completion` | 生成 Shell 补全 |

## 配置文件

位置：`~/.sshgo/config.yaml`

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

## License

MIT
