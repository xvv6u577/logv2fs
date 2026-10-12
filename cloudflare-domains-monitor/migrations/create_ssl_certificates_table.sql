-- 创建 SSL 证书监控表
CREATE TABLE IF NOT EXISTS ssl_certificates (
  id SERIAL PRIMARY KEY,
  domain VARCHAR(255) NOT NULL UNIQUE,
  expiry_date TIMESTAMP WITH TIME ZONE,
  issuer VARCHAR(255),
  tags TEXT[], -- PostgreSQL 数组类型，存储标签
  last_checked TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  status VARCHAR(50) DEFAULT 'pending', -- 'valid', 'expiring', 'expired', 'error', 'pending'
  error_message TEXT, -- 如果检查失败，存储错误信息
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_ssl_certs_domain ON ssl_certificates(domain);
CREATE INDEX IF NOT EXISTS idx_ssl_certs_expiry ON ssl_certificates(expiry_date);
CREATE INDEX IF NOT EXISTS idx_ssl_certs_status ON ssl_certificates(status);
CREATE INDEX IF NOT EXISTS idx_ssl_certs_tags ON ssl_certificates USING GIN(tags);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_ssl_certificates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器
DROP TRIGGER IF EXISTS trigger_ssl_certificates_updated_at ON ssl_certificates;
CREATE TRIGGER trigger_ssl_certificates_updated_at
  BEFORE UPDATE ON ssl_certificates
  FOR EACH ROW
  EXECUTE FUNCTION update_ssl_certificates_updated_at();

-- 创建预设标签的辅助视图（可选）
COMMENT ON TABLE ssl_certificates IS '存储域名 SSL 证书监控信息';
COMMENT ON COLUMN ssl_certificates.domain IS '域名（支持主域名和子域名）';
COMMENT ON COLUMN ssl_certificates.expiry_date IS '证书到期时间';
COMMENT ON COLUMN ssl_certificates.issuer IS '证书颁发者';
COMMENT ON COLUMN ssl_certificates.tags IS '标签数组，用于分类和筛选';
COMMENT ON COLUMN ssl_certificates.status IS '证书状态：valid(正常)、expiring(即将过期)、expired(已过期)、error(检查失败)、pending(待检查)';
COMMENT ON COLUMN ssl_certificates.last_checked IS '上次检查时间';
