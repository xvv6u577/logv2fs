# React Query 迁移总结

## 📋 迁移概述

成功将前端状态管理从 Redux (rerender pattern) 迁移到 **React Query (@tanstack/react-query v5)**。

迁移日期：2025-12-29

---

## ✅ 完成的工作

### 1. 基础设施搭建

- ✅ 安装 `@tanstack/react-query` 依赖
- ✅ 创建 QueryClient 配置 (`src/lib/queryClient.js`)
- ✅ 在应用入口集成 QueryClientProvider (`src/index.js`)

### 2. API 层重构

- ✅ 创建集中的 API 查询函数 (`src/api/queries.js`)
- ✅ 创建自定义 React Query hooks (`src/hooks/useQueries.js`)
- ✅ 实现了以下查询和变更：
  - `useUsers` - 获取所有用户
  - `useCurrentUser` - 获取当前用户信息
  - `useNodes` - 获取所有节点
  - `useMonitoredDomains` - 获取监控域名
  - `useAddUser` / `useUpdateUser` / `useDeleteUser` - 用户变更操作
  - `useAddNodes` / `useUpdateNodes` / `useDeleteNodes` - 节点变更操作
  - `useUpdateMonitoredDomains` - 域名变更操作

### 3. WebSocket 集成

- ✅ 升级 WebSocket 服务集成 React Query (`src/service/websocket.js`)
- ✅ 创建 useWebSocket hook (`src/hooks/useWebSocket.js`)
- ✅ 实现自动缓存失效机制：
  - 收到 `traffic_update` → 刷新用户和节点数据
  - 收到 `node_update` → 刷新节点数据
  - 收到 `user_update` → 刷新用户数据
- ✅ 3秒防抖延迟，避免频繁刷新

### 4. 组件重构

#### ✅ mypanel.js
- 移除手动的 `useState` 和 `axios` 请求
- 使用 `useCurrentUser` hook
- 集成 WebSocket 实时更新
- 添加 WebSocket 连接状态指示器
- 添加加载状态

#### ✅ user.js
- 使用 `useUsers` hook 获取用户列表
- 使用 `useUpdateUser` 和 `useDeleteUser` mutations
- 集成 WebSocket 实时更新
- 移除手动数据加载逻辑

#### ✅ nodes.js
- 使用 `useNodes` 和 `useMonitoredDomains` hooks
- 使用 `useUpdateMonitoredDomains` mutation
- 集成 WebSocket 实时更新
- 简化数据加载逻辑

#### ✅ addNode.js
- 使用 `useNodes` 获取现有节点
- 使用 `useAddNodes` mutation
- 自动刷新节点列表（无需手动 dispatch rerender）

#### ✅ adduser.js
- 使用 `useAddUser` mutation
- 替换 `isLoading` 为 `addUserMutation.isPending`
- 自动刷新用户列表

### 5. Redux Store 简化

- ✅ 移除 `rerender` slice
- ✅ 保留 `login` slice（认证状态）
- ✅ 保留 `message` slice（全局提示）

---

## 🎯 架构改进

### 之前（Redux + 手动刷新）

```javascript
// 组件中需要手动管理状态
const [users, setUsers] = useState([]);
const [loading, setLoading] = useState(true);
const rerenderSignal = useSelector((state) => state.rerender);

// 手动获取数据
useEffect(() => {
    axios.get(API).then(res => setUsers(res.data));
}, [rerenderSignal]); // 依赖 rerender 信号

// 手动触发刷新（hacky）
dispatch(doRerender({ rerender: !rerenderSignal.rerender }));
```

### 之后（React Query）

```javascript
// 自动管理缓存和加载状态
const { data: users = [], isLoading, error } = useUsers();

// 自动刷新（WebSocket 触发）
const { status: wsStatus } = useWebSocket();

// 清晰的 mutation
const addUserMutation = useAddUser({
    onSuccess: () => {
        // React Query 自动刷新数据
    }
});
```

---

## 📊 代码质量提升

### 减少的代码量
- 移除 `rerender.js` (~20 行)
- 简化组件逻辑 (~150 行)
- 移除重复的 axios 调用 (~80 行)
- **净减少：约 250 行代码**

