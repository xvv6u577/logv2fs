# Cloudflare IP 扫描器

一个用于扫描 Cloudflare IP 段、检测可用性并识别地理位置的 Node.js 工具。

## 功能特点

- ✅ 支持 CIDR 格式的 IP 段和单个 IP 地址混合输入
- ✅ 自动将 IP 转换为 16 进制格式
- ✅ 通过 Cloudflare CDN trace 接口验证 IP 可用性
- ✅ 每个 IP 测试 3 次，计算平均响应时间
- ✅ 自动提取 IP 所属国家/地区
- ✅ 支持高并发扫描（可配置，默认 64 个并发）
- ✅ 实时进度条显示（包含平均延迟）
- ✅ 响应时间统计（平均/最快/最慢）
- ✅ 自动超时处理（5 秒）
- ✅ 结果输出为 JSON 格式

## 安装

### 前置要求

- Node.js 18.0.0 或更高版本

### 安装依赖

```bash
npm install
```

## 使用方法

### 1. 准备 IP 段文件

创建一个文本文件（如 `ips.txt`），每行一个 CIDR 格式的 IP 段或单个 IP 地址：

```text
# Cloudflare IP 段（CIDR 格式）
104.16.0.0/24
104.17.0.0/24

# 单个 IP 地址
8.219.217.25
144.21.33.220
150.230.200.168
```

支持：
- CIDR 格式（如 `104.16.0.0/24`）
- 单个 IP（如 `8.219.217.25`）
- 以 `#` 开头的注释行

### 2. 运行扫描

```bash
node index.js ips.txt
```

或使用 npm script：

```bash
npm start ips.txt
```

### 3. 查看结果

扫描完成后，结果会保存在 `results.json` 文件中。

### 4. 示例输出

运行扫描时，控制台会显示：

```
🚀 开始扫描 Cloudflare IP...

📋 读取到 671 个 IP 条目（CIDR 或单个 IP）
🔍 共需扫描 671 个 IP 地址
⚙️  配置: 超时 5000ms, 并发数 64

扫描进度 |████████████████████| 100% | 671/671 | 可用: 45 | 平均延迟: 156ms

✅ 扫描完成！
📊 总计: 671 个 IP
✓  可用: 45 个 IP
✗  不可用: 626 个 IP

📍 地理位置分布:
   JP: 20 个 IP
   US: 15 个 IP
   SG: 10 个 IP

⚡ 响应时间统计:
   平均: 156 ms
   最快: 98 ms
   最慢: 245 ms

💾 结果已保存到: results.json
```

## 工作原理

### IP 可用性检测

程序通过以下步骤验证 Cloudflare IP：

1. **IP 转 16 进制**：将 IP 地址（如 `104.18.169.168`）转换为 16 进制格式（如 `0x6812A9A8`）
2. **构造测试 URL**：`https://0x6812A9A8.nip.lfree.org:443/cdn-cgi/trace`
3. **发送 HTTPS 请求**：尝试连接并获取响应
4. **解析结果**：从返回内容中提取 `loc=` 字段获取地理位置

### 示例响应

```text
fl=1009f40
h=0x6812a9a8.nip.lfree.org
ip=13.231.107.233
ts=1767761091.000
visit_scheme=https
colo=NRT
loc=JP
tls=TLSv1.3
warp=off
```

程序会提取 `loc=JP`，表示该 IP 位于日本。

## 工作原理

### 响应时间测试

每个 IP 会被测试 **3 次**，计算过程：

1. 第 1 次请求：记录响应时间（如 150ms）
2. 第 2 次请求：记录响应时间（如 145ms）
3. 第 3 次请求：记录响应时间（如 155ms）
4. 计算平均值：(150 + 145 + 155) / 3 = 150ms

如果任何一次请求失败，该 IP 将被标记为不可用。

## 配置参数

可以在 `index.js` 中修改配置：

```javascript
const CONFIG = {
  timeout: 5000,        // 超时时间（毫秒）
  maxConcurrency: 64,   // 最大并发数
  port: 443,            // 端口
  outputFile: 'results.json' // 输出文件名
};
```

## 输出格式

`results.json` 包含以下信息：

```json
{
  "scanTime": "2025-01-07T12:34:56.789Z",
  "totalScanned": 671,
  "totalAvailable": 45,
  "config": {
    "timeout": 5000,
    "maxConcurrency": 64,
    "port": 443,
    "outputFile": "results.json"
  },
  "locationStats": {
    "US": 20,
    "JP": 15,
    "SG": 10
  },
  "responseTimeStats": {
    "average": 156,
    "min": 98,
    "max": 245
  },
  "results": [
    {
      "ip": "104.18.169.168",
      "hexIp": "0x6812A9A8",
      "location": "JP",
      "available": true,
      "avgResponseTime": 150,
      "responseTimes": [148, 152, 150],
      "response": "fl=1009f40\nh=0x6812a9a8.nip.lfree.org\n..."
    }
  ]
}
```

字段说明：
- `avgResponseTime`: 3 次请求的平均响应时间（毫秒）
- `responseTimes`: 每次请求的具体响应时间数组
- `responseTimeStats`: 所有可用 IP 的响应时间统计

## 性能说明

- **并发数**：默认 64 个并发连接，可根据网络环境调整
- **超时时间**：5 秒，避免长时间等待无响应的 IP
- **测试次数**：每个 IP 测试 3 次，取平均响应时间
- **扫描速度**：由于每个 IP 测试 3 次，扫描时间是单次测试的 3 倍。671 个 IP 预计需要 5-10 分钟（取决于网络状况和成功率）

## 注意事项

1. **网络连接**：需要能够访问 Cloudflare 的网络
2. **防火墙**：确保出站 HTTPS (443) 端口未被阻止
3. **速率限制**：大量扫描可能触发 Cloudflare 的速率限制
4. **合法使用**：仅用于合法的网络测试和研究目的

## 故障排查

### 所有 IP 都显示不可用

- 检查网络连接
- 尝试手动访问测试 URL
- 增加超时时间
- 减少并发数

### 扫描速度过慢

- 增加并发数（但不要过高，可能导致网络拥塞）
- 减少超时时间

### 内存占用过高

- 减少并发数
- 分批处理 IP 段

## 许可证

MIT
