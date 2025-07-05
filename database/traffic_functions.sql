-- PostgreSQL存储函数：优化流量记录的upsert操作
-- 这些函数在数据库端执行JSON数组的条件更新，减少网络传输和提高性能

-- 用户流量记录upsert函数
-- 功能：插入或更新用户流量记录，同时处理daily_logs、monthly_logs、yearly_logs的条件更新
CREATE OR REPLACE FUNCTION upsert_user_traffic_log(
    p_email VARCHAR,
    p_timestamp TIMESTAMP,
    p_traffic BIGINT
) RETURNS VOID 
LANGUAGE plpgsql 
SECURITY DEFINER
AS $$
DECLARE
    v_date VARCHAR := TO_CHAR(p_timestamp, 'YYYYMMDD');
    v_month VARCHAR := TO_CHAR(p_timestamp, 'YYYYMM');
    v_year VARCHAR := TO_CHAR(p_timestamp, 'YYYY');
    
    v_daily_logs JSONB;
    v_monthly_logs JSONB;
    v_yearly_logs JSONB;
    v_daily_exists BOOLEAN;
    v_monthly_exists BOOLEAN;
    v_yearly_exists BOOLEAN;
BEGIN
    -- 使用WITH子句处理复杂的JSON更新逻辑
    WITH current_data AS (
        SELECT 
            COALESCE(daily_logs, '[]'::jsonb) as daily_logs,
            COALESCE(monthly_logs, '[]'::jsonb) as monthly_logs,
            COALESCE(yearly_logs, '[]'::jsonb) as yearly_logs
        FROM user_traffic_logs 
        WHERE email_as_id = p_email
        UNION ALL
        SELECT '[]'::jsonb, '[]'::jsonb, '[]'::jsonb
        WHERE NOT EXISTS (SELECT 1 FROM user_traffic_logs WHERE email_as_id = p_email)
        LIMIT 1
    ),
    updated_logs AS (
        SELECT 
            -- 更新或添加daily_logs
            CASE 
                WHEN EXISTS (
                    SELECT 1 FROM jsonb_array_elements(daily_logs) AS elem 
                    WHERE elem->>'date' = v_date
                ) THEN (
                    SELECT jsonb_agg(
                        CASE 
                            WHEN elem->>'date' = v_date 
                            THEN jsonb_set(elem, '{traffic}', ((elem->>'traffic')::bigint + p_traffic)::text::jsonb)
                            ELSE elem
                        END
                    )
                    FROM jsonb_array_elements(daily_logs) AS elem
                )
                ELSE daily_logs || jsonb_build_object('date', v_date, 'traffic', p_traffic)::jsonb
            END as new_daily_logs,
            
            -- 更新或添加monthly_logs
            CASE 
                WHEN EXISTS (
                    SELECT 1 FROM jsonb_array_elements(monthly_logs) AS elem 
                    WHERE elem->>'month' = v_month
                ) THEN (
                    SELECT jsonb_agg(
                        CASE 
                            WHEN elem->>'month' = v_month 
                            THEN jsonb_set(elem, '{traffic}', ((elem->>'traffic')::bigint + p_traffic)::text::jsonb)
                            ELSE elem
                        END
                    )
                    FROM jsonb_array_elements(monthly_logs) AS elem
                )
                ELSE monthly_logs || jsonb_build_object('month', v_month, 'traffic', p_traffic)::jsonb
            END as new_monthly_logs,
            
            -- 更新或添加yearly_logs
            CASE 
                WHEN EXISTS (
                    SELECT 1 FROM jsonb_array_elements(yearly_logs) AS elem 
                    WHERE elem->>'year' = v_year
                ) THEN (
                    SELECT jsonb_agg(
                        CASE 
                            WHEN elem->>'year' = v_year 
                            THEN jsonb_set(elem, '{traffic}', ((elem->>'traffic')::bigint + p_traffic)::text::jsonb)
                            ELSE elem
                        END
                    )
                    FROM jsonb_array_elements(yearly_logs) AS elem
                )
                ELSE yearly_logs || jsonb_build_object('year', v_year, 'traffic', p_traffic)::jsonb
            END as new_yearly_logs
        FROM current_data
    )
    -- 执行upsert操作
    INSERT INTO user_traffic_logs (
        email_as_id, name, password, uuid, user_id, status, role, used, credit, created_at, updated_at,
        hourly_logs, daily_logs, monthly_logs, yearly_logs
    )
    SELECT 
        p_email, p_email, 'default_traffic_user', gen_random_uuid()::text, gen_random_uuid()::text, 'plain', 'normal', p_traffic, 0, p_timestamp, p_timestamp,
        '[]'::jsonb, new_daily_logs, new_monthly_logs, new_yearly_logs
    FROM updated_logs
    ON CONFLICT (email_as_id) 
    DO UPDATE SET
        used = user_traffic_logs.used + p_traffic,
        updated_at = p_timestamp,
        daily_logs = EXCLUDED.daily_logs,
        monthly_logs = EXCLUDED.monthly_logs,
        yearly_logs = EXCLUDED.yearly_logs;
END;
$$;

