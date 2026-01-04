# API 端点映射参考

这个文档列出了所有在 `src/api/queries.js` 中使用的 API 端点。

## ⚠️ 重要：验证端点正确性

由于我是根据代码推断创建的端点，**请验证以下端点是否与后端API匹配**。

## 📋 当前使用的端点

### 用户相关
| 功能 | 方法 | 端点 | 文件位置 |
|------|------|------|----------|
| 获取所有用户 | GET | `n778cf` | `fetchUsers` |
| 获取当前用户 | GET | `user/{email}` | `fetchCurrentUser` |
| 添加用户 | POST | `signup` | `addUser` |
| 更新用户 | PUT | `edit` | `updateUser` |
| 删除用户 | DELETE | `deluser` | `deleteUser` |

### 节点相关
| 功能 | 方法 | 端点 | 文件位置 |
|------|------|------|----------|
| 获取所有节点 | GET | `c47kr8` | `fetchNodes` |
| 更新节点 | PUT | `759b0v` | `updateNodes` |

## 🔍 需要验证的端点

### 从旧代码中发现的不同端点

如果在旧代码中发现了以下端点，可能需要更新 `src/api/queries.js`：

```javascript
// 旧代码中可能使用的端点
"c47kr8"  // 可能是获取节点的另一个端点？
"t7k033"  // 可能是另一个节点相关端点？
```

## 🔧 如何验证

### 方法 1：查看后端路由定义

查看后端代码中的路由定义，例如：
- `routers/postgres/router.go` 或
- `routers/mongodb/router.go`

### 方法 2：检查浏览器网络请求

1. 打开浏览器开发者工具
2. 切换到 Network 标签
3. 在应用中执行操作
4. 查看实际发送的 API 请求

### 方法 3：搜索旧代码中的端点

```bash
# 在项目根目录执行
grep -r "7tpxya\|09j2ts\|759b0v" frontend/src/components/
```

## 🛠️ 如何修改端点

如果发现端点不正确，修改 `src/api/queries.js` 文件：

```javascript
// 例如，如果节点端点应该是 "c47kr8" 而不是 "7tpxya"
export const fetchNodes = async (token) => {
    const response = await axios.get(
        `${process.env.REACT_APP_API_HOST}c47kr8`,  // 修改这里
        { headers: { token } }
    );
    return response.data;
};
```

修改后，React Query 会自动使用新的端点，无需修改组件代码。

## ✅ 验证清单

- [ ] 用户列表加载正常
- [ ] 当前用户信息显示正常
- [ ] 添加用户功能正常
- [ ] 节点列表加载正常
- [ ] 添加节点功能正常
- [ ] WebSocket 实时更新正常
- [ ] 流量数据显示正常（无闪烁）

## 📝 备注

如果发现任何端点问题，请：
1. 修改 `src/api/queries.js` 中的端点
2. 刷新浏览器
3. React Query 会自动使用新端点
4. 更新此文档记录正确的端点
