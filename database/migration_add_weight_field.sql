-- 为 subscription_nodes 表添加 weight 字段
-- 用于节点排序，数值越小越靠前
-- 作者: logv2fs
-- 日期: 2025-12-28

BEGIN;

-- 步骤 1: 添加 weight 字段，默认值为 0
ALTER TABLE subscription_nodes 
ADD COLUMN IF NOT EXISTS weight INTEGER DEFAULT 0;

-- 步骤 2: 创建索引以优化排序查询
CREATE INDEX IF NOT EXISTS idx_subscription_nodes_weight 
ON subscription_nodes (weight);

-- 步骤 3: 为现有记录设置初始权重值
-- 策略：按创建时间顺序设置权重，早创建的节点权重小（排在前面）
WITH ranked_nodes AS (
    SELECT id, ROW_NUMBER() OVER (ORDER BY created_at) - 1 AS row_num
    FROM subscription_nodes
)
UPDATE subscription_nodes 
SET weight = ranked_nodes.row_num * 10  -- 乘以10留出调整空间
FROM ranked_nodes
WHERE subscription_nodes.id = ranked_nodes.id
  AND subscription_nodes.weight = 0;  -- 只更新尚未设置权重的记录

COMMIT;

-- 验证更改
SELECT column_name, data_type, column_default, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'subscription_nodes' AND column_name = 'weight';

-- 显示当前节点的权重分布
SELECT remark, weight, created_at 
FROM subscription_nodes 
ORDER BY weight ASC;
