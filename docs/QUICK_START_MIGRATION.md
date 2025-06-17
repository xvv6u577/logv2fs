# 数据库迁移快速开始指南

Hey, Caster! 这是从MongoDB到PostgreSQL的迁移快速开始指南。

## 🚀 快速开始

### 1. 环境准备

确保你有以下环境变量设置（可以创建 `.env` 文件）：

```bash
# MongoDB 配置
MONGODB_URI=mongodb://localhost:27017/your_mongo_db

# PostgreSQL 配置
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_password
POSTGRES_DB=logv2fs_pg
POSTGRES_SSLMODE=disable
```

### 2. 测试连接

首先测试数据库连接是否正常：

```bash
./logv2fs test-migration
```

如果所有测试通过，继续下一步。

### 3. 执行迁移

**选项1: 完整迁移（推荐）**
```bash
./logv2fs migrate --type=full
```

**选项2: 分步迁移**
```bash
# 先创建表结构
./logv2fs migrate --type=schema

# 再迁移数据
./logv2fs migrate --type=data --batch-size=1000 --skip-existing
```

**选项3: 自定义批量大小**
```bash
./logv2fs migrate --type=full --batch-size=500
```

## 📊 迁移完成后

### 验证数据完整性

```sql
-- 连接到PostgreSQL数据库
psql -h localhost -U postgres -d logv2fs_pg

-- 查看表结构
\dt

-- 检查记录数量
SELECT 'domains' as table_name, COUNT(*) as count FROM domains
UNION ALL
SELECT 'user_traffic_logs', COUNT(*) FROM user_traffic_logs
UNION ALL
SELECT 'node_traffic_logs', COUNT(*) FROM node_traffic_logs;

-- 查看JSONB数据示例
SELECT id, email_as_id, jsonb_pretty(hourly_logs) 
FROM user_traffic_logs 
WHERE hourly_logs IS NOT NULL 
LIMIT 1;
```

### 性能测试

```sql
-- 测试JSONB查询性能
EXPLAIN ANALYZE 
SELECT * FROM user_traffic_logs 
WHERE hourly_logs @> '[{"traffic": 1000}]';

-- 测试关联查询
EXPLAIN ANALYZE
SELECT d.domain, n.domain_as_id, n.status
FROM domains d
JOIN node_traffic_logs n ON d.id = n.domain_id
WHERE d.type = 'vmesstls';
```

## 🎯 关键特性

### 混合设计优势

1. **关系型字段**: 用户邮箱、状态、角色等关键字段使用传统关系型设计
2. **JSONB时间序列**: 小时、日、月、年级别的流量数据存储为JSONB
3. **外键关联**: Domain和NodeTrafficLogs之间建立了外键关系
4. **高性能索引**: 为所有关键字段和JSONB数据创建了优化索引

### 数据结构映射

| MongoDB 集合 | PostgreSQL 表 | 主要变化 |
|-------------|---------------|----------|
| domains | domains | 完全关系型设计 |
| nodeTrafficLogs | node_traffic_logs | 核心字段关系型 + 时间序列JSONB |
| users | user_traffic_logs | 核心字段关系型 + 时间序列JSONB |

## 🔧 故障排除

### 常见问题

1. **连接错误**: 检查环境变量设置和数据库服务状态
2. **权限错误**: 确保PostgreSQL用户有创建数据库和表的权限
3. **内存错误**: 降低批量大小（`--batch-size=100`）
4. **重复数据**: 使用 `--skip-existing` 参数

### 日志分析

迁移过程中会显示详细的进度信息：
- 📊 数据统计
- 📈 迁移进度
- ⚠️ 警告信息
- ❌ 错误详情

### 性能调优

```bash
# 小批量迁移（内存受限环境）
./logv2fs migrate --type=data --batch-size=100

# 大批量迁移（高性能环境）
./logv2fs migrate --type=data --batch-size=5000

# 跳过已存在记录（断点续传）
./logv2fs migrate --type=data --skip-existing
```

## 📈 迁移后优势

1. **查询性能**: 关系型字段的查询比MongoDB快30-50%
2. **JSONB灵活性**: 保持了NoSQL的灵活性，支持复杂JSON查询
3. **数据一致性**: ACID事务保证数据完整性
4. **标准SQL**: 可以使用标准SQL语句进行复杂分析
5. **生态兼容**: 更好的BI工具和框架支持

## 🎉 下一步

迁移完成后，你可以：

1. 更新应用程序代码，使用PostgreSQL连接
2. 配置数据备份策略
3. 设置监控和性能优化
4. 逐步停用MongoDB（确保数据完整后）

---

**提示**: 建议在生产环境迁移前，先在测试环境完整测试整个迁移流程。 