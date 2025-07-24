# IPv6 支持功能

## 概述

本项目已添加完整的IPv6地址支持，确保当节点使用IPv6地址时，生成的订阅URL和配置文件能够正确导入到各种代理客户端中。

## 实现的功能

### 1. IPv6地址检测
- 自动识别IPv4和IPv6地址
- 支持带端口号的地址格式
- 正确处理IPv6链路本地地址（如 `fe80::1%eth0`）

### 2. IPv6地址格式化
- IPv6地址在URL中自动添加方括号包围：`[2001:db8::1]`
- IPv6地址带端口号：`[2001:db8::1]:443`
- IPv4地址保持不变：`192.168.1.1`
- 域名输入不受影响：`example.com`

### 3. 支持的配置格式
- **订阅URL格式**：`vless://uuid@[2001:db8::1]:443?...`
- **Singbox JSON配置**：`"server": "[2001:db8::1]"`
- **Verge YAML配置**：`server: "[2001:db8::1]"`

## 技术实现

### 核心函数

#### `IsIPv6(ip string) bool`
检测字符串是否为IPv6地址：
```go
// 示例
IsIPv6("2001:db8::1")     // 返回 true
IsIPv6("192.168.1.1")     // 返回 false
IsIPv6("example.com")     // 返回 false
```

#### `FormatIPForURL(ip string) string`
格式化IP地址用于URL：
```go
// 示例
FormatIPForURL("2001:db8::1")        // 返回 "[2001:db8::1]"
FormatIPForURL("2001:db8::1:443")    // 返回 "[2001:db8::1]:443"
FormatIPForURL("192.168.1.1")        // 返回 "192.168.1.1"
FormatIPForURL("example.com")        // 返回 "example.com"
```

### 修改的文件

1. **`helpers/utility.go`**
   - 添加了 `IsIPv6()` 和 `FormatIPForURL()` 函数

2. **`controllers/controller.go`**
   - 修改了 `GetSubscripionURL()` 函数
   - 修改了 `ReturnSingboxJson()` 函数
   - 修改了 `ReturnVergeYAML()` 函数

3. **`controllers/config_pg.go`**
   - 修改了 `GetSubscripionURLPG()` 函数
   - 修改了 `ReturnSingboxJsonPG()` 函数
   - 修改了 `ReturnVergeYAMLPG()` 函数

## 支持的节点类型

所有三种节点类型都支持IPv6：

1. **Reality节点**
   - 订阅URL：`vless://uuid@[2001:db8::1]:443?...`
   - JSON配置：`"server": "[2001:db8::1]"`

2. **Hysteria2节点**
   - 订阅URL：`hysteria2://userid@[2001:db8::1]:443?...`
   - JSON配置：`"server": "[2001:db8::1]"`

3. **VlessCDN节点**
   - 订阅URL：`vless://uuid@[2001:db8::1]:443?...`
   - JSON配置：`"server": "[2001:db8::1]"`

## 兼容性

### 向后兼容
- IPv4地址的处理完全保持不变
- 域名输入不受影响
- 现有配置无需修改

### 客户端兼容性
- **Sing-box**：完全支持IPv6地址
- **Clash Verge**：完全支持IPv6地址
- **其他支持Sing-box协议的客户端**：应该都能正常工作

## 测试用例

### IPv4地址
- `192.168.1.1` → `192.168.1.1`
- `192.168.1.1:443` → `192.168.1.1:443`

### IPv6地址
- `2001:db8::1` → `[2001:db8::1]`
- `2001:db8::1:443` → `[2001:db8::1]:443`
- `::1` → `[::1]`
- `fe80::1%eth0` → `[fe80::1%eth0]`

### 域名
- `example.com` → `example.com`
- `example.com:443` → `example.com:443`

## 使用示例

### 添加IPv6节点
在管理界面添加节点时，可以直接输入IPv6地址：
- IP地址：`2001:db8::1`
- 端口：`443`
- 类型：`reality`

### 生成的订阅URL
```
vless://uuid@[2001:db8::1]:443?encryption=none&flow=xtls-rprx-vision&security=reality&sni=itunes.apple.com&fp=chrome&pbk=public_key&sid=short_id&type=tcp&headerType=none#节点备注
```

### 生成的JSON配置
```json
{
  "tag": "节点备注",
  "type": "vless",
  "server": "[2001:db8::1]",
  "server_port": 443,
  "uuid": "uuid",
  "flow": "xtls-rprx-vision",
  "packet_encoding": "xudp"
}
```

## 注意事项

1. **IPv6地址格式**：确保输入的IPv6地址格式正确
2. **网络环境**：确保客户端和服务器都支持IPv6连接
3. **防火墙设置**：确保IPv6端口在防火墙中开放
4. **DNS解析**：如果使用域名，确保DNS能够解析到IPv6地址

## 故障排除

### 常见问题

1. **IPv6地址无法连接**
   - 检查网络是否支持IPv6
   - 检查防火墙设置
   - 验证IPv6地址格式

2. **订阅导入失败**
   - 确保客户端支持IPv6
   - 检查订阅URL格式是否正确
   - 验证节点配置是否完整

3. **配置生成错误**
   - 检查IP地址格式
   - 确保端口号正确
   - 验证UUID格式

## 更新日志

- **2024年**：添加完整的IPv6支持
  - 实现IPv6地址检测和格式化
  - 支持所有节点类型的IPv6配置
  - 确保向后兼容性 