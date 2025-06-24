# PostgreSQL Sing-box 支持配置指南

## 🎯 **功能简介**

本文档介绍如何为 `./main singbox` 命令启用 PostgreSQL 数据库支持。

## 📋 **前提条件**

1. **PostgreSQL 数据库**已运行并可连接
2. **环境变量**已正确配置
3. **数据库表结构**已创建（可通过迁移命令创建）

## ⚙️ **环境变量配置**

### 方式一：使用 PostgreSQL URI

```bash
# 启用 PostgreSQL 支持
USE_POSTGRES=true

# PostgreSQL 连接 URI
postgresURI=postgresql://username:password@localhost:5432/logv2fs?sslmode=disable

# Sing-box 配置文件路径
SING_BOX_TEMPLATE_CONFIG=./config/template_singbox.json

# 其他必要配置
CURRENT_DOMAIN=your-domain.com
```

### 方式二：使用分离的环境变量

```bash
# 启用 PostgreSQL 支持
USE_POSTGRES=true

# PostgreSQL 连接参数
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=yourpassword
POSTGRES_DB=logv2fs
POSTGRES_SSLMODE=disable

# Sing-box 配置文件路径
SING_BOX_TEMPLATE_CONFIG=./config/template_singbox.json

# 其他必要配置
CURRENT_DOMAIN=your-domain.com
```

## 🚀 **使用步骤**

### 1. 数据库初始化（如果需要）

```bash
# 创建数据库和表结构
./main migrate schema

# 迁移数据（如果从 MongoDB 迁移）
./main migrate data --batch-size 100
```

### 2. 验证数据库连接

```bash
# 测试迁移连接
./main test-migration
```

### 3. 启动 Sing-box

```bash
# 启动 sing-box 服务（现在会自动使用 PostgreSQL）
./main singbox
```

## 📊 **工作原理**

### 数据库选择逻辑

```go
// 在 pkg/singbox.go 中
func UpdateOptionsFromDB(opt option.Options) (option.Options, error) {
    // 根据环境变量决定使用哪种数据库
    if database.IsUsingPostgres() {
        log.Println("使用 PostgreSQL 从数据库更新 sing-box 配置...")
        return updateOptionsFromPostgreSQL(opt)
    } else {
        log.Println("使用 MongoDB 从数据库更新 sing-box 配置...")
        return updateOptionsFromMongoDB(opt)
    }
}
```

### PostgreSQL 查询逻辑

```sql
-- 查询活跃用户
SELECT email_as_id, status, uuid, user_id 
FROM user_traffic_logs 
WHERE status = 'plain'
```

### 用户配置生成

- **VLess 用户**: `email_as_id + "-reality"`
- **Hysteria2 用户**: `email_as_id + "-hysteria2"`
- **统计用户**: 添加到 V2RayAPI Stats 中

## 🔧 **故障排除**

### 常见问题

1. **PostgreSQL 连接失败**
   ```
   错误: PostgreSQL数据库连接失败: xxx
   解决: 检查 postgresURI 或 POSTGRES_* 环境变量
   ```

2. **自动回退到 MongoDB**
   ```
   日志: PostgreSQL 数据库连接不可用，尝试回退到 MongoDB
   原因: PostgreSQL 连接失败，系统自动使用 MongoDB
   ```

3. **没有找到活跃用户**
   ```
   日志: PostgreSQL 中没有找到活跃用户
   原因: user_traffic_logs 表中没有 status='plain' 的用户
   ```

### 日志检查

启动 sing-box 时，注意以下日志：

```
# 成功使用 PostgreSQL
使用 PostgreSQL 从数据库更新 sing-box 配置...
成功从 PostgreSQL 加载了 X 个用户的配置

# 回退到 MongoDB
使用 MongoDB 从数据库更新 sing-box 配置...
成功从 MongoDB 加载了 X 个用户的配置
```

## 📈 **性能对比**

| 特性 | MongoDB | PostgreSQL |
|------|---------|------------|
| 查询性能 | 快速文档查询 | 关系查询+索引优化 |
| 数据一致性 | 最终一致性 | ACID 事务 |
| 扩展性 | 水平扩展 | 垂直扩展为主 |
| 复杂查询 | 聚合管道 | SQL 查询 |

## 🔄 **数据同步**

如果需要 MongoDB 和 PostgreSQL 之间的数据同步：

```bash
# 双向同步
./main sync --batch-size 100

# 仅 MongoDB -> PostgreSQL
./main sync --direction mongo-to-postgres

# 仅 PostgreSQL -> MongoDB  
./main sync --direction postgres-to-mongo
```

## ⚠️ **注意事项**

1. **环境变量优先级**: `USE_POSTGRES` 环境变量决定数据库选择
2. **向后兼容**: 如果未设置 `USE_POSTGRES=true`，自动使用 MongoDB
3. **错误处理**: PostgreSQL 连接失败会自动回退到 MongoDB
4. **配置文件**: 确保 `SING_BOX_TEMPLATE_CONFIG` 指向正确的配置文件

## 🎉 **成功验证**

如果看到以下日志，说明 PostgreSQL 支持已成功启用：

```
PostgreSQL连接成功
使用 PostgreSQL 从数据库更新 sing-box 配置...
成功从 PostgreSQL 加载了 X 个用户的配置
``` 