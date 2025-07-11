-- 为用户流量日志表添加 remark 字段
-- 用于存储用户备注信息

BEGIN;

-- 第一步：检查 remark 字段是否已存在
DO $$
BEGIN
    -- 如果字段不存在，则添加它
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'user_traffic_logs' 
        AND column_name = 'remark'
    ) THEN
        -- 添加 remark 字段
        ALTER TABLE user_traffic_logs ADD COLUMN remark TEXT;
        
        -- 为所有现有记录设置默认值（空字符串）
        UPDATE user_traffic_logs SET remark = '' WHERE remark IS NULL;
        
        RAISE NOTICE 'remark 字段已成功添加到 user_traffic_logs 表';
    ELSE
        RAISE NOTICE 'remark 字段已存在于 user_traffic_logs 表中';
    END IF;
END
$$;

COMMIT;

-- 验证更改
-- 查询表结构以确认更改
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'user_traffic_logs' AND column_name = 'remark';

-- 显示迁移完成信息
SELECT 'Migration completed: remark field added to user_traffic_logs table' AS status; 