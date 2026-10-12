/**
 * Cloudflare Worker — SSL Certificate Monitor + Todo Manager
 *
 * SSL check 策略：通过 cloudflare:sockets 建立原始 TCP 连接，
 * 手动发送 TLS 1.2 ClientHello，解析握手帧，从 X.509 DER 证书
 * 中提取 issuer 和 notAfter，不依赖任何第三方 API。
 *
 * AnyTLS 同步：scheduled() 开始时从 logv2fs httpserver 的 /v1/cert-targets
 * 拉取 AnyTLS 节点（SNI + IP + 端口），upsert 到 ssl_certificates（source='anytls'）。
 * 检查时连接 connect_host（节点 IP），ClientHello 的 SNI 仍为 domain。
 *
 * 告警：状态变为 expiring / expired（可选 error）时通过 Telegram 推送一次，
 * 以 last_alert_status 去重；恢复为 valid 时发送恢复通知并清空 last_alert_status。
 */

import { connect } from 'cloudflare:sockets';
import { createClient, SupabaseClient } from '@supabase/supabase-js';

// ─────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────

interface Env {
  SUPABASE_URL: string;
  /** Secret：wrangler secret put SUPABASE_ANON_KEY */
  SUPABASE_ANON_KEY: string;
  ASSETS: { fetch: typeof fetch };
  /** Var：logv2fs httpserver 根地址，如 https://panel.example.com；为空则跳过 AnyTLS 同步 */
  LOGV2FS_API_URL?: string;
  /** Secret：与 logv2fs .env 中 CERT_TARGETS_TOKEN 相同 */
  CERT_TARGETS_TOKEN?: string;
  /** Secret：Telegram Bot token；与 TELEGRAM_CHAT_ID 同时配置才会发送告警 */
  TELEGRAM_BOT_TOKEN?: string;
  /** Secret 或 Var：Telegram chat id */
  TELEGRAM_CHAT_ID?: string;
  /** Var：设为 "true" 时，检查失败（error）也发送告警 */
  ALERT_ON_ERROR?: string;
  /** Secret：可选。设置后 /api/* 需要 Authorization: Bearer <ADMIN_TOKEN> */
  ADMIN_TOKEN?: string;
  /** Secret：可选。Cloudflare API token（Zone Read + SSL and Certificates Read），
   *  用于读取托管在 Cloudflare 上的域名的边缘证书（Worker 无法 connect() 到 Cloudflare IP） */
  CF_API_TOKEN?: string;
}

interface Todo {
  id?: number;
  title: string;
  completed: boolean;
  created_at?: string;
}

type CertStatus = 'valid' | 'expiring' | 'expired' | 'error' | 'pending';
type CertSource = 'manual' | 'anytls';

interface SSLCertificate {
  id?: number;
  domain: string;
  port?: number; // 证书检查端口，默认 443
  expiry_date?: string;
  issuer?: string;
  tags?: string[];
  last_checked?: string;
  status?: CertStatus;
  error_message?: string;
  connect_host?: string | null; // 实际 TCP 连接目标（如节点 IP），为空时连接 domain
  source?: CertSource;
  last_alert_status?: CertStatus | null;
  created_at?: string;
  updated_at?: string;
}

/** 执行一次证书检查需要的行字段 */
interface CertRow {
  id: number;
  domain: string;
  port: number;
  connect_host?: string | null;
  source?: CertSource;
  tags?: string[] | null;
  last_alert_status?: CertStatus | null;
}

const CERT_ROW_COLUMNS = 'id, domain, port, connect_host, source, tags, last_alert_status';

interface CertInfo {
  expiry: Date;
  issuer: string;
}

// ─────────────────────────────────────────────────────────────
// Binary Utilities
// ─────────────────────────────────────────────────────────────

const u16be = (n: number): Uint8Array =>
  new Uint8Array([(n >> 8) & 0xff, n & 0xff]);

const u24be = (n: number): Uint8Array =>
  new Uint8Array([(n >> 16) & 0xff, (n >> 8) & 0xff, n & 0xff]);

function concat(...arrays: Uint8Array[]): Uint8Array {
  const total = arrays.reduce((s, a) => s + a.length, 0);
  const out = new Uint8Array(total);
  let offset = 0;
  for (const a of arrays) { out.set(a, offset); offset += a.length; }
  return out;
}

// ─────────────────────────────────────────────────────────────
// Buffered Stream Reader
// 解决 TCP 流分包问题：跨多个 chunk 读取固定字节数
// ─────────────────────────────────────────────────────────────

class StreamBuffer {
  private chunks: Uint8Array[] = [];
  private available = 0;

  constructor(private readonly reader: ReadableStreamDefaultReader<Uint8Array>) {}

  async read(n: number): Promise<Uint8Array> {
    while (this.available < n) {
      const { value, done } = await this.reader.read();
      if (done) throw new Error('TCP stream closed before expected bytes arrived');
      this.chunks.push(value);
      this.available += value.length;
    }
    return this.consume(n);
  }

  private consume(n: number): Uint8Array {
    const result = new Uint8Array(n);
    let offset = 0, remaining = n;
    while (remaining > 0) {
      const chunk = this.chunks[0];
      if (chunk.length <= remaining) {
        result.set(chunk, offset);
        offset += chunk.length;
        remaining -= chunk.length;
        this.available -= chunk.length;
        this.chunks.shift();
      } else {
        result.set(chunk.subarray(0, remaining), offset);
        this.chunks[0] = chunk.subarray(remaining);
        this.available -= remaining;
        remaining = 0;
      }
    }
    return result;
  }
}

