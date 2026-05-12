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
make httpserver       # 等价于 go run ./ httpserver
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
| `singbox/` | sing-box 集成层（从 MongoDB 加载用户配置、读取流量统计、loopback 控制端口与 client） |
| `websocket/` | WebSocket 实时推送 |
| `middleware/` | JWT 鉴权、CORS |
| `helpers/` | 通用工具（输入清洗、IP 格式化等） |
| `config/` | sing-box / Clash / 错误页等配置模板 |
| `frontend/` | React + Tailwind 前端 |
| `third_party/sing-box/` | sing-box v1.8.1 本地 fork，加导出的运行时 AddUser/RemoveUser API；`go.mod` 用 `replace` 指向 |

### 核心组件

- **Main Application**：Cobra CLI，目前注册两个子命令 `httpserver` 与 `singbox`
- **Database Layer**：MongoDB（`logV2rayTrafficDB`），通过 `database.GetCollection(model.X{})` 按模型自动取集合
- **Proxy Engine**：sing-box，配合 `singbox` 包做用户注入和流量读取
- **Real-time Layer**：WebSocket Hub 广播
- **Frontend**：React SPA + Redux

### MongoDB 集合

按 `model.X.CollectionName()` 与代码实际落库为准（括号内为模型文件与说明）：

| 集合名 | 说明 |
|--------|------|
| `USER_TRAFFIC_LOGS` | 用户主文档：账号、鉴权、状态、`used` 累计用量等（`model/types.go`）。按日/月/年的明细**不存本集合**。 |
| `user_traffic_periods` | 用户周期流量：`kind` 为 `daily` / `monthly` / `yearly`，`period` 为 `yyyymmdd` / `yyyymm` / `yyyy`，字段 `traffic` 累加（`model/types.go`）。 |
| `NODE_TRAFFIC_LOGS` | 节点主文档：域名、状态、时间戳等（`model/node.go`）。周期明细同样**不存本集合**。 |
| `node_traffic_periods` | 节点周期流量：结构与用户侧类似，(owner 键为 `domain_as_id`)（`model/node.go`）。 |
| `subscription_nodes` | 订阅用代理节点配置（类型、IP、端口、REALITY/Hysteria2/VLESS CDN 等）（`model/node.go`）。 |
| `payment_records` | 缴费事实表；月/年等费用统计在查询接口中实时聚合，不单独物化日摊集合（`model/payment.go`）。 |
| `CUSTOM_DATES` | 节点自定义日期（`model/types.go`）。 |

读用户/节点接口可通过 `$lookup` 把 `user_traffic_periods` / `node_traffic_periods` 拼回响应里的 `daily_logs`、`monthly_logs`、`yearly_logs`，仅为 API 兼容形态，持久化以周期集合为准。

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

### Traffic Logging（流量统计）

1. **采样**：`jobs.Cron_loggingJobs` 以 cron 表达式 `0 */15 * * * *` **每 15 分钟**执行一次（`jobs/cron.go`）。
2. **数据来源**：`singbox.UsageDataOfAll` 调用 sing-box 启用的 **V2Ray Stats API**，带 `Reset_: true` 查询并重置计数器；用正则匹配形如 `user>>>…>>>traffic>>>…` 的统计名，按用户名（tag 里 `-` 前缀）汇总本轮字节数（`singbox/singbox_common.go`）。
3. **写库（用户）**：`LogUserTraffic` 对用户主文档 `USER_TRAFFIC_LOGS` 做 `$inc used`，并在 `user_traffic_periods` 上对当日、当月、当年三条 `(email_as_id, kind, period)` 记录 **upsert 累加 `traffic`**（`jobs/cron.go`）。
4. **写库（节点）**：`LogNodeTraffic` 用环境变量 `CURRENT_DOMAIN` 作为 `domain_as_id`，在已存在节点文档时刷新 `NODE_TRAFFIC_LOGS` 的 `updated_at`；周期增量写入 `node_traffic_periods`（同上按日/月/年 upsert）。
5. **无小时粒度**：代码中没有 `hourly_logs` 集合或小时周期；API 中的 `daily_logs` / `monthly_logs` / `yearly_logs` 来自周期集合或 `$lookup` 拼装。

### Configuration Generation
针对不同客户端生成订阅配置：
- **sing-box**：JSON（基于 `config/template_singbox.json`），路由 `GET /singbox/:name`
- **Clash Verge rev**：YAML（基于 `config/template_verge.yaml`），路由 `GET /verge/:name`
- **Shadowrocket / Surge / v2rayN 等**：**同一路由、同一格式** —— `GET /static/:name` 返回 **Base64** 编码的纯文本，内容为多行分享链接（`vless://…`、`hysteria2://…` 等，由 `controllers/controller.go` 中 `GetSubscripionURL` 按节点类型拼接后 `StdEncoding.EncodeToString`）。客户端填入「订阅 URL」即可；v2rayN 与 Shadowrocket 共用该端点与编码方式。

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
GET    /static/:name            Base64 多协议分享链接订阅（Shadowrocket / v2rayN 等）
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

# === sing-box 运行时用户管理（runtime user mgmt） ===
# 仅监听 loopback，不要暴露公网。httpserver 与 singbox 两个进程
# 必须配置相同的 token，否则联动失败（DB 仍为事实源不影响业务）。
SINGBOX_CONTROL_LISTEN=127.0.0.1:8479
SINGBOX_CONTROL_TOKEN=please-change-me-to-a-long-random-string
```

### Runtime User Management（sing-box 运行时增/禁/删用户）

httpserver 与 singbox 是两个独立进程，原本通过 MongoDB 单向同步，DB 改动需要重启 singbox 才生效。现在通过 fork sing-box（`third_party/sing-box/`）暴露 `AddUser/RemoveUser` 导出 API，并在 singbox 进程上跑一个 loopback 控制 HTTP 服务，httpserver 在四条用户路径（SignUp / DisableUser / EnableUser / DeleteUserByUserName）成功落库后异步推送变更。失败软降级，DB 仍是唯一事实源，详见 `singbox/control_server.go`、`singbox/control_client.go`、`singbox/usermgr.go`、`third_party/sing-box/`。

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
