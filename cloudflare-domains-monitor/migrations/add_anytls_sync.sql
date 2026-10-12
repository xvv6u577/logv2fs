-- AnyTLS 节点证书同步所需的表结构变更（可重复执行）
--
-- 1. 唯一键由 domain 改为 (domain, port)：同一域名可在多个端口分别监控
-- 2. connect_host：实际 TCP 连接目标（如 AnyTLS 节点 IP），为空时连接 domain；SNI 始终为 domain
-- 3. source：'manual'（页面手动添加）| 'anytls'（从 logv2fs /v1/cert-targets 同步）
-- 4. last_alert_status：最近一次已发送告警的状态，用于 Telegram 告警去重

-- 1) 删除仅作用于 domain 列的 UNIQUE 约束（默认名 ssl_certificates_domain_key，这里按列查找以防改过名）
DO $$
DECLARE
  con record;
BEGIN
  FOR con IN
    SELECT c.conname
    FROM pg_constraint c
    JOIN pg_class t ON t.oid = c.conrelid
    WHERE t.relname = 'ssl_certificates'
      AND c.contype = 'u'
      AND array_length(c.conkey, 1) = 1
      AND c.conkey[1] = (
        SELECT attnum FROM pg_attribute
        WHERE attrelid = t.oid AND attname = 'domain'
      )
  LOOP
    EXECUTE format('ALTER TABLE ssl_certificates DROP CONSTRAINT %I', con.conname);
  END LOOP;
END $$;

-- 2) 新的复合唯一约束 (domain, port)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'uq_ssl_certificates_domain_port'
  ) THEN
    ALTER TABLE ssl_certificates
      ADD CONSTRAINT uq_ssl_certificates_domain_port UNIQUE (domain, port);
  END IF;
END $$;

-- 3) 新增列
ALTER TABLE ssl_certificates ADD COLUMN IF NOT EXISTS connect_host VARCHAR(255);
ALTER TABLE ssl_certificates ADD COLUMN IF NOT EXISTS source VARCHAR(20) NOT NULL DEFAULT 'manual';
ALTER TABLE ssl_certificates ADD COLUMN IF NOT EXISTS last_alert_status VARCHAR(50);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_ssl_certificates_source'
  ) THEN
    ALTER TABLE ssl_certificates
      ADD CONSTRAINT chk_ssl_certificates_source CHECK (source IN ('manual', 'anytls'));
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_ssl_certs_source ON ssl_certificates(source);

COMMENT ON COLUMN ssl_certificates.connect_host IS '实际 TCP 连接目标（如 AnyTLS 节点 IP），为空时连接 domain；SNI 始终使用 domain';
COMMENT ON COLUMN ssl_certificates.source IS '记录来源：manual（手动添加）| anytls（从 logv2fs 同步）';
COMMENT ON COLUMN ssl_certificates.last_alert_status IS '最近一次已发送 Telegram 告警时的状态，用于去重；恢复 valid 后清空';
