# 移除缴费集合（运维手册）

计费已从 logv2fs 剥离到独立系统。应用代码不再读写 `payment_records`。
本手册描述**线上数据库**的一次性清理步骤。请先备份，再 drop，再迁移用户状态。

**本仓库命令不会自动 drop 集合。**

## 前置

- 已部署不含缴费 API 的新版本后端（`controllers/payment.go` 等已删除）。
- 有 `mongodump` / `mongosh`（或 Compass）访问 `logV2rayTrafficDB` 的权限。
- 数据库名：`logV2rayTrafficDB`；集合名：`payment_records`。

## 1. 备份

```bash
# 仅导出缴费集合
mongodump --uri="$MONGO_URI" \
  --db=logV2rayTrafficDB \
  --collection=payment_records \
  --out=./backup-payment_records-$(date +%Y%m%d)

# 可选：再确认条数
mongosh "$MONGO_URI" --quiet --eval \
  'db.getSiblingDB("logV2rayTrafficDB").payment_records.countDocuments()'
```

确认 `backup-.../logV2rayTrafficDB/payment_records.bson` 存在且大小合理后再继续。

## 2. 删除集合

```bash
mongosh "$MONGO_URI" --eval \
  'db.getSiblingDB("logV2rayTrafficDB").payment_records.drop()'
```

索引随集合一起消失，无需单独删。

## 3. 迁移 overdue → disabled

历史上若有用户 `status=overdue`（欠费停用），统一改为 `disabled`（与管理员手动禁用一致）。

在已配置好 `.env`（含 Mongo 连接）的机器上：

```bash
# 先 dry-run 看数量
./logv2fs migrate-overdue-status

# 确认后真正写库
./logv2fs migrate-overdue-status --dry-run=false
```

该命令只更新 `USER_TRAFFIC_LOGS` 的 `status` 与 `updated_at`，**不**向各节点广播启停。
若迁移后仍有用户在节点侧处于禁用态，属预期（原 overdue 本就应不可用）；需要重新开通时用管理后台「启用」。

## 4. 校验

```bash
mongosh "$MONGO_URI" --quiet --eval '
  const dbn = db.getSiblingDB("logV2rayTrafficDB");
  print("payment_records exists:", dbn.getCollectionNames().includes("payment_records"));
  print("overdue left:", dbn.USER_TRAFFIC_LOGS.countDocuments({status:"overdue"}));
'
```

期望：集合不存在；`overdue` 为 0。
