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