-- 节点流量记录upsert函数
-- 功能：插入或更新节点流量记录，同时处理daily_logs、monthly_logs、yearly_logs的条件更新
CREATE OR REPLACE FUNCTION upsert_node_traffic_log(
    p_domain VARCHAR,
    p_timestamp TIMESTAMP,
    p_traffic BIGINT
) RETURNS VOID 
LANGUAGE plpgsql 
SECURITY DEFINER
AS $$
DECLARE
    v_date VARCHAR := TO_CHAR(p_timestamp, 'YYYYMMDD');
    v_month VARCHAR := TO_CHAR(p_timestamp, 'YYYYMM');
    v_year VARCHAR := TO_CHAR(p_timestamp, 'YYYY');
BEGIN
    -- 使用WITH子句处理复杂的JSON更新逻辑
    WITH current_data AS (
        SELECT 
            COALESCE(daily_logs, '[]'::jsonb) as daily_logs,
            COALESCE(monthly_logs, '[]'::jsonb) as monthly_logs,
            COALESCE(yearly_logs, '[]'::jsonb) as yearly_logs
        FROM node_traffic_logs 
        WHERE domain_as_id = p_domain
        UNION ALL
        SELECT '[]'::jsonb, '[]'::jsonb, '[]'::jsonb
        WHERE NOT EXISTS (SELECT 1 FROM node_traffic_logs WHERE domain_as_id = p_domain)
        LIMIT 1
    ),
    updated_logs AS (
        SELECT 
            -- 更新或添加daily_logs
            CASE 
                WHEN EXISTS (
                    SELECT 1 FROM jsonb_array_elements(daily_logs) AS elem 
                    WHERE elem->>'date' = v_date
                ) THEN (
                    SELECT jsonb_agg(
                        CASE 
                            WHEN elem->>'date' = v_date 
                            THEN jsonb_set(elem, '{traffic}', ((elem->>'traffic')::bigint + p_traffic)::text::jsonb)
                            ELSE elem
                        END
                    )
                    FROM jsonb_array_elements(daily_logs) AS elem
                )
                ELSE daily_logs || jsonb_build_object('date', v_date, 'traffic', p_traffic)::jsonb
            END as new_daily_logs,
            
            -- 更新或添加monthly_logs
            CASE 
                WHEN EXISTS (
                    SELECT 1 FROM jsonb_array_elements(monthly_logs) AS elem 
                    WHERE elem->>'month' = v_month
                ) THEN (
                    SELECT jsonb_agg(
                        CASE 
                            WHEN elem->>'month' = v_month 
                            THEN jsonb_set(elem, '{traffic}', ((elem->>'traffic')::bigint + p_traffic)::text::jsonb)
                            ELSE elem
                        END
                    )
                    FROM jsonb_array_elements(monthly_logs) AS elem
                )
                ELSE monthly_logs || jsonb_build_object('month', v_month, 'traffic', p_traffic)::jsonb
            END as new_monthly_logs,
            
            -- 更新或添加yearly_logs
            CASE 
                WHEN EXISTS (
                    SELECT 1 FROM jsonb_array_elements(yearly_logs) AS elem 
                    WHERE elem->>'year' = v_year
                ) THEN (
                    SELECT jsonb_agg(
                        CASE 
                            WHEN elem->>'year' = v_year 
                            THEN jsonb_set(elem, '{traffic}', ((elem->>'traffic')::bigint + p_traffic)::text::jsonb)
                            ELSE elem
                        END
                    )
                    FROM jsonb_array_elements(yearly_logs) AS elem
                )
                ELSE yearly_logs || jsonb_build_object('year', v_year, 'traffic', p_traffic)::jsonb
            END as new_yearly_logs
        FROM current_data
    )
    -- 执行upsert操作
    INSERT INTO node_traffic_logs (
        domain_as_id, remark, status, created_at, updated_at,
        hourly_logs, daily_logs, monthly_logs, yearly_logs
    )
    SELECT 
        p_domain, p_domain, 'active', p_timestamp, p_timestamp,
        '[]'::jsonb, new_daily_logs, new_monthly_logs, new_yearly_logs
    FROM updated_logs
    ON CONFLICT (domain_as_id) 
    DO UPDATE SET
        updated_at = p_timestamp,
        daily_logs = EXCLUDED.daily_logs,
        monthly_logs = EXCLUDED.monthly_logs,
        yearly_logs = EXCLUDED.yearly_logs;
END;
$$;

-- 创建索引以优化JSON查询性能（如果尚不存在）
-- 注意：在生产环境中，应该根据实际查询模式优化索引
CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_email ON user_traffic_logs(email_as_id);
CREATE INDEX IF NOT EXISTS idx_node_traffic_logs_domain ON node_traffic_logs(domain_as_id);

-- 可选：创建JSON字段的GIN索引以加速JSON查询
-- CREATE INDEX IF NOT EXISTS idx_user_traffic_daily_logs ON user_traffic_logs USING GIN (daily_logs);
-- CREATE INDEX IF NOT EXISTS idx_user_traffic_monthly_logs ON user_traffic_logs USING GIN (monthly_logs);
-- CREATE INDEX IF NOT EXISTS idx_user_traffic_yearly_logs ON user_traffic_logs USING GIN (yearly_logs);
-- CREATE INDEX IF NOT EXISTS idx_node_traffic_daily_logs ON node_traffic_logs USING GIN (daily_logs);
-- CREATE INDEX IF NOT EXISTS idx_node_traffic_monthly_logs ON node_traffic_logs USING GIN (monthly_logs);
-- CREATE INDEX IF NOT EXISTS idx_node_traffic_yearly_logs ON node_traffic_logs USING GIN (yearly_logs); 