# 流量日志拆表与 hourly_logs 移除

## 概述

历史上 `USER_TRAFFIC_LOGS` / `NODE_TRAFFIC_LOGS` 主文档内嵌了 `hourly_logs / daily_logs / monthly_logs / yearly_logs` 四个数组，造成：

- 文档随时间无限增长，单条 BSON 体积膨胀，热点更新成本高；
- 写入逻辑要 `Find → 遍历 → $push/$inc + arrayFilters`，复杂度高；
- `hourly_logs` 实际早已不写入，仅作为历史包袱占位。

本次将"周期级流量"拆出到独立集合，并彻底移除 `hourly_logs`。

## 新结构

| 集合 | 说明 |
|---|---|
| `USER_TRAFFIC_LOGS` | 用户主文档；保留元数据（含 `used` 累计） |
| `NODE_TRAFFIC_LOGS` | 节点主文档；保留元数据 |
| `user_traffic_periods` | 拆出的用户周期流量 |
| `node_traffic_periods` | 拆出的节点周期流量 |

### 周期表字段

```go
type UserTrafficPeriod struct {
    ID        primitive.ObjectID // _id
    EmailAsId string             // 关联 USER_TRAFFIC_LOGS.email_as_id
    Kind      string             // "daily" | "monthly" | "yearly"
    Period    string             // daily=yyyymmdd, monthly=yyyymm, yearly=yyyy
    Traffic   int64
    UpdatedAt time.Time
}
```

`NodeTrafficPeriod` 结构同理，关联键为 `domain_as_id`。

### 索引

启动时自动确保（见 `database/connection.go:ensureCoreIndexes`）：

- `user_traffic_periods`: `(email_as_id, kind, period)` UNIQUE
- `node_traffic_periods`: `(domain_as_id, kind, period)` UNIQUE

唯一索引是写入端 upsert 幂等性的硬性前提，也是迁移命令"可重复执行不重复"的基础。

## 写入端

`jobs/cron.go` 中的 `LogUserTraffic` / `LogNodeTraffic` 简化为：

1. 主文档：`$inc {used: +N} + $set {updated_at}`（用户）；节点仅刷新 `updated_at`。
2. 周期表：按 daily/monthly/yearly 三次 `UpdateOne(filter, $inc traffic + $set updated_at, upsert=true)`。

封装在 `upsertTrafficPeriod` 工具函数内。

## 读端兼容（前端零改动）

读接口通过 `$lookup` 把周期表的数据重新拼回原 `daily_logs / monthly_logs / yearly_logs` 数组，保持给前端的 JSON 结构不变。

工具函数：

- `controllers/controller.go:lookupUserPeriodStage(kind, periodAlias, asField, limit)`
- `controllers/node.go:lookupNodePeriodStage(kind, periodAlias, asField, limit)`

应用点：

| 入口 | limit |
|---|---|
| `GetAllUsers`（列表） | 每种粒度 10 |
| `GetUserByName`（详情） | 不限 |
| `GetSingboxNodes`（节点列表） | 不限 |

`DeleteUserByUserName` 删除用户时，会同步 `DeleteMany user_traffic_periods{email_as_id}` 清理残留。

## 迁移命令

新增 cobra 子命令 `migrate-traffic-logs`（见 `cmd/migratetrafficlogs/`）。

```bash
# 1. 预演（默认 --dry-run=true，不写库）
./main migrate-traffic-logs

# 2. 真正执行
./main migrate-traffic-logs --dry-run=false
```

行为：

1. 遍历两张主集合，把 `daily_logs / monthly_logs / yearly_logs` 拆出，
   用 `upsert + $setOnInsert` 写入对应周期集合。
2. 真跑模式下追加 `$unset {daily_logs, monthly_logs, yearly_logs, hourly_logs}` 清理老字段。
3. 输出汇总：`新写入 / 已存在跳过 / 清理老字段` 计数。

由于周期表唯一索引的存在，命令**可重复执行**：已存在的 `(owner, kind, period)` 不会被覆盖也不会重复插入。

## 上线步骤

1. **备份**：`mongodump --db logV2rayTrafficDB --out backup-YYYYMMDD`
2. **部署新二进制**（启动自动建索引）
3. **预演**：`./main migrate-traffic-logs`
4. **真跑**：`./main migrate-traffic-logs --dry-run=false`
5. 观察前端面板/节点页流量数据是否正常；下次 cron tick（最长 15 分钟）后会有新数据写入

## 回滚

1. 停止新版进程
2. 恢复主集合：`mongorestore --db logV2rayTrafficDB --drop backup-YYYYMMDD/logV2rayTrafficDB/USER_TRAFFIC_LOGS.bson`（节点同理）
3. Drop 周期集合：

   ```js
   db.user_traffic_periods.drop()
   db.node_traffic_periods.drop()
   ```

4. 切回旧版二进制
