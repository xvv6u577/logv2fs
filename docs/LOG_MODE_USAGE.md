# 日志模式配置功能

本功能允许根据 `GIN_MODE` 环境变量自动设置数据库日志级别，以优化不同环境下的日志输出。

## 🚀 功能特性

根据 `GIN_MODE` 环境变量自动设置相应的数据库日志级别：

| GIN_MODE | 日志级别 | 说明 | 适用场景 |
|----------|----------|------|----------|
| `debug` | `logger.Info` | 显示详细日志信息 | 开发调试环境 |
| `release` | `logger.Error` | 只显示错误信息 | 生产环境 |
| `test` | `logger.Silent` | 静默模式，不输出日志 | 测试环境 |
| 未设置/其他 | `logger.Info` | 默认详细日志 | 默认情况 |

## 📖 使用方法

### 1. 设置环境变量

在您的 `.env` 文件或系统环境变量中设置：

```bash
# 开发环境
GIN_MODE=debug

# 生产环境  
GIN_MODE=release

# 测试环境
GIN_MODE=test
```

### 2. 启动应用

```bash
# 方法1：通过环境变量文件
echo "GIN_MODE=debug" >> .env
./logv2fs httpserver

# 方法2：直接设置环境变量
GIN_MODE=release ./logv2fs httpserver

# 方法3：测试模式
GIN_MODE=test ./logv2fs httpserver
```

## 🎯 实现原理

代码实现位于 `database/postgres_connection.go` 文件中：

```go
// getLogLevel 根据 GIN_MODE 环境变量返回对应的日志级别
func getLogLevel() logger.LogLevel {
    ginMode := os.Getenv("GIN_MODE")
    
    switch ginMode {
    case "debug":
        return logger.Info  // 开发环境，显示详细日志
    case "release":
        return logger.Error // 生产环境，只显示错误信息
    case "test":
        return logger.Silent // 测试环境，静默模式
    default:
        // 默认情况下使用 Info 级别
        return logger.Info
    }
}
```

## 🔧 代码变更

主要修改位于 `InitPostgreSQL()` 函数：

```go
// 之前（硬编码）
Logger: logger.Default.LogMode(logger.Info),

// 之后（动态配置）
Logger: logger.Default.LogMode(getLogLevel()),
```

## 💡 最佳实践

1. **开发环境**：使用 `debug` 模式查看详细的数据库操作日志
2. **生产环境**：使用 `release` 模式减少日志输出，提高性能
3. **测试环境**：使用 `test` 模式保持测试输出的清洁
4. **CI/CD**：在不同的部署阶段自动设置相应的 `GIN_MODE`

## 🎪 效果示例

### Debug 模式输出示例
```
2024/01/20 10:30:15 PostgreSQL连接成功
[GORM] 2024/01/20 10:30:15 SELECT * FROM user_traffic_logs WHERE email_as_id = $1
[GORM] 2024/01/20 10:30:15 INSERT INTO user_traffic_logs ...
```

### Release 模式输出示例
```
2024/01/20 10:30:15 PostgreSQL连接成功
[ERROR] 2024/01/20 10:30:16 Connection failed: ...
```

### Test 模式输出示例
```
(无数据库相关日志输出)
```

---

这个功能实现了 **Interactive Task Loop**（交互式任务循环）的设计理念，根据不同的运行环境自动调整行为，提升了系统的适应性和可维护性。 