// ─────────────────────────────────────────────────────────────
// TLS 1.2 ClientHello Builder
//
// 故意不携带 `supported_versions` 扩展（TLS 1.3 的信号），
// 强制服务端协商 TLS 1.2，使 Certificate 握手消息保持明文。
// ─────────────────────────────────────────────────────────────

function buildClientHello(hostname: string): Uint8Array {
  const name = new TextEncoder().encode(hostname);

  // SNI 扩展 (type=0x0000)：告知服务端目标域名，获取正确证书
  const sniBody = concat(
    u16be(1 + 2 + name.length), // server_name_list_length
    new Uint8Array([0x00]),       // name_type: host_name(0)
    u16be(name.length),
    name,
  );
  const sniExt = concat(new Uint8Array([0x00, 0x00]), u16be(sniBody.length), sniBody);

  // supported_groups (type=0x000a)：secp256r1, secp384r1, secp521r1
  const groupsExt = new Uint8Array([
    0x00, 0x0a, 0x00, 0x08, 0x00, 0x06,
    0x00, 0x17, 0x00, 0x18, 0x00, 0x19,
  ]);

  // ec_point_formats (type=0x000b)：uncompressed
  const pointExt = new Uint8Array([0x00, 0x0b, 0x00, 0x02, 0x01, 0x00]);

  // signature_algorithms (type=0x000d)：ECDSA + RSA-PSS + RSA PKCS#1
  const sigAlgExt = new Uint8Array([
    0x00, 0x0d, 0x00, 0x14, 0x00, 0x12,
    0x04, 0x03, 0x05, 0x03, 0x06, 0x03,
    0x08, 0x04, 0x08, 0x05, 0x08, 0x06,
    0x04, 0x01, 0x05, 0x01, 0x06, 0x01,
  ]);

  const extensions = concat(sniExt, groupsExt, pointExt, sigAlgExt);

  // 仅包含 TLS 1.2 的 cipher suites
  const ciphers = new Uint8Array([
    0xc0, 0x2b, 0xc0, 0x2f, // ECDHE-{ECDSA,RSA}-AES128-GCM-SHA256
    0xc0, 0x2c, 0xc0, 0x30, // ECDHE-{ECDSA,RSA}-AES256-GCM-SHA384
    0x00, 0x9c, 0x00, 0x9d, // RSA-AES128/256-GCM-SHA256/384
    0x00, 0x35, 0x00, 0x2f, // AES256-SHA, AES128-SHA (fallback)
  ]);

  const body = concat(
    new Uint8Array([0x03, 0x03]),           // client_version: TLS 1.2
    crypto.getRandomValues(new Uint8Array(32)), // random (32 bytes)
    new Uint8Array([0x00]),                 // session_id_length: 0
    u16be(ciphers.length), ciphers,
    new Uint8Array([0x01, 0x00]),           // compression_methods: [null]
    u16be(extensions.length), extensions,
  );

  // Handshake message: type(1) + length(3) + body
  const handshake = concat(new Uint8Array([0x01]), u24be(body.length), body);

  // TLS Record: content_type(1) + version(2) + length(2) + payload
  // version 写 0x0301 (TLS 1.0) 是兼容性惯例
  return concat(new Uint8Array([0x16, 0x03, 0x01]), u16be(handshake.length), handshake);
}

// ─────────────────────────────────────────────────────────────
// TLS Record Parser
// 从 TLS 1.2 握手流中提取第一张叶证书（DER 编码）
// ─────────────────────────────────────────────────────────────

const CONTENT_ALERT     = 0x15;
const CONTENT_HANDSHAKE = 0x16;
const HS_CERTIFICATE    = 0x0b;

/**
 * 逐条读取 TLS Record，拼接握手帧缓冲区，
 * 找到 Certificate(type=11) 报文后返回第一张证书的 DER 字节。
 * 支持握手消息跨多个 TLS Record 的情形。
 */
async function extractCertDER(stream: StreamBuffer): Promise<Uint8Array> {
  // 跨 Record 的握手消息缓冲区
  let hsBuf = new Uint8Array(0);

  for (let iter = 0; iter < 64; iter++) {
    // TLS Record 头: content_type(1) + version(2) + length(2)
    const header  = await stream.read(5);
    const type    = header[0];
    const recvLen = (header[3] << 8) | header[4];
    const payload = await stream.read(recvLen);

    if (type === CONTENT_ALERT) {
      throw new Error(`TLS Alert received (level=${payload[0]}, desc=${payload[1]})`);
    }
    if (type !== CONTENT_HANDSHAKE) continue;

    hsBuf = concat(hsBuf, payload);

    // 从缓冲区中解析握手消息（可能有多条）
    let pos = 0;
    while (pos + 4 <= hsBuf.length) {
      const msgType = hsBuf[pos];
      const msgLen  = (hsBuf[pos + 1] << 16) | (hsBuf[pos + 2] << 8) | hsBuf[pos + 3];

      // 消息尚未完整接收，等待下一条 Record
      if (pos + 4 + msgLen > hsBuf.length) break;

      if (msgType === HS_CERTIFICATE) {
        const msg = hsBuf.subarray(pos + 4, pos + 4 + msgLen);
        // Certificate 消息格式:
        //   cert_list_length(3) + [ cert_length(3) + cert_DER ] * N
        const firstCertLen = (msg[3] << 16) | (msg[4] << 8) | msg[5];
        // slice() 复制一份，避免引用已废弃的缓冲区
        return msg.slice(6, 6 + firstCertLen);
      }

      pos += 4 + msgLen;
    }

    // 保留未处理的不完整数据
    hsBuf = pos > 0 ? hsBuf.subarray(pos) : hsBuf;
  }

  throw new Error('TLS Certificate message not found within handshake');
}

