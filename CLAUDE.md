# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

logv2fs 是一个基于 Go 的 VPN/代理服务端，构建在 sing-box 之上，配套一个 React 前端。当前架构使用 **MongoDB 作为唯一数据库**，并通过 WebSocket 推送实时变更。

## Core Development Philosophy

### KISS (Keep It Simple, Stupid)

Simplicity should be a key goal in design. Choose straightforward solutions over complex ones whenever possible. Simple solutions are easier to understand, maintain, and debug.

### YAGNI (You Aren't Gonna Need It)

Avoid building functionality on speculation. Implement features only when they are needed, not when you anticipate they might be useful in the future.

### Design Principles

- **Dependency Inversion**: High-level modules should not depend on low-level modules. Both should depend on abstractions.
- **Open/Closed Principle**: Software entities should be open for extension but closed for modification.
- **Single Responsibility**: Each function, class, and module should have one clear purpose.
- **Fail Fast**: Check for potential errors early and raise exceptions immediately when issues occur.


## Common Development Commands

### Backend Development
```bash
# 启动主应用服务
./main httpserver  # Web API 服务（包含静态前端文件）
./main singbox     # sing-box 代理服务（含定时流量统计）

# 开发态
make backend       # 等价于 go run ./ httpserver
make singbox       # 带完整 build tags 的 sing-box

# 生产构建（需要完整 build tags）
export GOOS=linux; export GOARCH=amd64
go build -tags="with_gvisor,with_quic,with_wireguard,with_utls,with_reality_server,with_clash_api,with_v2ray_api,with_grpc" -o ./logv2fs ./
```

### Frontend Development
```bash
cd frontend
npm i && npm run build  # 生产构建
make web                # 开发态，自动代理到后端 API
```

## Architecture Overview

### 顶层目录结构

| 目录 | 角色 |
|---|---|
| `cmd/` | Cobra CLI 入口（`httpserver`、`singbox` 两个子命令）|
| `controllers/` | Gin HTTP 控制器（用户、节点、缴费） |
| `routers/` | Gin 路由组装（公共路由 + 鉴权路由） |
| `database/` | MongoDB 连接与集合获取 |
| `model/` | 业务模型（含 BSON 标签） |
| `jobs/` | 定时任务（每 15 分钟拉取 sing-box 流量并写库） |
| `singbox/` | sing-box 集成层（从 MongoDB 加载用户配置、读取流量统计） |
| `websocket/` | WebSocket 实时推送 |
| `middleware/` | JWT 鉴权、CORS |
| `helpers/` | 通用工具（输入清洗、IP 格式化等） |
| `config/` | sing-box / Clash / 错误页等配置模板 |
| `frontend/` | React + Tailwind 前端 |

### 核心组件

- **Main Application**：Cobra CLI，目前注册两个子命令 `httpserver` 与 `singbox`
- **Database Layer**：MongoDB（`logV2rayTrafficDB`），通过 `database.GetCollection(model.X{})` 按模型自动取集合
- **Proxy Engine**：sing-box，配合 `singbox` 包做用户注入和流量读取
- **Real-time Layer**：WebSocket Hub 广播
- **Frontend**：React SPA + Redux

### MongoDB 集合

- `USER_TRAFFIC_LOGS` —— 用户信息及小时/日/月/年流量日志
- `NODE_TRAFFIC_LOGS` —— 节点流量统计
- `subscription_nodes` —— 代理节点配置
- `payment_records` —— 缴费记录（费用领域唯一事实表，月/年统计在接口里实时分摊算出，不再物化每日分摊集合）
- `CUSTOM_DATES` —— 节点自定义日期

集合名通过 `model.X.CollectionName()` 暴露，避免硬编码。

### 关键文件

