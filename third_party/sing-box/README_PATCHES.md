# sing-box 本地 fork patch 说明

本目录基于 sing-box v1.12.25，仅追加了少量“运行时增/删用户”导出 API。
项目根 `go.mod` 通过 `replace github.com/sagernet/sing-box => ./third_party/sing-box`
让所有依赖走这份 fork。

将来升级 sing-box 时按下方清单逐个 rebase 即可，所有改动均为 additive（没有删除任何
原有逻辑），以最大化减少冲突。每段改动都用 "patch by logv2fs" 注释做了标记。

## 改动清单

| 文件 | 改动内容 |
|---|---|
| `box.go` | 保存并导出 `Box.V2RayServer()`，供业务侧访问流量统计服务 |
| `experimental/v2rayapi/stats.go` | 在 `*StatsService` 上实现 `AddUser` / `RemoveUser` |
| `protocol/vless/inbound.go` | 加读写锁，并新增 `AddUser` / `RemoveUser` / 用户表重建逻辑（append-only + 墓碑索引） |
| `protocol/vmess/inbound.go` | 同 VLESS，对应 `option.VMessUser`（AlterId） |
| `protocol/trojan/inbound.go` | 同 VLESS，墓碑用 `Password == ""` 标记 |
| `protocol/hysteria2/inbound.go` | 保存用户配置、保护用户名读取，并新增运行时增删与用户表重建逻辑 |
| `protocol/anytls/inbound.go` | 保存用户表；因 sing-anytls 无 UpdateUsers，AddUser/RemoveUser 通过重建 Service 生效（墓碑清空 Password） |

## 设计原则

1. **append-only + 墓碑**：所有 inbound 的 `users` 切片永远只增长，删除时把认证关键
   字段（UUID/Password）清空。索引稳定，不会让已建立连接的 `userIndex` 越界或指向
   错位用户。代价是切片会随频繁增删慢慢膨胀，下次进程重启会被 `UpdateOptionsFromMongoDB`
   全量重建清零，业务量下可接受。

2. **接口最小化扩展**：仅在 `V2RayStatsService` 上加 `AddUser/RemoveUser`，没有给
   `adapter.Inbound` 加新接口（避免影响所有现有实现）。业务侧通过类型断言到具体类型
   （`*inbound.VLESS` 等）来调 AddUser/RemoveUser。

3. **最小侵入原则**：协议行为沿用上游实现，本地修改只覆盖运行时用户管理和统计服务访问。

## 验证

```bash
go test -tags="with_quic with_v2ray_api" ./singbox ./cmd/...
```
