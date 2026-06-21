# 快速开始

## 1. 添加分组（可选）

虽然分组不是必需的，但建议先创建组织：

```bash
sshgo group add prod --description "Production servers"
sshgo group add dev --description "Development servers"
sshgo group add staging --description "Staging servers"
```

查看所有分组：

```bash
sshgo group list
```

## 2. 添加连接

```bash
# 基本连接
sshgo add web-prod --host 10.0.0.1 --user deploy --group prod

# 自定义端口
sshgo add web-prod --host 10.0.0.1 --user deploy -p 2222 --group prod

# 指定密钥文件
sshgo add db-prod --host 10.0.0.2 --user dbadmin \
  --identity-file ~/.ssh/id_rsa_prod --group prod

# 注意：sshgo 要求密钥文件仅本人可读（如 0o600），
# 与 OpenSSH 策略一致。权限过宽（如 0o644）会在 add 时被拒绝，
# 避免后续连接时出现 "Permissions 0644 ... are too open" 错误。

# 配置心跳
sshgo add db-prod --host 10.0.0.2 --user dbadmin --group prod
# 心跳参数后续通过编辑 config.yaml 添加：
# keepalive_interval: 30
# server_alive_count: 3
```

## 3. 连接服务器

```bash
# 快速连接
sshgo web-prod

# 等价于
sshgo connect web-prod

# 从历史中选择
sshgo connect --recent
```

## 4. 查看和管理

```bash
# 表格列出所有连接
sshgo list

# 按分组过滤
sshgo list --group prod

# JSON 格式输出
sshgo list --format json

# 按名称排序
sshgo list --sort name

# 查看单个配置详情
sshgo show web-prod
```

## 5. 导入现有 SSH 配置

如果你已有 `~/.ssh/config`，一键导入：

```bash
# 默认从 ~/.ssh/config 导入
sshgo import

# 指定文件
sshgo import --file /path/to/ssh-config

# 覆盖已有同名配置
sshgo import --overwrite
```

导入的配置自动标记为 `Imported from SSH config` 分组。

## 下一步

- 阅读 [命令参考](commands.md) 了解所有命令
- 阅读 [配置文件格式](config.md) 了解 YAML 结构
- 阅读 [高级用法](advanced.md) 了解跳板机、端口转发、批量执行