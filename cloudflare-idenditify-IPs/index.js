import fs from 'fs';
import https from 'https';
import { setTimeout } from 'timers/promises';
import cliProgress from 'cli-progress';
import pLimit from 'p-limit';

// ========== 配置参数 ==========
const CONFIG = {
  timeout: 5000,        // 超时时间：5秒
  maxConcurrency: 100,   // 最大并发数：100
  port: 443,            // Cloudflare端口
  outputFile: 'results.json' // 输出文件
};

// ========== IP 转换为 16 进制 ==========
/**
 * 将 IP 地址转换为 16 进制格式
 * @param {string} ip - IP 地址，如 "104.18.169.168"
 * @returns {string} 16 进制格式，如 "0x6812A9A8"
 */
function ipToHex(ip) {
  const parts = ip.split('.');
  const hex = parts.map(part => {
    const num = parseInt(part, 10);
    return num.toString(16).padStart(2, '0');
  }).join('');
  return '0x' + hex.toUpperCase();
}

// ========== CIDR 转 IP 列表 ==========
/**
 * 将 CIDR 格式转换为 IP 列表
 * @param {string} cidr - CIDR 格式，如 "1.1.1.0/24"
 * @returns {string[]} IP 地址列表
 */
function cidrToIpList(cidr) {
  const [baseIp, prefixLength] = cidr.split('/');
  const prefix = parseInt(prefixLength, 10);

  // 将 IP 转换为 32 位整数
  const ipToInt = (ip) => {
    const parts = ip.split('.');
    return (parseInt(parts[0]) << 24) +
      (parseInt(parts[1]) << 16) +
      (parseInt(parts[2]) << 8) +
      parseInt(parts[3]);
  };

  // 将 32 位整数转换为 IP
  const intToIp = (num) => {
    return [
      (num >>> 24) & 0xFF,
      (num >>> 16) & 0xFF,
      (num >>> 8) & 0xFF,
      num & 0xFF
    ].join('.');
  };

  const baseIpInt = ipToInt(baseIp);
  const hostBits = 32 - prefix;
  const numIps = Math.pow(2, hostBits);

  // 计算网络地址（清除主机位）
  const networkInt = baseIpInt & (~0 << hostBits);

  const ips = [];
  for (let i = 0; i < numIps; i++) {
    ips.push(intToIp(networkInt + i));
  }

  return ips;
}

// ========== 测试 IP 可用性（单次） ==========
/**
 * 测试 Cloudflare IP 是否可用，并获取地理位置（单次测试）
 * @param {string} ip - 待测试的 IP 地址
 * @returns {Promise<Object|null>} 返回结果对象或 null
 */
function testCloudflareIPOnce(ip) {
  return new Promise((resolve) => {
    const hexIP = ipToHex(ip);
    const startTime = Date.now();

    const options = {
      hostname: `${hexIP}.nip.lfree.org`,
      port: CONFIG.port,
      path: `/cdn-cgi/trace?_t=${Date.now()}`,
      method: 'GET',
      rejectUnauthorized: false, // 忽略证书验证
      timeout: CONFIG.timeout
    };

    const req = https.request(options, (res) => {
      let data = '';

      res.on('data', (chunk) => {
        data += chunk;
      });

      res.on('end', () => {
        const endTime = Date.now();
        const responseTime = endTime - startTime;

        // 从返回内容中提取 loc 字段
        const locMatch = data.match(/loc=([A-Z]{2})/);

        if (locMatch && locMatch[1]) {
          resolve({
            success: true,
            ip: ip,
            hexIp: hexIP,
            location: locMatch[1],
            responseTime: responseTime,
            response: data
          });
        } else {
          // 有响应但没有 loc 字段
          resolve({ success: false });
        }
      });
    });

    req.on('error', () => {
      // 连接失败，跳过
      resolve({ success: false });
    });

    req.on('timeout', () => {
      req.destroy();
      resolve({ success: false });
    });

    req.end();
  });
}

// ========== 测试 IP 可用性（3 次取平均） ==========
/**
 * 测试 Cloudflare IP 3 次并计算平均响应时间
 * @param {string} ip - 待测试的 IP 地址
 * @returns {Promise<Object|null>} 返回结果对象或 null
 */
async function testCloudflareIP(ip) {
  const results = [];

  // 请求 3 次
  for (let i = 0; i < 3; i++) {
    const result = await testCloudflareIPOnce(ip);
    if (result.success) {
      results.push(result);
    } else {
      // 如果有一次失败，直接返回 null
      return null;
    }
  }

  // 如果 3 次都成功，计算平均响应时间
  if (results.length === 3) {
    const avgResponseTime = Math.round(
      results.reduce((sum, r) => sum + r.responseTime, 0) / 3
    );

    return {
      ip: results[0].ip,
      hexIp: results[0].hexIp,
      location: results[0].location,
      available: true,
      avgResponseTime: avgResponseTime,
      responseTimes: results.map(r => r.responseTime),
      response: results[0].response
    };
  }

  return null;
}

