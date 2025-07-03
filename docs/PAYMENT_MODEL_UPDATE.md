# 费用管理模型更新说明

## 更新时间
2024年

## 更新内容

### 1. 数据模型变更

#### 删除的字段
- `payment_method` (支付方式) - 不再记录支付方式
- `payment_date` (缴费日期) - 替换为服务期间

#### 新增的字段
- `start_date` (服务开始日期) - VPN服务的开始日期
- `end_date` (服务结束日期) - VPN服务的结束日期

### 2. 功能变更

#### 费用录入界面 (Add Payment)
- **删除**：支付方式选择
- **新增**：服务开始日期和结束日期选择
- **新增**：自动计算并显示服务天数
- **优化**：默认设置结束日期为开始日期后30天

#### 费用统计界面 (Payment Stats)
- **删除**：支付方式分类统计
- **保留**：日/月/年统计功能
- **优化**：统计基于服务开始日期

#### 用户缴费记录显示
- **更新**：显示服务期间（开始日期 ~ 结束日期）而不是单个缴费日期
- **新增**：显示服务天数
- **删除**：支付方式显示

### 3. API接口变更

#### POST /v1/payment
旧格式：
```json
{
  "user_email_as_id": "user@example.com",
  "amount": 100,
  "payment_method": "alipay",
  "payment_date": "2024-01-01",
  "remark": "月度缴费"
}
```

新格式：
```json
{
  "user_email_as_id": "user@example.com",
  "amount": 100,
  "start_date": "2024-01-01",
  "end_date": "2024-01-31",
  "remark": "月度续费"
}
```

### 4. 数据库变更

#### MongoDB索引
- 删除：`payment_date`、`payment_method` 索引
- 新增：`start_date`、`end_date` 索引

#### PostgreSQL索引
- 删除：`idx_payment_records_payment_date`、`idx_payment_records_payment_method`
- 新增：`idx_payment_records_start_date`、`idx_payment_records_end_date`

### 5. 迁移注意事项

如果系统中已有旧的缴费记录，需要进行数据迁移：

1. **备份现有数据**
2. **运行新的迁移命令**：`./logv2fs migrate payment`
3. **手动迁移旧数据**（如需要）：
   - 将 `payment_date` 作为 `start_date`
   - 根据业务规则设置 `end_date`（如：start_date + 30天）
   - 删除 `payment_method` 字段

### 6. 业务逻辑说明

- **服务期限**：每条缴费记录现在代表一段VPN服务期限
- **续费管理**：管理员录入续费时需要明确指定服务的起止日期
- **统计逻辑**：所有统计基于服务开始日期（start_date）进行分组

### 7. 优势

1. **更精确的服务管理**：明确每次缴费对应的服务期限
2. **更好的用户体验**：用户可以清楚看到自己的服务有效期
3. **简化数据模型**：移除了不必要的支付方式字段
4. **更合理的业务逻辑**：符合VPN服务按时间计费的特点 