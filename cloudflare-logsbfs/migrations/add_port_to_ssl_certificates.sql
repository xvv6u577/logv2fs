-- 为 ssl_certificates 增加端口字段
-- 支持形如 "cppla-sgp2.undervineyard.com:8443" 的添加方式：
-- domain 列仍只存纯域名，端口单独存放，默认 443（HTTPS 默认端口）。
ALTER TABLE ssl_certificates
  ADD COLUMN IF NOT EXISTS port INTEGER NOT NULL DEFAULT 443;

-- 端口合法范围校验
ALTER TABLE ssl_certificates
  ADD CONSTRAINT chk_ssl_certificates_port_range CHECK (port BETWEEN 1 AND 65535);

COMMENT ON COLUMN ssl_certificates.port IS '证书检查端口，默认 443；添加域名时可通过 "域名:端口" 的形式指定';
