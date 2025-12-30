# 🚀 React Query 迁移 - 快速启动指南

## ✅ 重构已完成！

恭喜！您的前端已成功从 Redux rerender 模式迁移到 React Query。

---

## 📦 第一步：安装依赖

依赖已经安装，但如果需要重新安装：

```bash
cd frontend
npm install --legacy-peer-deps
```

---

## 🎯 第二步：启动应用

```bash
cd frontend
npm start
```

---

## ⚠️ 第三步：验证 API 端点（重要！）

打开 `API_ENDPOINTS.md` 文件，验证所有 API 端点是否正确。

特别注意以下端点可能需要调整：
- 节点相关：`7tpxya`, `759b0v`
- 域名相关：`09j2ts`

如需修改，编辑 `src/api/queries.js` 文件即可。

---

## 🧪 第四步：测试功能

### 必测功能清单

#### 1. 用户管理 (`/user`)
- [ ] 用户列表是否正常加载
- [ ] 搜索和筛选是否正常
- [ ] 添加用户是否成功
- [ ] 编辑用户是否成功
- [ ] WebSocket 状态指示器是否显示

#### 2. 我的面板 (`/mypanel`)
- [ ] 用户信息是否正常显示
- [ ] 今日/本月/本年流量是否显示
- [ ] 订阅链接是否可复制
- [ ] WebSocket 连接状态是否显示
- [ ] **重要：流量数据是否还会闪烁？**（应该已修复）

#### 3. 节点管理 (`/nodes`)
- [ ] 节点列表是否正常加载
- [ ] 节点流量统计是否显示
- [ ] 域名监控列表是否加载
- [ ] 添加域名是否成功
- [ ] WebSocket 实时更新是否正常

#### 4. 添加节点 (`/addnode`)
- [ ] 现有节点列表是否显示
- [ ] 添加新节点是否成功
- [ ] 添加后节点列表是否自动刷新

---

## 🔍 常见问题排查

### 问题1：页面一直显示加载中

**原因：** API 端点可能不正确

**解决：**
1. 打开浏览器开发者工具 (F12)
2. 查看 Console 和 Network 标签
3. 查找 404 或 500 错误
4. 根据实际请求的 URL，修改 `src/api/queries.js`

### 问题2：数据不刷新

**原因：** WebSocket 可能未连接

**解决：**
1. 检查页面上的 WebSocket 状态指示器
2. 查看 Console 中的 WebSocket 日志
3. 确保后端 WebSocket 服务正常运行

### 问题3：流量数据仍然闪烁

**原因：** WebSocket 消息过于频繁

**解决：**
1. 检查 `src/service/websocket.js` 中的 `debounceDelay`（当前3秒）
2. 如需调整：增加延迟时间或减少后端推送频率

### 问题4：添加/编辑操作后数据未更新

**原因：** 缓存失效可能未触发

**解决：**
1. 检查 mutation 的 `onSuccess` 回调
2. 确保 WebSocket 正在监听正确的消息类型
3. 手动刷新页面验证数据是否已保存

---

## 🎨 新特性

### 1. 自动加载状态

组件现在自动处理加载状态：

```javascript
const { data: users, isLoading } = useUsers();

if (isLoading) return <Loading />; // 自动显示加载动画
```

### 2. 自动错误处理

错误会自动显示为消息提示：

```javascript
const { error } = useUsers();

useEffect(() => {
    if (error) {
        dispatch(alert({ show: true, content: error.toString() }));
    }
}, [error]);
```

### 3. WebSocket 实时更新

所有页面都会在收到 WebSocket 消息后自动刷新数据：

```javascript
const { status, isConnected } = useWebSocket();
// status: 'connected' | 'connecting' | 'reconnecting' | 'disconnected'
```

### 4. 智能缓存

数据会缓存 5 分钟，减少不必要的 API 请求：

```javascript
// 第一次访问：发送 API 请求
// 5分钟内再次访问：使用缓存数据
// WebSocket 更新：立即刷新缓存
```

---

## 📚 开发指南

### 添加新的 API 查询

1. 在 `src/api/queries.js` 添加查询函数：

```javascript
export const fetchNewData = async (token) => {
    const response = await axios.get(
        `${process.env.REACT_APP_API_HOST}endpoint`,
        { headers: { token } }
    );
    return response.data;
};
```

2. 在 `src/hooks/useQueries.js` 添加 hook：

```javascript
export const useNewData = () => {
    const token = useSelector((state) => state.login.token);
    
    return useQuery({
        queryKey: ['newData'],
        queryFn: () => fetchNewData(token),
        enabled: !!token,
    });
};
```

3. 在组件中使用：

```javascript
import { useNewData } from '../hooks/useQueries';

function MyComponent() {
    const { data, isLoading, error } = useNewData();
    // ...
}
```

### 添加新的 Mutation

1. 在 `src/api/queries.js` 添加 mutation 函数：

```javascript
export const updateData = async ({ data, token }) => {
    const response = await axios.put(
        `${process.env.REACT_APP_API_HOST}endpoint`,
        data,
        { headers: { token } }
    );
    return response.data;
};
```

2. 在 `src/hooks/useQueries.js` 添加 hook：

```javascript
export const useUpdateData = (options = {}) => {
    const token = useSelector((state) => state.login.token);
    const queryClient = useQueryClient();
    
    return useMutation({
        mutationFn: (data) => updateData({ data, token }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['newData'] });
        },
        ...options,
    });
};
```

---

## 🎓 学习资源

- [React Query 官方文档](https://tanstack.com/query/latest)
- [实用的 React Query](https://tkdodo.eu/blog/practical-react-query)
- [React Query 最佳实践](https://tkdodo.eu/blog/react-query-fa-qs)

---

## 📞 需要帮助？

如果遇到任何问题：

1. 检查 `REACT_QUERY_MIGRATION.md` 了解详细的迁移说明
2. 查看 `API_ENDPOINTS.md` 验证 API 端点
3. 查看浏览器 Console 和 Network 标签
4. 检查后端日志

---

## 🎉 享受新的开发体验！

React Query 会让您的数据管理变得更简单、更可靠。不再需要手动管理刷新信号，不再担心数据闪烁问题！

Happy Coding! 🚀
