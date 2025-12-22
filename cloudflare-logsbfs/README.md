# Cloudflare Worker与Supabase PostgreSQL集成

这是一个Cloudflare Worker项目，展示了如何将Supabase PostgreSQL集成到Cloudflare Worker中，创建一个简单的待办事项(Todo)API。

## 功能特点

- 使用Cloudflare Worker作为无服务器后端
- 使用Supabase作为PostgreSQL数据库服务
- 完整的CRUD操作API
- 行级安全性(RLS)支持
- CORS支持

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
3. 运行`migrations/create_todos_table.sql`中的SQL语句，创建所需的表和策略

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

| 端点 | 方法 | 描述 |
|------|------|------|
| `/` | GET | 返回欢迎信息 |
| `/api/todos` | GET | 获取所有待办事项 |
| `/api/todos` | POST | 创建新的待办事项 |
| `/api/todos/:id` | GET | 获取单个待办事项 |
| `/api/todos/:id` | PUT | 更新待办事项 |
| `/api/todos/:id` | DELETE | 删除待办事项 |

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

## 安全注意事项

- 本示例使用了Supabase的匿名密钥，这适合公共读取操作，但在生产环境中，你可能需要实现适当的身份验证
- 示例中的RLS策略允许任何人访问数据，在实际应用中，你应该根据用户ID限制访问

## 许可证

MIT 