// ─────────────────────────────────────────────────────────────
// X.509 ASN.1 DER Parser
// 仅解析 issuer 和 validity.notAfter，足够轻量
// ─────────────────────────────────────────────────────────────

interface TLV {
  tag: number;
  value: Uint8Array;
  end: number; // 在父 data 中，该 TLV 之后的偏移量
}

function derTLV(data: Uint8Array, offset: number): TLV {
  const tag = data[offset++];

  // 解码长度字段（支持多字节长度）
  let len: number;
  if (data[offset] < 0x80) {
    len = data[offset++];
  } else {
    const numBytes = data[offset++] & 0x7f;
    len = 0;
    for (let i = 0; i < numBytes; i++) len = (len << 8) | data[offset++];
  }

  return { tag, value: data.subarray(offset, offset + len), end: offset + len };
}

/** 解析 ASN.1 OID 字节为点分十进制字符串，如 "2.5.4.3" */
function derParseOID(data: Uint8Array): string {
  const parts = [Math.floor(data[0] / 40), data[0] % 40];
  let val = 0;
  for (let i = 1; i < data.length; i++) {
    val = (val << 7) | (data[i] & 0x7f);
    if (!(data[i] & 0x80)) { parts.push(val); val = 0; }
  }
  return parts.join('.');
}

/** 解析 UTCTime (0x17) 或 GeneralizedTime (0x18) */
function derParseTime(tag: number, data: Uint8Array): Date {
  const s = new TextDecoder().decode(data);
  let y: number, rest: string;
  if (tag === 0x17) {
    // UTCTime: YYMMDDHHMMSSZ
    const yy = parseInt(s.slice(0, 2));
    y = yy + (yy >= 50 ? 1900 : 2000); // RFC 5280 规定
    rest = s.slice(2);
  } else {
    // GeneralizedTime: YYYYMMDDHHMMSSZ
    y = parseInt(s.slice(0, 4));
    rest = s.slice(4);
  }
  return new Date(Date.UTC(
    y,
    parseInt(rest.slice(0, 2)) - 1,
    parseInt(rest.slice(2, 4)),
    parseInt(rest.slice(4, 6)),
    parseInt(rest.slice(6, 8)),
    parseInt(rest.slice(8, 10)),
  ));
}

const KNOWN_OIDS: Record<string, string> = {
  '2.5.4.3':  'CN',
  '2.5.4.10': 'O',
  '2.5.4.11': 'OU',
  '2.5.4.6':  'C',
  '2.5.4.7':  'L',
  '2.5.4.8':  'ST',
};

/**
 * 将 RDNSequence 内容解析为可读字符串
 * 例：CN=example.com, O=Example Inc, C=US
 */
function derParseName(data: Uint8Array): string {
  const parts: string[] = [];
  let pos = 0;

  while (pos < data.length) {
    const rdn = derTLV(data, pos); // SET (0x31)
    pos = rdn.end;

    let atvPos = 0;
    while (atvPos < rdn.value.length) {
      const atv  = derTLV(rdn.value, atvPos); // SEQUENCE (0x30)
      atvPos     = atv.end;
      const oid  = derTLV(atv.value, 0);
      const val  = derTLV(atv.value, oid.end);
      const name = KNOWN_OIDS[derParseOID(oid.value)] ?? derParseOID(oid.value);
      try {
        parts.push(`${name}=${new TextDecoder().decode(val.value)}`);
      } catch {
        parts.push(`${name}=[binary]`);
      }
    }
  }

  return parts.join(', ');
}

/**
 * 解析 DER 编码的 X.509 证书，返回 issuer 和 notAfter。
 *
 * TBSCertificate 结构（RFC 5280）：
 *   version         [0] OPTIONAL  → tag 0xa0
 *   serialNumber    INTEGER       → tag 0x02
 *   signature       SEQUENCE      → tag 0x30 (skip)
 *   issuer          SEQUENCE      → tag 0x30 (parse)
 *   validity        SEQUENCE      → tag 0x30
 *     notBefore     Time
 *     notAfter      Time          ← 取这个
 */
function parseX509(der: Uint8Array): CertInfo {
  const cert = derTLV(der, 0);           // Certificate SEQUENCE
  const tbs  = derTLV(cert.value, 0);   // TBSCertificate SEQUENCE
  const d    = tbs.value;

  let pos = 0;
  if (d[pos] === 0xa0) pos = derTLV(d, pos).end; // skip [0] version
  pos = derTLV(d, pos).end;                       // skip serialNumber
  pos = derTLV(d, pos).end;                       // skip signatureAlgorithm

  const issuerTLV = derTLV(d, pos);
  const issuer    = derParseName(issuerTLV.value);
  pos             = issuerTLV.end;

  const validity    = derTLV(d, pos);
  const notBefore   = derTLV(validity.value, 0);
  const notAfterTLV = derTLV(validity.value, notBefore.end);
  const expiry      = derParseTime(notAfterTLV.tag, notAfterTLV.value);

  return { expiry, issuer };
}

// ─────────────────────────────────────────────────────────────
// SSL Certificate Check（核心业务逻辑）
// ─────────────────────────────────────────────────────────────

/** 证书即将过期的预警阈值（天） */
const EXPIRY_WARNING_DAYS = 30;

function calcStatus(expiry: Date): CertStatus {
  const days = Math.floor((expiry.getTime() - Date.now()) / 86_400_000);
  if (days < 0) return 'expired';
  if (days <= EXPIRY_WARNING_DAYS) return 'expiring';
  return 'valid';
}