// ========== 主函数 ==========
async function main() {
  const args = process.argv.slice(2);

  if (args.length === 0) {
    console.log('用法: node index.js <IP段文件.txt>');
    console.log('示例: node index.js ips.txt');
    process.exit(1);
  }

  const inputFile = args[0];

  // 检查文件是否存在
  if (!fs.existsSync(inputFile)) {
    console.error(`❌ 文件不存在: ${inputFile}`);
    process.exit(1);
  }

  console.log('🚀 开始扫描 Cloudflare IP...\n');

  // 读取 IP 列表
  const rawLines = fs.readFileSync(inputFile, 'utf-8')
    .split('\n')
    .map(line => line.trim())
    .filter(line => line && !line.startsWith('#')); // 过滤空行和注释

  console.log(`📋 读取到 ${rawLines.length} 个 IP 条目（CIDR、IP:端口#国家 或单个 IP）`);

  // 将所有格式转换为 IP 列表，使用 Set 去重
  const ipSet = new Set();
  
  for (const line of rawLines) {
    try {
      let processedLine = line;
      
      // 处理 IP:端口#国家 格式（如：103.102.228.113:443#NL）
      if (processedLine.includes('#')) {
        // 移除 # 及后面的国家代码
        processedLine = processedLine.split('#')[0];
      }
      
      if (processedLine.includes(':')) {
        // 移除端口号，只保留 IP
        processedLine = processedLine.split(':')[0];
      }
      
      // 判断是 CIDR 格式还是单个 IP
      if (processedLine.includes('/')) {
        // CIDR 格式，展开为 IP 列表
        const ips = cidrToIpList(processedLine);
        ips.forEach(ip => ipSet.add(ip));
      } else {
        // 单个 IP，添加到集合（自动去重）
        ipSet.add(processedLine);
      }
    } catch (error) {
      console.error(`⚠️  解析失败: ${line}`);
    }
  }
  
  // 将 Set 转换为数组
  const allIps = Array.from(ipSet);

  console.log(`🔍 共需扫描 ${allIps.length} 个 IP 地址`);
  console.log(`⚙️  配置: 超时 ${CONFIG.timeout}ms, 并发数 ${CONFIG.maxConcurrency}\n`);

  // 创建进度条
  const progressBar = new cliProgress.SingleBar({
    format: '扫描进度 |{bar}| {percentage}% | {value}/{total} | 可用: {available} | 平均延迟: {avgTime}ms',
    barCompleteChar: '\u2588',
    barIncompleteChar: '\u2591',
    hideCursor: true
  });

  progressBar.start(allIps.length, 0, { available: 0, avgTime: '-' });

  // 并发控制
  const limit = pLimit(CONFIG.maxConcurrency);
  const results = [];
  let completed = 0;
  let availableCount = 0;

  // 创建任务队列
  const tasks = allIps.map(ip =>
    limit(async () => {
      const result = await testCloudflareIP(ip);

      if (result) {
        results.push(result);
        availableCount++;
      }

      completed++;

      // 计算当前平均响应时间
      const currentAvgTime = results.length > 0
        ? Math.round(results.reduce((sum, r) => sum + r.avgResponseTime, 0) / results.length)
        : '-';

      progressBar.update(completed, {
        available: availableCount,
        avgTime: currentAvgTime
      });

      return result;
    })
  );

  // 等待所有任务完成
  await Promise.all(tasks);

  progressBar.stop();

  console.log(`\n✅ 扫描完成！`);
  console.log(`📊 总计: ${allIps.length} 个 IP`);
  console.log(`✓  可用: ${availableCount} 个 IP`);
  console.log(`✗  不可用: ${allIps.length - availableCount} 个 IP\n`);

  // 按国家/地区统计
  const locationStats = {};
  results.forEach(result => {
    const loc = result.location;
    locationStats[loc] = (locationStats[loc] || 0) + 1;
  });

  console.log('📍 地理位置分布:');
  Object.entries(locationStats)
    .sort((a, b) => b[1] - a[1])
    .forEach(([loc, count]) => {
      console.log(`   ${loc}: ${count} 个 IP`);
    });

  // 响应时间统计
  if (results.length > 0) {
    const allResponseTimes = results.map(r => r.avgResponseTime);
    const avgTime = Math.round(allResponseTimes.reduce((sum, t) => sum + t, 0) / allResponseTimes.length);
    const minTime = Math.min(...allResponseTimes);
    const maxTime = Math.max(...allResponseTimes);

    console.log('\n⚡ 响应时间统计:');
    console.log(`   平均: ${avgTime} ms`);
    console.log(`   最快: ${minTime} ms`);
    console.log(`   最慢: ${maxTime} ms`);
  }

  // 保存结果到 JSON 文件
  const responseTimeStats = results.length > 0 ? {
    average: Math.round(results.reduce((sum, r) => sum + r.avgResponseTime, 0) / results.length),
    min: Math.min(...results.map(r => r.avgResponseTime)),
    max: Math.max(...results.map(r => r.avgResponseTime))
  } : null;

  const output = {
    scanTime: new Date().toISOString(),
    totalScanned: allIps.length,
    totalAvailable: availableCount,
    config: CONFIG,
    locationStats: locationStats,
    responseTimeStats: responseTimeStats,
    results: results
  };

  fs.writeFileSync(CONFIG.outputFile, JSON.stringify(output, null, 2));
  console.log(`\n💾 结果已保存到: ${CONFIG.outputFile}`);
}

// 执行主函数
main().catch(error => {
  console.error('❌ 发生错误:', error);
  process.exit(1);
});
