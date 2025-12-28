# 节点权重排序功能文档

## 📋 功能概述

为 `subscription_nodes` 表的每个节点添加了 **权重（Weight）** 字段，用于在生成订阅文件时对节点进行排序。

### 核心特性

- **权重类型**：整数（支持正数、负数和零）
- **排序规则**：数值越小越靠前（升序排列）
- **默认值**：0
- **应用范围**：影响所有订阅生成函数

## 🎯 影响的功能

权重排序在以下三个订阅生成函数中生效：

1. **GetSubscriptionURL()** / **GetSubscriptionURLPG()** 
   - 生成 vless/hysteria2 订阅链接（Base64编码）

2. **ReturnSingboxJson()** / **ReturnSingboxJsonPG()**
   - 生成 Singbox JSON 配置文件

3. **ReturnVergeYAML()** / **ReturnVergeYAMLPG()**
   - 生成 Clash/Verge YAML 配置文件

## 🔧 技术实现

### 1. 数据模型修改

#### MongoDB 版本 (`model/node.go`)
```go
type SubscriptionNode struct {
    // ... 其他字段
    Weight int `json:"weight" bson:"weight"` // 权重字段
}
```

#### PostgreSQL 版本 (`model/postgres_models.go`)
```go
type SubscriptionNodePG struct {
    // ... 其他字段
    Weight int `json:"weight" gorm:"default:0;index"` // 带索引的权重字段
}
```

### 2. 数据库迁移

#### PostgreSQL 迁移文件
位置：`database/migration_add_weight_field.sql`

执行迁移：
```bash
# 连接到 PostgreSQL 数据库
psql -U your_username -d your_database -f database/migration_add_weight_field.sql
```

迁移脚本会：
- 添加 `weight` 字段（默认值为 0）
- 创建索引以优化排序查询
- 为现有节点按创建时间分配初始权重

#### MongoDB
MongoDB 是无模式数据库，无需迁移脚本。已有文档会自动使用 weight 的零值（0）。

### 3. 查询排序

#### MongoDB 查询示例
```go
cur, err := mongodb.GetCollection(model.SubscriptionNode{}).Find(
    context.TODO(), 
    bson.D{},
    options.Find().SetSort(bson.D{{Key: "weight", Value: 1}}), // 升序排序
)
```

#### PostgreSQL 查询示例
```go
db.Where("type != ?", "work").
   Order("weight ASC").  // 升序排序
   Find(&activeGlobalNodes)
```

### 4. 前端界面

位置：`frontend/src/components/addNode.js`

新增功能：
- ✅ 权重输入框（支持负数）
- ✅ 在节点卡片中显示权重
- ✅ 复制节点时包含权重信息
- ✅ 友好的提示文字

## 📝 使用示例

### 权重设置策略

#### 示例 1：基础排序
```
节点A: weight = 0   → 排第2位
节点B: weight = -10 → 排第1位（最前）
节点C: weight = 10  → 排第3位
节点D: weight = 10  → 排第4位（相同权重按原顺序）
```

#### 示例 2：分组排序
```
# 高优先级节点（总是在前面）
香港节点1: weight = -100
香港节点2: weight = -99

# 中优先级节点
美国节点1: weight = 0
美国节点2: weight = 1

# 低优先级节点（备用）
备用节点1: weight = 100
备用节点2: weight = 101
```

#### 示例 3：按地区分组
```
# 亚洲节点（-50 ~ -1）
日本: weight = -50
韩国: weight = -49
新加坡: weight = -48

# 欧洲节点（0 ~ 49）
英国: weight = 0
德国: weight = 1

# 美洲节点（50 ~ 99）
美国: weight = 50
加拿大: weight = 51
```

## 🎨 前端操作指南

### 添加新节点

1. 打开"添加节点"页面
2. 填写基本信息（节点类型、备注、域名等）
3. 在"高级选项"中找到"权重 (Weight)"字段
4. 输入权重值：
   - 负数：高优先级（排在前面）
   - 0：默认优先级
   - 正数：低优先级（排在后面）
5. 点击"添加到列表"
6. 点击"提交所有节点"保存

### 修改现有节点权重

1. 在节点列表中找到目标节点
2. 点击"复制到表单"按钮
3. 修改权重值
4. 删除原节点
5. 添加修改后的节点到列表
6. 提交保存

## 📊 数据库字段详情

### PostgreSQL

| 字段名 | 类型 | 默认值 | 索引 | 说明 |
|--------|------|--------|------|------|
| weight | INTEGER | 0 | ✅ | 节点权重，数值越小越靠前 |

### MongoDB

| 字段名 | 类型 | 默认值 | 索引 | 说明 |
|--------|------|--------|------|------|
| weight | int | 0 | 可选 | 节点权重，数值越小越靠前 |

## ⚡ 性能优化

### PostgreSQL
- ✅ 已为 `weight` 字段创建索引
- ✅ 排序查询性能优化
- ✅ 支持快速范围查询

### MongoDB
可选择为 weight 字段创建索引：
```javascript
db.subscription_nodes.createIndex({ weight: 1 })
```

## 🔄 向后兼容性

- ✅ 已有节点自动使用默认权重 0
- ✅ 不影响现有功能
- ✅ 旧的订阅链接继续工作
- ✅ API 接口完全兼容

## 🐛 故障排查

### 问题 1：PostgreSQL 迁移失败
```bash
# 检查字段是否已存在
SELECT column_name FROM information_schema.columns 
WHERE table_name = 'subscription_nodes' AND column_name = 'weight';

# 如果字段已存在，跳过迁移或手动添加索引
CREATE INDEX IF NOT EXISTS idx_subscription_nodes_weight 
ON subscription_nodes (weight);
```

### 问题 2：节点排序不生效
1. 确认数据库中 weight 字段已正确设置
2. 检查查询代码是否包含排序语句
3. 清除客户端缓存，重新获取订阅

### 问题 3：前端无法输入负数
- 确保 input 类型为 `number`
- 检查浏览器控制台是否有 JavaScript 错误

## 📚 相关文件

### 后端
- `model/node.go` - MongoDB 数据模型
- `model/postgres_models.go` - PostgreSQL 数据模型
- `controllers/mongodb/controller.go` - MongoDB 控制器
- `controllers/postgres/config.go` - PostgreSQL 配置控制器
- `database/migration_add_weight_field.sql` - PostgreSQL 迁移脚本

### 前端
- `frontend/src/components/addNode.js` - 节点管理组件

## 🎯 最佳实践

1. **权重分配**
   - 使用10的倍数作为基础权重（便于插入新节点）
   - 为不同地区/类型的节点分配不同的权重范围

2. **维护建议**
   - 定期审查节点权重，确保排序合理
   - 为新增节点设置适当的初始权重
   - 使用负数表示"永远排在前面"的节点

3. **测试建议**
   - 添加节点后，立即测试订阅链接
   - 验证节点在客户端中的显示顺序
   - 测试极端权重值（如 -1000、1000）

## 🔐 安全注意事项

- 权重值仅影响节点排序，不影响节点可用性
- 权重范围建议限制在 -1000 到 1000 之间
- 管理员权限才能修改节点权重

## 📅 版本历史

- **v1.0.0** (2025-12-28)
  - ✨ 初始实现
  - 支持 MongoDB 和 PostgreSQL
  - 前端界面完善
  - 包含数据库迁移脚本

## 🤝 贡献指南

如需扩展此功能：
1. Fork 本仓库
2. 创建特性分支
3. 提交 Pull Request
4. 更新相关文档

---

**作者**: logv2fs 团队  
**最后更新**: 2025-12-28  
**许可证**: MIT