- `controllers/controller.go` —— 用户/登录/订阅/节点 JSON 等
- `controllers/node.go` —— 节点增删改查
- `controllers/payment.go` —— 缴费录入与统计
- `model/types.go` —— 用户与流量日志模型
- `model/node.go` —— 节点与订阅节点模型
- `model/payment.go` —— 缴费模型与统计结构
- `model/singbox.go` —— sing-box / Verge 配置 JSON/YAML 结构
- `singbox/singbox_common.go` —— `UsageDataOfAll`、`InitOptionsFromConfig`
- `singbox/singbox_mongodb.go` —— `UpdateOptionsFromMongoDB`
- `jobs/cron.go` —— 流量定时任务
- `websocket/websocket.go` —— WebSocket Hub

## Development Patterns

### Traffic Logging
sing-box 的 v2ray API stats 每 15 分钟被 `jobs.Cron_loggingJobs` 抓一次，按用户/节点维度增量写入 MongoDB 的 `hourly_logs / daily_logs / monthly_logs / yearly_logs`，同时累加 `used` 字段。

### Configuration Generation
针对不同客户端生成订阅配置：
- **sing-box**：JSON（基于 `config/template_singbox.json`）
- **Clash Verge rev**：YAML（基于 `config/template_verge.yaml`）
- **Shadowrocket**：Base64 订阅 URL

### Real-time Features
WebSocket 推送用户流量、节点状态、缴费、用户启停等事件。

## API Structure

### 鉴权
JWT + bcrypt，区分 `admin` 与 `normal` 角色，受保护路由全部经 `middleware.Authentication()`。

### 主要端点
```
POST   /v1/login                登录
POST   /v1/signup               用户注册
POST   /v1/edit/:name           编辑用户
GET    /v1/user/:name           查询用户
GET    /v1/n778cf               用户列表（混淆路径）
GET    /v1/getconfig/:name      生成代理配置
GET    /v1/c47kr8               sing-box 节点列表
GET    /v1/subscription-nodes   订阅节点
PUT    /v1/upsert-nodes         批量插入/覆盖节点
POST   /v1/payment              录入缴费
GET    /v1/payment/statistics   缴费统计
GET    /singbox/:name           sing-box JSON 订阅
GET    /verge/:name             Verge YAML 订阅
GET    /static/:name            Shadowrocket Base64 订阅
GET    /ws                      WebSocket
```

完整列表见 `routers/authorized.go` 与 `routers/public.go`。

## Environment Configuration

```bash
mongoURI="mongodb://127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+1.5.4"

CURRENT_DOMAIN="localhost"
SERVER_ADDRESS=""
SERVER_PORT=8079
GIN_MODE=release    # debug / release / test

SING_BOX_TEMPLATE_CONFIG=./config/template_singbox.json
```

### Build Tags
生产构建必须带：
```
with_gvisor,with_quic,with_wireguard,with_utls,with_reality_server,with_clash_api,with_v2ray_api,with_grpc
```

## Testing and Deployment

### Local Development
```bash
make backend && make web   # 后端 API + 前端 dev server
```

### Production Deployment
```bash
docker-compose up -d
```

## IPv6 Support

应用全面支持 IPv6：
- 节点 IP 字段同时接受 IPv4 / IPv6
- 订阅 URL 生成会用 `helpers.FormatIPForURL` 自动加方括号
- 数据库字段长度足够容纳完整 IPv6

## Security Considerations

- 所有用户输入走 `github.com/mrz1836/go-sanitize` 清洗
- 注入防护、CORS、自动证书管理
- 密码 bcrypt 哈希，JWT 过期与刷新
- WebSocket 仅广播给已鉴权连接

## ⚠️ Important Notes

- **NEVER ASSUME OR GUESS** —— 不确定就先问
- **修改 import 路径或模型字段时，记得同步检查 `frontend/` 与文档**
- **新增路由时，需同步更新 `routers/public.go` 或 `routers/authorized.go`**
- **代码改动后，跑 `go build -tags=...` 完整编译一次再提交**