/** 去掉 IPv6 字面量两侧的方括号，connect() 需要裸地址 */
function normalizeHost(host: string): string {
  const h = host.trim();
  return h.startsWith('[') && h.endsWith(']') ? h.slice(1, -1) : h;
}

interface CheckResult {
  status: CertStatus;
  days?: number;
  expiry?: Date;
  issuer?: string;
  error?: string;
}

/** 建立 TCP 连接、发送 ClientHello、解析证书；不触碰数据库 */
async function probeCertificate(connectHost: string, sni: string, port: number): Promise<CertInfo> {
  const socket = connect({ hostname: connectHost, port });
  const writer = socket.writable.getWriter();
  const reader = socket.readable.getReader();
  let timer: ReturnType<typeof setTimeout> | undefined;

  try {
    return await Promise.race([
      (async () => {
        // SNI（ClientHello 中携带的域名）始终使用纯域名，不能是 IP 或带端口
        await writer.write(buildClientHello(sni));
        const certDER = await extractCertDER(new StreamBuffer(reader));
        return parseX509(certDER);
      })(),
      // 15 秒超时，防止 Worker 挂起
      new Promise<never>((_, reject) => {
        timer = setTimeout(() => reject(new Error('Connection timed out after 15s')), 15_000);
      }),
    ]);
  } finally {
    if (timer) clearTimeout(timer);
    // 超时时 reader 仍有未完成的 read，releaseLock 会抛错并掩盖真实错误；先关连接，释放锁失败忽略
    await socket.close().catch(() => {});
    try { writer.releaseLock(); } catch { /* ignore */ }
    try { reader.releaseLock(); } catch { /* ignore */ }
  }
}

// ─────────────────────────────────────────────────────────────
// Cloudflare 托管域名：通过 Cloudflare API 读取边缘证书
// ─────────────────────────────────────────────────────────────

const CF_API_BASE = 'https://api.cloudflare.com/client/v4';
/** 写入 issuer 的标记，前端据此显示 Cloudflare 徽章 */
const CF_EDGE_MARKER = ' (Cloudflare 边缘)';
const CF_ZONE_CACHE_TTL_MS = 5 * 60_000;

interface CfZone { id: string; name: string }
interface CfCertificate { hosts?: string[]; issuer?: string; expires_on?: string; status?: string }
interface CfCertPack { status?: string; hosts?: string[]; certificates?: CfCertificate[] }

let cfZoneCache: { at: number; promise: Promise<CfZone[]> } | null = null;

async function cfApi<T>(env: Env, path: string): Promise<{ result: T; result_info?: { total_pages?: number } }> {
  const resp = await fetch(`${CF_API_BASE}${path}`, {
    headers: { Authorization: `Bearer ${env.CF_API_TOKEN}`, Accept: 'application/json' },
  });
  const body: any = await resp.json().catch(() => null);
  if (!resp.ok || !body?.success) {
    const msg = body?.errors?.map((e: any) => `${e.code}: ${e.message}`).join('; ') || `HTTP ${resp.status}`;
    throw new Error(`Cloudflare API ${path.split('?')[0]} failed: ${msg}`);
  }
  return body;
}

/** 列出 token 可见的全部 zone（分页）；同一 isolate 内缓存 5 分钟，并发检查共享同一请求 */
function listCfZones(env: Env): Promise<CfZone[]> {
  if (cfZoneCache && Date.now() - cfZoneCache.at < CF_ZONE_CACHE_TTL_MS) return cfZoneCache.promise;
  const promise = (async () => {
    const zones: CfZone[] = [];
    for (let page = 1; page <= 50; page++) {
      const body = await cfApi<CfZone[]>(env, `/zones?per_page=50&page=${page}`);
      zones.push(...body.result.map(z => ({ id: z.id, name: z.name.toLowerCase() })));
      if (page >= (body.result_info?.total_pages ?? 1)) break;
    }
    return zones;
  })();
  cfZoneCache = { at: Date.now(), promise };
  promise.catch(() => { if (cfZoneCache?.promise === promise) cfZoneCache = null; });
  return promise;
}

/** 找出 name 等于 domain 或为其后缀（.zone）的 zone，取最长匹配 */
function matchZone(zones: CfZone[], domain: string): CfZone | null {
  const d = domain.toLowerCase().replace(/\.$/, '');
  let best: CfZone | null = null;
  for (const z of zones) {
    if ((d === z.name || d.endsWith(`.${z.name}`)) && (!best || z.name.length > best.name.length)) best = z;
  }
  return best;
}

/** 0 = 不覆盖，1 = 通配符覆盖（单级），2 = 精确匹配 */
function hostCoverage(hosts: string[] | undefined, domain: string): 0 | 1 | 2 {
  let cov: 0 | 1 | 2 = 0;
  for (const raw of hosts ?? []) {
    const h = raw.toLowerCase();
    if (h === domain) return 2;
    if (h.startsWith('*.')) {
      const base = h.slice(2);
      const i = domain.indexOf('.');
      if (i > 0 && domain.slice(i + 1) === base) cov = 1;
    }
  }
  return cov;
}

/**
 * 从 Cloudflare API 读取 domain 的边缘证书到期时间。
 * 返回 null 表示不适用（未配置 token / 无匹配 zone / 无匹配证书），由调用方回退到直连探测。
 */
