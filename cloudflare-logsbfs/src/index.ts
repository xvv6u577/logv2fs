/**
 * Cloudflare Worker — SSL Certificate Monitor + Todo Manager
 *
 * SSL check 策略：通过 cloudflare:sockets 建立原始 TCP 连接，
 * 手动发送 TLS 1.2 ClientHello，解析握手帧，从 X.509 DER 证书
 * 中提取 issuer 和 notAfter，不依赖任何第三方 API。
 */

import { connect } from 'cloudflare:sockets';
import { createClient, SupabaseClient } from '@supabase/supabase-js';

// ─────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────

interface Env {
  SUPABASE_URL: string;
  SUPABASE_ANON_KEY: string;
  ASSETS: { fetch: typeof fetch };
}

interface Todo {
  id?: number;
  title: string;
  completed: boolean;
  created_at?: string;
}

type CertStatus = 'valid' | 'expiring' | 'expired' | 'error' | 'pending';

interface SSLCertificate {
  id?: number;
  domain: string;
  expiry_date?: string;
  issuer?: string;
  tags?: string[];
  last_checked?: string;
  status?: CertStatus;
  error_message?: string;
  created_at?: string;
  updated_at?: string;
}

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

/**
 * 直接连接域名的 443 端口，完成 TLS 握手，
 * 从服务端证书中解析有效期和颁发者，写入数据库。
 */
async function checkSSLCertificate(domain: string, supabase: SupabaseClient): Promise<void> {
  const socket = connect({ hostname: domain, port: 443 });
  const writer = socket.writable.getWriter();
  const reader = socket.readable.getReader();

  try {
    await Promise.race([
      (async () => {
        // 1. 发送 TLS 1.2 ClientHello
        await writer.write(buildClientHello(domain));

        // 2. 从 TLS 握手流中提取证书 DER
        const certDER = await extractCertDER(new StreamBuffer(reader));

        // 3. 解析 X.509 证书
        const { expiry, issuer } = parseX509(certDER);
        const status = calcStatus(expiry);
        const days   = Math.floor((expiry.getTime() - Date.now()) / 86_400_000);

        // 4. 更新数据库
        const { error } = await supabase
          .from('ssl_certificates')
          .update({
            expiry_date:   expiry.toISOString(),
            issuer,
            status,
            last_checked:  new Date().toISOString(),
            error_message: null,
          })
          .eq('domain', domain);

        if (error) throw error;
        console.log(`[SSL] ${domain}: ${status} (${days}d remaining)`);
      })(),

      // 15 秒超时，防止 Worker 挂起
      new Promise<never>((_, reject) =>
        setTimeout(() => reject(new Error('Connection timed out after 15s')), 15_000)
      ),
    ]);
  } catch (err: any) {
    console.error(`[SSL] ${domain} failed: ${err.message}`);
    await supabase
      .from('ssl_certificates')
      .update({
        status:        'error',
        error_message: err.message ?? 'Unknown error',
        last_checked:  new Date().toISOString(),
      })
      .eq('domain', domain);
  } finally {
    writer.releaseLock();
    reader.releaseLock();
    await socket.close().catch(() => {});
  }
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

const DOMAIN_RE = /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$/i;

// ─────────────────────────────────────────────────────────────
// Route Handlers
// ─────────────────────────────────────────────────────────────

async function handleCertificates(
  req: Request,
  path: string,
  supabase: SupabaseClient,
  ctx: ExecutionContext,
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
    if (!body.domain)              return fail('域名不能为空');
    if (!DOMAIN_RE.test(body.domain)) return fail('域名格式不正确');

    const { data, error } = await supabase
      .from('ssl_certificates')
      .insert([{ domain: body.domain.toLowerCase(), tags: body.tags ?? [], status: 'pending' }])
      .select();

    if (error) {
      if (error.code === '23505') return fail('该域名已存在', 409);
      throw error;
    }

    if (data?.[0]) ctx.waitUntil(checkSSLCertificate(data[0].domain, supabase));
    return ok(data, 201);
  }

  // POST /api/certificates/check-all
  if (path === '/api/certificates/check-all' && method === 'POST') {
    const { data, error } = await supabase.from('ssl_certificates').select('domain');
    if (error) throw error;
    if (data?.length) {
      ctx.waitUntil(
        Promise.allSettled(data.map(c => checkSSLCertificate(c.domain, supabase)))
      );
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
      .select('domain')
      .eq('id', id)
      .single();
    if (error || !cert) return fail('未找到该域名', 404);
    ctx.waitUntil(checkSSLCertificate(cert.domain, supabase));
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
    const supabase = createClient(env.SUPABASE_URL, env.SUPABASE_ANON_KEY);

    const { data, error } = await supabase.from('ssl_certificates').select('domain');
    if (error) { console.error('[Cron] Failed to fetch domains:', error); return; }
    if (!data?.length) { console.log('[Cron] No domains to check'); return; }

    console.log(`[Cron] Checking ${data.length} domains...`);
    await Promise.allSettled(data.map(c => checkSSLCertificate(c.domain, supabase)));
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

    const supabase = createClient(env.SUPABASE_URL, env.SUPABASE_ANON_KEY);

    try {
      if (pathname.startsWith('/api/certificates')) {
        return await handleCertificates(req, pathname, supabase, ctx);
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