# 用户禁用功能说明

## 功能概述

管理员可以通过界面禁用或启用普通用户账户；系统也会在缴费到期（含 7 天宽限期）后自动将用户设为欠费停用。

## 用户状态

| 状态 | 含义 | 触发方式 |
|------|------|----------|
| `plain` | 正常使用 | 注册、启用、续费自动恢复 |
| `disabled` | 管理员手动禁用 | 管理员点击「禁用」 |
| `overdue` | 欠费停用 | 每日 cron：有缴费记录但已超出宽限期 |

历史数据中的 `deleted` 已通过 `migrate-user-status` 迁移为 `disabled`；API 读路径仍兼容 `deleted`。

## API 端点

- `PUT /v1/disableuser/:name` — 手动禁用（`disabled`）
- `PUT /v1/enableuser/:name` — 手动启用（`plain`）

## 缴费到期自动禁用

- **进程**：`httpserver` 每日 cron（默认凌晨 01:00）
- **对象**：`plain` 普通用户，且至少有一条缴费记录
- **豁免**：从未录入缴费的用户不自动禁用
- **宽限**：`end_date` 后 7 个自然日仍可用，第 8 天起设为 `overdue` 并广播各节点 `SafeDisableUser`
- **续费恢复**：录入或更新缴费后，若新区间（含宽限）覆盖今天，则 `overdue` / `disabled` 均自动恢复为 `plain`

### 环境变量

```bash
ENABLE_PAYMENT_EXPIRY_CRON=true      # false 关闭定时任务
PAYMENT_EXPIRY_CRON=0 0 1 * * *      # cron 表达式（6 段，含秒）
PAYMENT_EXPIRY_GRACE_DAYS=7          # 宽限天数
```

### 数据迁移

```bash
./logv2fs migrate-user-status --dry-run=false
```

## 注意事项

1. 禁用/欠费不会删除用户数据，仅变更 `status`
2. 非 `plain` 用户无法获取正常订阅链接
3. 管理员账户不会被自动禁用
4. 节点同步依赖 `SINGBOX_CONTROL_TOKEN` 与各节点 control 端口