async function cfEdgeCertificate(env: Env, domain: string): Promise<CertInfo | null> {
  if (!env.CF_API_TOKEN) return null;
  const zone = matchZone(await listCfZones(env), domain);
  if (!zone) return null;

  const body = await cfApi<CfCertPack[]>(env, `/zones/${zone.id}/ssl/certificate_packs?status=all&per_page=50`);
  const d = domain.toLowerCase();
  const exact: CfCertificate[] = [];
  const wildcard: CfCertificate[] = [];
  for (const pack of body.result ?? []) {
    if (pack.status !== 'active') continue;
    for (const cert of pack.certificates ?? []) {
      if (cert.status !== 'active' || !cert.expires_on) continue;
      const cov = hostCoverage(cert.hosts ?? pack.hosts, d);
      if (cov === 2) exact.push(cert);
      else if (cov === 1) wildcard.push(cert);
    }
  }
  const chosen = exact.length ? exact : wildcard;
  let best: { expiry: Date; issuer: string } | null = null;
  for (const c of chosen) {
    const expiry = new Date(c.expires_on!);
    if (isNaN(expiry.getTime())) continue;
    if (!best || expiry > best.expiry) best = { expiry, issuer: c.issuer || 'Unknown' };
  }
  if (!best) {
    console.log(`[SSL] ${domain}: zone ${zone.name} has no active certificate covering it`);
    return null;
  }
  return { expiry: best.expiry, issuer: `${best.issuer}${CF_EDGE_MARKER}` };
}

/**
 * 按行检查证书：连接 connect_host（为空则 domain）的指定端口，
 * SNI 使用 domain；结果按 id 写回数据库，并处理告警状态迁移。
 */
async function checkSSLCertificate(row: CertRow, supabase: SupabaseClient, env: Env): Promise<CheckResult> {
  const port = row.port || DEFAULT_PORT;
  const host = row.connect_host ? normalizeHost(row.connect_host) : row.domain;
  const label = host === row.domain ? `${row.domain}:${port}` : `${row.domain}:${port} via ${host}`;
  let result: CheckResult;

  try {
    // 未指定 connect_host 时，先尝试 Cloudflare API（托管在 Cloudflare 的域名无法直连探测）
    let info: CertInfo | null = null;
    let cfNote = '';
    if (!row.connect_host && env.CF_API_TOKEN) {
      try {
        info = await cfEdgeCertificate(env, row.domain);
        if (info) console.log(`[SSL] ${label}: using Cloudflare API edge certificate`);
      } catch (err: any) {
        cfNote = err?.message ?? String(err);
        console.error(`[SSL] ${label}: Cloudflare API lookup failed: ${cfNote}`);
      }
    }
    if (!info) {
      try {
        info = await probeCertificate(host, row.domain, port);
      } catch (err: any) {
        const msg = err?.message ?? 'Unknown error';
        const hint = cfNote
          ? `（Cloudflare API 查询失败：${cfNote}）`
          : '（若该域名托管在 Cloudflare，Worker 无法直连其 IP；请确认 CF_API_TOKEN 能读取该 zone 的证书）';
        throw new Error(`${msg} ${hint}`);
      }
    }
    const { expiry, issuer } = info;
    const days = Math.floor((expiry.getTime() - Date.now()) / 86_400_000);
    result = { status: calcStatus(expiry), days, expiry, issuer };

    const { error } = await supabase
      .from('ssl_certificates')
      .update({
        expiry_date:   expiry.toISOString(),
        issuer,
        status:        result.status,
        last_checked:  new Date().toISOString(),
        error_message: null,
      })
      .eq('id', row.id);
    if (error) throw error;
    console.log(`[SSL] ${label} -> ${result.status} (${days}d remaining)`);
  } catch (err: any) {
    const message = err?.message ?? 'Unknown error';
    console.error(`[SSL] ${label} failed: ${message}`);
    result = { status: 'error', error: message };
    await supabase
      .from('ssl_certificates')
      .update({
        status:        'error',
        error_message: message,
        last_checked:  new Date().toISOString(),
      })
      .eq('id', row.id);
  }

  try {
    await handleAlertTransition(row, result, supabase, env);
  } catch (err: any) {
    console.error(`[Alert] ${label} alert handling failed: ${err?.message ?? err}`);
  }
  return result;
}

// ─────────────────────────────────────────────────────────────
// Telegram 告警（以 last_alert_status 去重）
// ─────────────────────────────────────────────────────────────

const ALERT_ICONS: Partial<Record<CertStatus, string>> = {
  expiring: '⚠️',
  expired:  '⛔',
  error:    '❗',
  valid:    '✅',
};

function telegramConfigured(env: Env): boolean {
  return Boolean(env.TELEGRAM_BOT_TOKEN && env.TELEGRAM_CHAT_ID);
}

async function sendTelegram(env: Env, text: string): Promise<boolean> {
  if (!telegramConfigured(env)) return false;
  try {
    const resp = await fetch(`https://api.telegram.org/bot${env.TELEGRAM_BOT_TOKEN}/sendMessage`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ chat_id: env.TELEGRAM_CHAT_ID, text, disable_web_page_preview: true }),
    });
    if (!resp.ok) {
      console.error(`[Alert] Telegram API responded ${resp.status}`);
      return false;
    }
    return true;
  } catch (err: any) {
    console.error(`[Alert] Telegram request failed: ${err?.message ?? err}`);
    return false;
  }
}

function formatAlert(row: CertRow, r: CheckResult, recovered: boolean): string {
  const target = `${row.domain}:${row.port || DEFAULT_PORT}`;
  const lines = [
    recovered
      ? `${ALERT_ICONS.valid} 证书已恢复正常：${target}`
      : `${ALERT_ICONS[r.status] ?? ''} 证书${r.status === 'expired' ? '已过期' : r.status === 'expiring' ? '即将过期' : '检查失败'}：${target}`,
  ];
  if (row.connect_host) lines.push(`连接地址：${row.connect_host}`);
  if (row.source === 'anytls') lines.push('来源：AnyTLS 同步');
  if (row.tags?.length) lines.push(`标签：${row.tags.join(', ')}`);
  if (r.expiry) lines.push(`到期时间：${r.expiry.toISOString()}（剩余 ${r.days} 天）`);
  if (r.issuer) lines.push(`颁发者：${r.issuer}`);
  if (r.error) lines.push(`错误：${r.error}`);
  return lines.join('\n');
}