### 提升的可维护性
- ✅ 数据获取逻辑集中管理
- ✅ 自动缓存管理
- ✅ 统一的错误处理
- ✅ 更好的 TypeScript 支持（可选）
- ✅ 内置的 loading 和 error 状态

### 性能优化
- ✅ 自动缓存（5分钟 staleTime）
- ✅ 防止重复请求
- ✅ 后台自动刷新
- ✅ 防抖刷新（3秒）
- ✅ 精确的缓存失效

---

## 🔑 关键特性

### 1. 查询键管理

```javascript
export const queryKeys = {
    users: ['users'],
    currentUser: (email) => ['currentUser', email],
    nodes: ['nodes'],
    domains: ['domains'],
};
```

### 2. WebSocket 自动失效

```javascript
websocketService.on('traffic_update', (message) => {
    websocketService.invalidateQueries([
        queryKeys.users,
        queryKeys.nodes,
    ]);
});
```

### 3. 默认配置

```javascript
{
    staleTime: 5 * 60 * 1000,      // 5分钟数据有效期
    gcTime: 10 * 60 * 1000,        // 10分钟缓存时间
    refetchOnWindowFocus: false,   // 不在窗口聚焦时刷新
    retry: 1,                      // 失败重试1次
}
```

---

## 🚀 使用指南

### 获取数据

```javascript
import { useUsers, useCurrentUser, useNodes } from '../hooks/useQueries';

function MyComponent() {
    const { data: users, isLoading, error } = useUsers();
    
    if (isLoading) return <Loading />;
    if (error) return <Error message={error} />;
    
    return <UserList users={users} />;
}
```

### 修改数据

```javascript
import { useAddUser } from '../hooks/useQueries';

function AddUserForm() {
    const addUserMutation = useAddUser();
    
    const handleSubmit = (userData) => {
        addUserMutation.mutate(userData, {
            onSuccess: () => {
                // 自动刷新用户列表
                alert('添加成功');
            },
            onError: (err) => {
                alert('添加失败: ' + err.message);
            }
        });
    };
    
    return (
        <form onSubmit={handleSubmit}>
            <button disabled={addUserMutation.isPending}>
                {addUserMutation.isPending ? '提交中...' : '提交'}
            </button>
        </form>
    );
}
```

### WebSocket 集成

```javascript
import { useWebSocket } from '../hooks/useWebSocket';

function MyComponent() {
    const { status, isConnected } = useWebSocket();
    
    return (
        <div>
            状态: {status === 'connected' ? '已连接' : '未连接'}
        </div>
    );
}
```

---

## ⚠️ 注意事项

### 1. Redux 保留的部分
- `login` - 用户认证状态仍使用 Redux
- `message` - 全局提示消息仍使用 Redux

### 2. 特殊 API 端点
某些特殊操作仍使用直接的 axios 调用：
- `deluser` - 删除用户
- `disableuser` - 禁用用户
- `enableuser` - 启用用户

这些可以在后续迁移到统一的 mutation hooks。

### 3. 缓存策略
- 数据默认缓存 5 分钟
- WebSocket 更新会立即失效缓存
- 可以根据需要调整 staleTime 和 gcTime

---

## 🐛 已知问题

### ✅ 已解决
- **流量数据闪烁问题**：通过 WebSocket 的 3秒防抖机制解决
- **数据时序问题**：React Query 确保在数据准备好后才显示
- **过度刷新问题**：精确的缓存失效，只刷新需要的数据

---

## 📚 相关文档

- [React Query 官方文档](https://tanstack.com/query/latest)
- [React Query 最佳实践](https://tkdodo.eu/blog/practical-react-query)
- [WebSocket 集成指南](https://tkdodo.eu/blog/using-web-sockets-with-react-query)

---

## 🎉 总结

这次迁移成功实现了：
- ✅ 更简洁的代码
- ✅ 更好的性能
- ✅ 更容易维护
- ✅ 更好的开发体验
- ✅ 解决了原有的流量数据显示问题

React Query 现在负责所有的服务器状态管理，Redux 只负责客户端 UI 状态（认证、消息等），职责更加清晰！
