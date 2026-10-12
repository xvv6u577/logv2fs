# 应用中心 - Cloudflare Worker 多功能管理平台

这是一个基于 Cloudflare Worker 和 Supabase PostgreSQL 的现代化多功能管理平台，包含待办事项和 SSL 证书监控两大核心功能。

## ✨ 功能特点

### 🎯 核心功能
- **待办事项管理** - 完整的 CRUD 操作，支持完成状态切换
- **SSL 证书监控** - 自动检测域名证书到期时间，支持标签分类
- **单页应用架构** - 流畅的用户体验，无需页面刷新
- **响应式设计** - 完美适配手机、平板、桌面设备

### 🔧 技术特点
- 使用 Cloudflare Worker 作为无服务器后端
- 使用 Supabase 作为 PostgreSQL 数据库服务
- 定时任务自动检查证书（Cron Triggers）
- 现代化 UI 设计（Tailwind CSS + Glassmorphism）
- 完整的 CORS 支持

## 前置条件

- [Node.js](https://nodejs.org/) (推荐v18+)
- [Wrangler CLI](https://developers.cloudflare.com/workers/wrangler/install-and-update/)
- [Supabase账户](https://supabase.com/)

## 设置步骤

### 1. 克隆并安装依赖

```bash
git clone <your-repo-url>
cd logsbfs-cloudflare
npm install
```

### 2. 设置Supabase

1. 在[Supabase](https://supabase.com/)创建一个新项目
2. 记下项目URL和匿名密钥(anon key)
3. 在 Supabase SQL 编辑器中依次运行以下迁移文件：
   - `migrations/create_todos_table.sql` - 创建待办事项表
   - `migrations/create_ssl_certificates_table.sql` - 创建 SSL 证书监控表

### 3. 配置环境变量

有两种方式配置环境变量：

#### 选项1：使用wrangler.jsonc（开发用）

编辑`wrangler.jsonc`文件，填入你的Supabase凭据：

```jsonc
"vars": {
  "SUPABASE_URL": "https://your-project-id.supabase.co",
  "SUPABASE_ANON_KEY": "your-anon-key"
}
```

#### 选项2：使用Cloudflare Secrets（推荐用于生产环境）

```bash
wrangler secret put SUPABASE_URL
# 然后输入你的Supabase URL

wrangler secret put SUPABASE_ANON_KEY
# 然后输入你的Supabase匿名密钥
```

### 4. 本地开发

```bash
npm run dev
```

这将启动本地开发服务器，默认在`http://localhost:8787`。

### 5. 部署到Cloudflare

```bash
npm run deploy
```

## API端点

### 待办事项 API

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/todos` | GET | 获取所有待办事项 |
| `/api/todos` | POST | 创建新的待办事项 |
| `/api/todos/:id` | GET | 获取单个待办事项 |
| `/api/todos/:id` | PUT | 更新待办事项 |
| `/api/todos/:id` | DELETE | 删除待办事项 |

### SSL 证书监控 API

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/certificates` | GET | 获取所有证书信息 |
| `/api/certificates` | POST | 添加新域名 |
| `/api/certificates/:id` | DELETE | 删除域名 |
| `/api/certificates/:id/tags` | PUT | 更新域名标签 |
| `/api/certificates/:id/check` | POST | 手动检查单个证书 |
| `/api/certificates/check-all` | POST | 手动检查所有证书 |
| `/api/tags` | GET | 获取所有可用标签 |

## 请求示例

### 获取所有待办事项

```bash
curl https://your-worker.your-subdomain.workers.dev/api/todos
```

### 创建新的待办事项

```bash
curl -X POST https://your-worker.your-subdomain.workers.dev/api/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "新的待办事项", "completed": false}'
```

### 更新待办事项

```bash
curl -X PUT https://your-worker.your-subdomain.workers.dev/api/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"title": "已更新的待办事项", "completed": true}'
```

### 删除待办事项

```bash
curl -X DELETE https://your-worker.your-subdomain.workers.dev/api/todos/1
```

### 添加域名进行证书监控

```bash
curl -X POST https://your-worker.your-subdomain.workers.dev/api/certificates \
  -H "Content-Type: application/json" \
  -d '{"domain": "example.com", "tags": ["production", "important"]}'
```

### 手动刷新证书

```bash
curl -X POST https://your-worker.your-subdomain.workers.dev/api/certificates/check-all
```

## 🎨 功能截图说明

### 主界面
- 顶部导航栏：在待办事项和证书监控之间切换
- 现代化设计：磨砂玻璃效果、渐变背景、流畅动画

### 待办事项
- 添加、编辑、删除、标记完成
- 简洁优雅的列表展示

### SSL 证书监控
- **Card 布局**：每个域名一张卡片
- **倒计时显示**：清晰展示剩余天数
- **颜色编码**：
  - 🟢 绿色：>30天（安全）
  - 🟡 黄色：7-30天（注意）
  - 🔴 红色：<7天（紧急）
  - ⚫ 灰色：检查失败
- **标签管理**：创建标签、分类域名、按标签筛选
- **自动刷新**：每6小时自动检查所有证书

## 📚 详细文档

更多详细信息请参考：
- [SSL 证书监控完整指南](./SSL_MONITORING_GUIDE.md)

## 🔧 定时任务

项目配置了 Cron Triggers，每 6 小时自动检查所有域名的证书：
```jsonc
"triggers": {
  "crons": ["0 */6 * * *"]
}
```

**注意**：Cron Triggers 仅在部署到 Cloudflare 后生效，本地开发环境不支持。

## 安全注意事项

- 本示例使用了Supabase的匿名密钥，适合演示和个人使用
- 生产环境建议实现适当的身份验证（Supabase Auth）
- 当前版本不做数据隔离，所有访问者共享数据
- 建议配置 Supabase RLS（行级安全策略）限制访问

## 故障排查

### 证书检查失败
- 确认域名已配置 HTTPS
- 验证域名拼写正确
- 稍后重试（第三方 API 可能暂时不可用）

### Cron 任务不执行
- 确认已部署到 Cloudflare（本地开发不支持 Cron）
- 在 Cloudflare 控制台查看 Worker 日志

## 许可证

MIT 

## AnyTLS 节点证书监控（自动同步 + Telegram 告警）

Worker 每次 Cron 运行时先从 logv2fs httpserver 拉取 AnyTLS 节点（`GET /v1/cert-targets`），
写入 `ssl_certificates`（`source='anytls'`，标签 `anytls` + 节点备注），再统一检查证书：

- 连接地址为节点 IP（`connect_host`），TLS ClientHello 的 SNI 为节点 SNI（为空时回退到 domain），与客户端行为一致；
- 唯一键改为 `(domain, port)`，同一域名可在多个端口分别监控；
- 节点从 logv2fs 下线后，对应的 anytls 记录会在下次同步时删除（接口返回空列表时不删除，防误配置）；
- 证书状态变为 `expiring`（≤30 天）/ `expired`（可选 `error`）时推送一次 Telegram，恢复 `valid` 时推送恢复通知；以 `last_alert_status` 去重。

### 部署步骤

1. **Supabase**：在 SQL Editor 执行 `migrations/add_anytls_sync.sql`（可重复执行）。必须先于新 Worker 部署。
2. **logv2fs httpserver**：在 `.env` 中添加 `CERT_TARGETS_TOKEN=<随机长字符串>`（如 `openssl rand -hex 32`），重新编译并重启 `httpserver`。
   未设置该变量时接口返回 503。验证：
   ```bash
   curl -H "X-Cert-Token: $CERT_TARGETS_TOKEN" https://<panel 域名>/v1/cert-targets
   ```
3. **Worker 配置**：把 `wrangler.jsonc` 中 `vars.LOGV2FS_API_URL` 改为 httpserver 根地址（不含 `/v1`），然后设置 secrets：
   ```bash
   npx wrangler secret put SUPABASE_ANON_KEY    # 必需（已从 wrangler.jsonc 明文中移除）
   npx wrangler secret put CERT_TARGETS_TOKEN   # 与 logv2fs .env 一致
   npx wrangler secret put TELEGRAM_BOT_TOKEN   # 可选
   npx wrangler secret put TELEGRAM_CHAT_ID     # 可选
   npx wrangler secret put CF_API_TOKEN         # 可选：Cloudflare API token（Zone Read + SSL and Certificates Read），
                                                # 托管在 Cloudflare 的域名改为读取边缘证书包的到期时间
   npx wrangler secret put ADMIN_TOKEN          # 可选，设置后 /api/* 需 Bearer 鉴权
   ```
   `ALERT_ON_ERROR` 设为 `"true"` 时检查失败也会告警。
4. **部署**：`npm run deploy`。随后可在页面点"检查全部"，或等待下一次 Cron（每 6 小时）触发同步。

### 说明

- 设置 `ADMIN_TOKEN` 后，网页首次请求收到 401 时会弹窗要求输入令牌，并保存在浏览器 localStorage。
- 检查器强制 TLS 1.2 以读取明文证书；若节点配置了 `min_version: "1.3"` 将检查失败。
- Worker 无法 `connect()` 到 Cloudflare 自身 IP，因此托管在 Cloudflare（橙色云 / Worker 路由）的域名无法直连探测。配置 `CF_API_TOKEN` 后，未指定 `connect_host` 的记录会先按域名匹配 token 可见的 zone（最长后缀），读取 `ssl/certificate_packs` 中状态为 active 且覆盖该域名的证书（精确匹配优先于通配符），取最晚到期时间；issuer 带 `(Cloudflare 边缘)` 标记，页面显示 Cloudflare 徽章。API 失败或无匹配证书时回退到直连探测。
- 若 httpserver 前有 Cloudflare 代理，Worker 访问它是普通 HTTPS 请求，不受 `connect()` 的 Cloudflare IP 限制；AnyTLS 节点本身应为 DNS-only 直连。
