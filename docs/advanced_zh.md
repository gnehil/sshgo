# 高级用法

## 从现有 SSH 配置迁移

如果你有大量已有的 `~/.ssh/config` 配置：

```bash
# 一键导入
sshgo import

# 查看导入结果
sshgo list --group "Imported from SSH config"
```

导入后支持：
- 保留 Host 别名
- 导入 HostName, User, Port, IdentityFile, ProxyJump
- Include 指令展开（递归解析）

## 双向同步工作流

```bash
# 1. sshgo 中添加或修改配置
sshgo add new-server --host 10.0.0.3 --user admin --group dev

# 2. 同步到 SSH config（使原生 ssh 命令也能使用）
sshgo sync

# 3. 现在原生 ssh 也可以使用
ssh new-server

# 4. 如果手动编辑了 ~/.ssh/config，重新导入
sshgo import --overwrite
```

## 批量操作模式

### 巡检检查

```bash
# 检查所有生产服务器磁盘
sshgo exec --group prod "df -h /"

# 检查所有 web 服务器的服务状态
sshgo exec 'web-*' "systemctl is-active nginx"

# 批量重启服务
sshgo exec "app-1,app-2,app-3" "sudo systemctl restart app"
```

### 配置收集

```bash
# 收集所有服务器的系统信息
sshgo exec --group prod "hostname && cat /etc/os-release | head -3 && uptime"
```

**注意**：批量执行在 SSH BatchMode 下运行，需要密码认证的服务器会失败。建议全部使用密钥认证。

### 超时控制

批量执行无默认超时限制，如果某些服务器无响应会挂起。可以配合 SSH 选项使用：

```bash
# 在 config.yaml 中为不稳定连接设置心跳
profiles:
  - name: "unstable-server"
    host: "10.0.0.5"
    user: "admin"
    keepalive_interval: 10
    server_alive_count: 2
```

## 跳板机高级场景

### 多层跳板链

典型的企业网络架构（外层 → 内层跳板 → 目标）：

```bash
# 方式一：CLI 添加
sshgo add-jump internal-db \
  --jump deploy@bastion.company.com:2222 \
  --jump ops@gateway:22

# 方式二：直接编辑 YAML
# 详见 docs/config.md 中的 JumpHost 章节
```

### 跳板机与端口转发结合

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

连接后本地可通过 `localhost:5432` 访问远程数据库：

```bash
sshgo internal-db
# 在另一个终端
psql -h localhost -p 5432 -U postgres
```

## Shell 补全持久化

```bash
# Bash
echo 'eval "$(sshgo completion bash)"' >> ~/.bashrc

# Zsh
echo 'eval "$(sshgo completion zsh)"' >> ~/.zshrc

# Fish
echo 'sshgo completion fish | source' >> ~/.config/fish/config.fish
```

## 配置文件管理技巧

### 备份恢复

sshgo 每次修改配置前自动创建备份：

```yaml
~/.sshgo/
├── config.yaml
├── config.yaml.bak.20250115_103045
├── config.yaml.bak.20250116_091230
└── history.json
```

手动恢复：

```bash
cp ~/.sshgo/config.yaml.bak.20250115_103045 ~/.sshgo/config.yaml
```

### 配置文件版本控制

可以将 `~/.sshgo/config.yaml` 纳入版本控制（密码不进配置文件）：

```bash
git init ~/.sshgo
git add config.yaml groups.yaml
git commit -m "Initial SSH config version"
```

### 手动编辑

```bash
# 使用你的默认编辑器
sshgo edit

# 或指定编辑器
EDITOR=nano sshgo edit
```

## 退出码

| 退出码 | 含义 | 使用场景 |
|--------|------|----------|
| `0` | 成功 | - |
| `1` | 配置/参数错误 | 脚本中检测参数合法性 |
| `2` | 连接失败 | 自动化连接失败后重试逻辑 |
| `3` | 批量执行部分失败 | CI/CD 中判断哪些服务器失败 |
| `4` | 密钥链存取失败 | 安全审计 |
| `5` | 配置文件损坏 | 检测配置文件格式问题 |

```bash
# 示例：脚本中使用退出码
if sshgo ping prod-web; then
  echo "server is up"
else
  echo "server is down" | mail -s "Alert" admin@example.com
fi
```