/**
 * 状态迁移规则：
 *  - expiring / expired（以及 ALERT_ON_ERROR=true 时的 error）与 last_alert_status 不同 → 发送并记录；
 *  - valid 且 last_alert_status 非空 → 发送恢复通知（如已配置 Telegram）并清空；
 *  - error 但未开启 ALERT_ON_ERROR → 不改变 last_alert_status（避免瞬时失败打断去重）；
 *  - Telegram 未配置时不发送、也不记录（配置后会补发当前状态）。
 */
async function handleAlertTransition(row: CertRow, r: CheckResult, supabase: SupabaseClient, env: Env): Promise<void> {
  const last = row.last_alert_status ?? null;
  const alertOnError = (env.ALERT_ON_ERROR ?? '').toLowerCase() === 'true';
  const alertable = r.status === 'expiring' || r.status === 'expired' || (alertOnError && r.status === 'error');

  if (alertable) {
    if (r.status === last) return;
    if (!telegramConfigured(env)) return;
    if (await sendTelegram(env, formatAlert(row, r, false))) {
      await supabase.from('ssl_certificates').update({ last_alert_status: r.status }).eq('id', row.id);
    }
    return;
  }

  if (r.status === 'valid' && last) {
    if (telegramConfigured(env)) await sendTelegram(env, formatAlert(row, r, true));
    await supabase.from('ssl_certificates').update({ last_alert_status: null }).eq('id', row.id);
  }
}

// ─────────────────────────────────────────────────────────────
// AnyTLS 节点同步（logv2fs /v1/cert-targets → ssl_certificates）
// ─────────────────────────────────────────────────────────────

interface CertTarget {
  sni: string;
  ip: string;
  port: string;
  remark: string;
}

function isCertTargetList(v: unknown): v is CertTarget[] {
  return Array.isArray(v) && v.every(t =>
    t && typeof t === 'object' && typeof (t as any).sni === 'string' && typeof (t as any).port === 'string');
}

/**
 * 从 logv2fs 拉取 AnyTLS 节点并 upsert 到 ssl_certificates。
 *  - 以 (domain, port) 为唯一键；已存在的手动记录会被接管为 source='anytls'，标签合并保留；
 *  - 不再返回的 anytls 记录会被删除；若接口返回空列表则跳过删除（防误配置清空）；
 *  - 未配置 LOGV2FS_API_URL / CERT_TARGETS_TOKEN 时直接跳过。
 */
async function syncAnyTLSTargets(env: Env, supabase: SupabaseClient): Promise<void> {
  if (!env.LOGV2FS_API_URL || !env.CERT_TARGETS_TOKEN) {
    console.log('[Sync] LOGV2FS_API_URL or CERT_TARGETS_TOKEN not set; skip AnyTLS sync');
    return;
  }

  const url = `${env.LOGV2FS_API_URL.replace(/\/+$/, '')}/v1/cert-targets`;
  const resp = await fetch(url, { headers: { 'X-Cert-Token': env.CERT_TARGETS_TOKEN, Accept: 'application/json' } });
  if (!resp.ok) {
    console.error(`[Sync] ${url} responded ${resp.status}; skip`);
    return;
  }
  const body: unknown = await resp.json();
  if (!isCertTargetList(body)) {
    console.error('[Sync] unexpected response shape; skip');
    return;
  }

  // 规范化 + 去重
  const targets = new Map<string, { domain: string; port: number; connect_host: string | null; remark: string }>();
  for (const t of body) {
    const domain = t.sni.trim().toLowerCase();
    const port = Number(t.port) || DEFAULT_PORT;
    if (!DOMAIN_RE.test(domain) || port < 1 || port > 65535) {
      console.warn(`[Sync] skip invalid target ${t.sni}:${t.port}`);
      continue;
    }
    const key = `${domain}:${port}`;
    if (!targets.has(key)) {
      targets.set(key, { domain, port, connect_host: t.ip?.trim() ? normalizeHost(t.ip) : null, remark: t.remark ?? '' });
    }
  }

  const { data: existing, error: exErr } = await supabase
    .from('ssl_certificates')
    .select('id, domain, port, source, tags');
  if (exErr) throw exErr;
  const existingByKey = new Map<string, { id: number; source?: string; tags?: string[] | null }>();
  for (const e of existing ?? []) existingByKey.set(`${e.domain}:${e.port}`, e);

  const rows = [...targets.entries()].map(([key, t]) => {
    const prevTags = existingByKey.get(key)?.tags ?? [];
    const tags = [...new Set([...prevTags, 'anytls', ...(t.remark ? [t.remark] : [])])];
    return { domain: t.domain, port: t.port, connect_host: t.connect_host, source: 'anytls', tags };
  });

  if (rows.length) {
    const { error } = await supabase.from('ssl_certificates').upsert(rows, { onConflict: 'domain,port' });
    if (error) throw error;
  }

  const stale = (existing ?? []).filter(e => e.source === 'anytls' && !targets.has(`${e.domain}:${e.port}`));
  if (stale.length && targets.size === 0) {
    console.warn(`[Sync] API returned no targets; keep ${stale.length} existing anytls rows`);
  } else if (stale.length) {
    const { error } = await supabase.from('ssl_certificates').delete().in('id', stale.map(e => e.id));
    if (error) throw error;
  }
  console.log(`[Sync] AnyTLS targets upserted=${rows.length} removed=${targets.size === 0 ? 0 : stale.length}`);
}

