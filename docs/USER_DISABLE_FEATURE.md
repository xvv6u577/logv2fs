# 用户禁用功能说明

## 功能概述

管理员可以通过界面禁用或启用普通用户账户。计费/续费已从 logv2fs 剥离，本系统不再根据缴费自动改状态。

## 用户状态

| 状态 | 含义 | 触发方式 |
|------|------|----------|
| `plain` | 正常使用 | 注册、管理员启用 |
| `disabled` | 管理员手动禁用 | 管理员点击「禁用」 |
| `overdue` | 历史遗留「欠费停用」 | **不再写入**；清理见 `docs/REMOVE_PAYMENT_RECORDS.md`，用 `migrate-overdue-status` 迁为 `disabled` |

历史数据中的 `deleted` 已通过 `migrate-user-status` 迁移为 `disabled`；API 读路径仍兼容 `deleted`（归一成 `disabled`）。

## API 端点

- `PUT /v1/disableuser/:name` — 手动禁用（`disabled`）
- `PUT /v1/enableuser/:name` — 手动启用（`plain`）

## 注意事项

1. 禁用不会删除用户数据，仅变更 `status`
2. 非 `plain` 用户无法获取正常订阅链接
3. 管理员账户不会被手动禁用
4. 节点同步依赖 `SINGBOX_CONTROL_TOKEN` 与各节点 control 端口
