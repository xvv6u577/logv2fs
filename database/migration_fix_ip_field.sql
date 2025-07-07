-- 修复 subscription_nodes 表中的 IP 字段类型
-- 将 IP 字段从 inet 类型更改为 text 类型，以支持域名输入

BEGIN;

-- 第一步：创建一个临时列来存储转换后的数据
ALTER TABLE subscription_nodes ADD COLUMN ip_temp TEXT;

-- 第二步：将现有的 inet 数据转换为 text 格式
UPDATE subscription_nodes SET ip_temp = CAST(ip AS TEXT);

-- 第三步：删除原来的 inet 列
ALTER TABLE subscription_nodes DROP COLUMN ip;

-- 第四步：重命名临时列为原来的列名
ALTER TABLE subscription_nodes RENAME COLUMN ip_temp TO ip;

-- 第五步：为 IP 字段创建索引（可选，用于优化查询）
CREATE INDEX idx_subscription_nodes_ip ON subscription_nodes (ip);

COMMIT;

-- 验证更改
-- 查询表结构以确认更改
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'subscription_nodes' AND column_name = 'ip'; 