// ─────────────────────────────────────────────────────────────
// HTTP Utilities
// ─────────────────────────────────────────────────────────────

const CORS_HEADERS: HeadersInit = {
  'Access-Control-Allow-Origin':  '*',
  'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
  'Access-Control-Allow-Headers': 'Content-Type, Authorization',
  'Content-Type':                 'application/json',
};

const json   = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: CORS_HEADERS });

const ok     = (data: unknown, status = 200) => json({ success: true,  data }, status);
const fail   = (message: string, status = 400) => json({ success: false, error: message }, status);
const server = (err: any) =>
  json({ success: false, error: err?.message ?? 'Server error', details: err?.details ?? null }, 500);

/** 常量时间字符串比较 */
function timingSafeEqual(a: string, b: string): boolean {
  if (a.length !== b.length) return false;
  let diff = 0;
  for (let i = 0; i < a.length; i++) diff |= a.charCodeAt(i) ^ b.charCodeAt(i);
  return diff === 0;
}

/** ADMIN_TOKEN 未配置时放行（向后兼容）；配置后要求 Authorization: Bearer <token> */
function isAuthorized(req: Request, env: Env): boolean {
  if (!env.ADMIN_TOKEN) return true;
  const auth = req.headers.get('Authorization') ?? '';
  const token = auth.startsWith('Bearer ') ? auth.slice(7).trim() : '';
  return token !== '' && timingSafeEqual(token, env.ADMIN_TOKEN);
}

const DOMAIN_RE = /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$/i;

/** 证书检查的默认端口（HTTPS 标准端口） */
const DEFAULT_PORT = 443;

/**
 * 解析用户输入的域名，支持 "example.com" 或 "example.com:8443" 两种形式。
 * 返回纯域名 + 端口号；输入不合法时返回 null（供上层给出友好错误提示）。
 */
function parseDomainInput(input: string): { domain: string; port: number } | null {
  const trimmed = input.trim().toLowerCase();
  const colonIndex = trimmed.lastIndexOf(':');

  // 没有冒号，说明未指定端口，使用默认端口 443
  if (colonIndex === -1) {
    return DOMAIN_RE.test(trimmed) ? { domain: trimmed, port: DEFAULT_PORT } : null;
  }

  const domain    = trimmed.slice(0, colonIndex);
  const portStr   = trimmed.slice(colonIndex + 1);
  const port      = Number(portStr);

  if (!DOMAIN_RE.test(domain)) return null;
  if (!/^\d+$/.test(portStr) || port < 1 || port > 65535) return null;

  return { domain, port };
}

// ─────────────────────────────────────────────────────────────
// Route Handlers
// ─────────────────────────────────────────────────────────────

async function handleCertificates(
  req: Request,
  path: string,
  supabase: SupabaseClient,
  ctx: ExecutionContext,
  env: Env,
): Promise<Response> {
  const method = req.method;

  // GET /api/certificates
  if (path === '/api/certificates' && method === 'GET') {
    const { data, error } = await supabase
      .from('ssl_certificates')
      .select('*')
      .order('created_at', { ascending: false });
    if (error) throw error;
    return ok(data);
  }

  // POST /api/certificates
  if (path === '/api/certificates' && method === 'POST') {
    const body = await req.json() as SSLCertificate;
    if (!body.domain) return fail('域名不能为空');

    // 支持 "example.com" 或 "example.com:8443" 两种输入形式
    const parsed = parseDomainInput(body.domain);
    if (!parsed) return fail('域名格式不正确，端口范围应为 1-65535');

    // 唯一键为 (domain, port)：同一域名可在多个端口分别监控；
    // 相同 (domain, port) 重复添加时覆盖 tags，并重置为 pending 触发一次新的证书检查。
    const { data, error } = await supabase
      .from('ssl_certificates')
      .upsert(
        [{ domain: parsed.domain, port: parsed.port, tags: body.tags ?? [], status: 'pending' }],
        { onConflict: 'domain,port' },
      )
      .select(CERT_ROW_COLUMNS);

    if (error) throw error;

    const created = data?.[0] as CertRow | undefined;
    if (created) ctx.waitUntil(checkSSLCertificate(created, supabase, env));
    return ok(data, 201);
  }

  // POST /api/certificates/check-all
  if (path === '/api/certificates/check-all' && method === 'POST') {
    // 先同步 AnyTLS 节点，失败不影响后续检查
    try { await syncAnyTLSTargets(env, supabase); } catch (e) { console.error('[Sync] failed:', e); }
    const { data, error } = await supabase.from('ssl_certificates').select(CERT_ROW_COLUMNS);
    if (error) throw error;
    const rows = (data ?? []) as CertRow[];
    if (rows.length) {
      ctx.waitUntil(Promise.allSettled(rows.map(c => checkSSLCertificate(c, supabase, env))));
    }
    return ok({ message: `正在检查 ${data?.length ?? 0} 个域名...` });
  }

  // /api/certificates/:id[/subpath]
  const m = path.match(/^\/api\/certificates\/(\d+)(\/[a-z-]+)?$/);
  if (!m) return fail('无效的请求路径', 400);

  const id      = parseInt(m[1]);
  const subPath = m[2] ?? '';

  // DELETE /api/certificates/:id
  if (!subPath && method === 'DELETE') {
    const { error } = await supabase.from('ssl_certificates').delete().eq('id', id);
    if (error) throw error;
    return ok(null);
  }

  // PUT /api/certificates/:id/tags
  if (subPath === '/tags' && method === 'PUT') {
    const { tags } = await req.json() as { tags: string[] };
    const { data, error } = await supabase
      .from('ssl_certificates')
      .update({ tags: tags ?? [] })
      .eq('id', id)
      .select();
    if (error) throw error;
    if (!data?.length) return fail('未找到该域名', 404);
    return ok(data);
  }

  // POST /api/certificates/:id/check
  if (subPath === '/check' && method === 'POST') {
    const { data: cert, error } = await supabase
      .from('ssl_certificates')
      .select(CERT_ROW_COLUMNS)
      .eq('id', id)
      .single();
    if (error || !cert) return fail('未找到该域名', 404);
    ctx.waitUntil(checkSSLCertificate(cert as CertRow, supabase, env));
    return ok({ message: '正在检查证书...' });
  }

  return fail('无效的请求', 400);
}

