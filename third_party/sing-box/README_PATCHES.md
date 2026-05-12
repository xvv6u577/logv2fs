# sing-box 本地 fork patch 说明

本目录是 sing-box v1.8.1 的原样拷贝，仅追加了少量"运行时增/删用户"导出 API。
项目根 `go.mod` 通过 `replace github.com/sagernet/sing-box => ./third_party/sing-box`
让所有依赖走这份 fork。

将来升级 sing-box 时按下方清单逐个 rebase 即可，所有改动均为 additive（没有删除任何
原有逻辑），以最大化减少冲突。每段改动都用 "patch by logv2fs" 注释做了标记。

## 改动清单

| 文件 | 改动内容 |
|---|---|
| `box.go` | 新增 `Box.Inbounds()` 与 `Box.Inbound(tag)` 方法 |
| `adapter/experimental.go` | 在 `V2RayStatsService` 接口上追加 `AddUser(name)` / `RemoveUser(name)` |
| `experimental/v2rayapi/stats.go` | 在 `*StatsService` 上实现 `AddUser` / `RemoveUser` |
| `inbound/vless.go` | 加 `usersAccess sync.RWMutex`；为 `newConnection/newPacketConnection` 中的 `h.users[i]` 读取加锁；新增 `AddUser` / `RemoveUser` / `rebuildServiceLocked`（append-only + 墓碑索引） |
| `inbound/vmess.go` | 同 VLESS，对应 `option.VMessUser`（AlterId） |
| `inbound/trojan.go` | 同 VLESS，墓碑用 `Password == ""` 标记 |
| `inbound/hysteria2.go` | 增加 `users []option.Hysteria2User` 与 `usersAccess sync.RWMutex`；初始化时把 `options.Users` 复制一份保留；newConnection/newPacketConnection 对 `userNameList` 读取加锁；新增 `AddUser` / `RemoveUser` / `rebuildServiceLocked` |

## 设计原则

1. **append-only + 墓碑**：所有 inbound 的 `users` 切片永远只增长，删除时把认证关键
   字段（UUID/Password）清空。索引稳定，不会让已建立连接的 `userIndex` 越界或指向
   错位用户。代价是切片会随频繁增删慢慢膨胀，下次进程重启会被 `UpdateOptionsFromMongoDB`
   全量重建清零，业务量下可接受。

2. **接口最小化扩展**：仅在 `V2RayStatsService` 上加 `AddUser/RemoveUser`，没有给
   `adapter.Inbound` 加新接口（避免影响所有现有实现）。业务侧通过类型断言到具体类型
   （`*inbound.VLESS` 等）来调 AddUser/RemoveUser。

3. **零删除原则**：原始 sing-box 代码全部保留，只增不改不删。这样下次跟随上游升级时
   只需要把 patch 里的新增行重新 apply 到新版本即可。

## 验证

```bash
go build -tags="with_gvisor,with_quic,with_wireguard,with_utls,with_reality_server,with_clash_api,with_v2ray_api,with_grpc" ./...
```
