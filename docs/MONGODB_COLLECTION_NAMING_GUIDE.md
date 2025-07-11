# MongoDB 集合名称统一管理指南

## 概述

为了统一管理 MongoDB 集合名称，我们为所有 MongoDB 模型添加了 `CollectionName()` 方法，模仿 PostgreSQL 的 `TableName()` 方法风格。

## 架构设计

### 1. 模型层 (Model Layer)

每个 MongoDB 模型都必须实现 `CollectionName()` 方法：

```go
// PaymentRecord MongoDB版本的缴费记录
type PaymentRecord struct {
    // ... 字段定义 ...
}

// CollectionName 返回MongoDB集合名称
func (PaymentRecord) CollectionName() string {
    return "payment_records"
}
```

### 2. 数据库连接层 (Database Layer)

在 `database/mongodb_connection.go` 中定义了统一的接口和管理函数：

```go
// 定义集合名称接口
type CollectionNamer interface {
    CollectionName() string
}

// GetCollection 获取指定模型的MongoDB集合的便捷方法
func GetCollection(model CollectionNamer) *mongo.Collection {
    return OpenCollectionByModel(Client, model)
}
```

### 3. 控制器层 (Controller Layer)

在控制器中使用新的集合管理方法：

```go
var (
    // 使用新的集合管理方法，从模型中获取集合名称
    paymentRecordsCol *mongo.Collection = database.GetCollection(model.PaymentRecord{})
    userTrafficLogsCol *mongo.Collection = database.GetCollection(model.UserTrafficLogs{})
)
```

## 已实现的模型

### 用户和流量相关
- `UserTrafficLogs` → `USER_TRAFFIC_LOGS`
- `NodeTrafficLogs` → `NODE_TRAFFIC_LOGS`

### 节点管理相关
- `SubscriptionNode` → `subscription_nodes`
- `ExpiryCheckDomainInfo` → `expiry_check_domains`

### 缴费管理相关
- `PaymentRecord` → `payment_records`
- `DailyPaymentAllocation` → `daily_payment_allocations`

## 使用方法

### 1. 在控制器中获取集合

**旧方法（不推荐）：**
```go
paymentRecordsCol := database.OpenCollection(database.Client, "payment_records")
```

**新方法（推荐）：**
```go
paymentRecordsCol := database.GetCollection(model.PaymentRecord{})
```

### 2. 在其他地方动态获取集合

```go
// 直接通过模型获取集合名称
collectionName := model.PaymentRecord{}.CollectionName()

// 或者使用 GetCollection 方法
collection := database.GetCollection(model.PaymentRecord{})
```

## 优势

1. **集中管理**：集合名称在模型中统一定义，避免硬编码散布在代码各处
2. **类型安全**：编译时就能发现集合名称相关的错误
3. **易于维护**：修改集合名称只需要在模型中修改一处
4. **一致性**：与 PostgreSQL 的 `TableName()` 方法保持一致的设计风格
5. **可扩展性**：新增模型时自动获得集合名称管理功能

## 注意事项

1. **向后兼容**：现有的 `OpenCollection` 方法仍然可用，但建议逐步迁移到新方法
2. **命名规范**：MongoDB 集合名称应该与模型的业务含义保持一致
3. **测试**：修改集合名称后需要确保相关的测试用例也相应更新

## 迁移指南

对于现有代码的迁移，建议按以下步骤进行：

1. 为现有的 MongoDB 模型添加 `CollectionName()` 方法
2. 更新控制器中的集合声明，使用 `database.GetCollection()`
3. 逐步替换代码中硬编码的集合名称
4. 更新相关的测试用例

## 示例

完整的使用示例：

```go
// 在模型中定义
type UserTrafficLogs struct {
    // ... 字段定义 ...
}

func (UserTrafficLogs) CollectionName() string {
    return "USER_TRAFFIC_LOGS"
}

// 在控制器中使用
var userTrafficLogsCol = database.GetCollection(model.UserTrafficLogs{})

// 在函数中动态使用
func someFunction() {
    collection := database.GetCollection(model.UserTrafficLogs{})
    // ... 使用 collection 进行数据库操作 ...
}
```

这种设计确保了 MongoDB 集合名称的统一管理，提高了代码的可维护性和一致性。 