async function handleTodos(
  req: Request,
  path: string,
  supabase: SupabaseClient,
): Promise<Response> {
  const method = req.method;

  // GET /api/todos
  if (path === '/api/todos' && method === 'GET') {
    const { data, error } = await supabase
      .from('todos').select('*').order('created_at', { ascending: false });
    if (error) throw error;
    return ok(data);
  }

  // POST /api/todos
  if (path === '/api/todos' && method === 'POST') {
    const { title, completed } = await req.json() as Todo;
    if (!title) return fail('标题不能为空');
    const { data, error } = await supabase
      .from('todos')
      .insert([{ title, completed: completed ?? false }])
      .select();
    if (error) throw error;
    return ok(data, 201);
  }

  // /api/todos/:id
  const m = path.match(/^\/api\/todos\/(\d+)$/);
  if (!m) return fail('无效的请求路径', 400);
  const id = parseInt(m[1]);

  // GET /api/todos/:id
  if (method === 'GET') {
    const { data, error } = await supabase.from('todos').select('*').eq('id', id).single();
    if (error) {
      if (error.code === 'PGRST116') return fail('未找到该待办事项', 404);
      throw error;
    }
    return ok(data);
  }

  // PUT /api/todos/:id
  if (method === 'PUT') {
    const { title, completed } = await req.json() as Todo;
    const { data, error } = await supabase
      .from('todos').update({ title, completed }).eq('id', id).select();
    if (error) throw error;
    if (!data?.length) return fail('未找到该待办事项', 404);
    return ok(data);
  }

  // DELETE /api/todos/:id
  if (method === 'DELETE') {
    const { error } = await supabase.from('todos').delete().eq('id', id);
    if (error) throw error;
    return ok(null);
  }

  return fail('无效的请求', 400);
}

async function handleTags(supabase: SupabaseClient): Promise<Response> {
  const { data, error } = await supabase.from('ssl_certificates').select('tags');
  if (error) throw error;
  const all = new Set<string>();
  data?.forEach(({ tags }) => tags?.forEach((t: string) => all.add(t)));
  return ok([...all].sort());
}

// ─────────────────────────────────────────────────────────────
// Worker Entry Point
// ─────────────────────────────────────────────────────────────

export default {
  /** Cron Trigger：定时检查所有域名的证书 */
  async scheduled(_ctrl: ScheduledController, env: Env, _ctx: ExecutionContext): Promise<void> {
    console.log('[Cron] Starting scheduled SSL check...');
    if (!env.SUPABASE_URL || !env.SUPABASE_ANON_KEY) {
      console.error('[Cron] SUPABASE_URL / SUPABASE_ANON_KEY not configured');
      return;
    }
    const supabase = createClient(env.SUPABASE_URL, env.SUPABASE_ANON_KEY);

    // 1. 先同步 AnyTLS 节点；失败不影响后续检查
    try {
      await syncAnyTLSTargets(env, supabase);
    } catch (err: any) {
      console.error('[Cron] AnyTLS sync failed:', err?.message ?? err);
    }

    // 2. 检查全部记录（含刚同步进来的 pending 记录）
    const { data, error } = await supabase.from('ssl_certificates').select(CERT_ROW_COLUMNS);
    if (error) { console.error('[Cron] Failed to fetch domains:', error); return; }
    const rows = (data ?? []) as CertRow[];
    if (!rows.length) { console.log('[Cron] No domains to check'); return; }

    console.log(`[Cron] Checking ${rows.length} domains...`);
    await Promise.allSettled(rows.map(c => checkSSLCertificate(c, supabase, env)));
    console.log('[Cron] Done');
  },

  /** HTTP 请求处理 */
  async fetch(req: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const { pathname } = new URL(req.url);

    // 非 API 请求转发给静态资产
    if (!pathname.startsWith('/api/')) {
      return env.ASSETS.fetch(req);
    }

    // CORS 预检
    if (req.method === 'OPTIONS') {
      return new Response(null, { status: 204, headers: CORS_HEADERS });
    }

    if (!isAuthorized(req, env)) {
      return fail('Unauthorized', 401);
    }

    if (!env.SUPABASE_URL || !env.SUPABASE_ANON_KEY) {
      return fail('SUPABASE_URL / SUPABASE_ANON_KEY not configured', 500);
    }
    const supabase = createClient(env.SUPABASE_URL, env.SUPABASE_ANON_KEY);

    try {
      if (pathname.startsWith('/api/certificates')) {
        return await handleCertificates(req, pathname, supabase, ctx, env);
      }
      if (pathname === '/api/tags' && req.method === 'GET') {
        return await handleTags(supabase);
      }
      if (pathname.startsWith('/api/todos')) {
        return await handleTodos(req, pathname, supabase);
      }
      return fail('Not Found', 404);
    } catch (err: any) {
      console.error('[Worker] Unhandled error:', err);
      return server(err);
    }
  },
} satisfies ExportedHandler<Env>;