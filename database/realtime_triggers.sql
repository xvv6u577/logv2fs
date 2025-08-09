-- PostgreSQL Realtime 触发器函数
-- 用于在数据变更时发送 NOTIFY 消息到 WebSocket 监听器

-- 创建通用的变更通知函数
CREATE OR REPLACE FUNCTION notify_change()
RETURNS TRIGGER AS $$
DECLARE
    payload JSON;
    event_type TEXT;
BEGIN
    -- 确定事件类型
    IF TG_OP = 'INSERT' THEN
        event_type := 'INSERT';
        payload := json_build_object(
            'event', event_type,
            'table', TG_TABLE_NAME,
            'data', row_to_json(NEW)
        );
    ELSIF TG_OP = 'UPDATE' THEN
        event_type := 'UPDATE';
        payload := json_build_object(
            'event', event_type,
            'table', TG_TABLE_NAME,
            'data', row_to_json(NEW),
            'old_data', row_to_json(OLD)
        );
    ELSIF TG_OP = 'DELETE' THEN
        event_type := 'DELETE';
        payload := json_build_object(
            'event', event_type,
            'table', TG_TABLE_NAME,
            'data', row_to_json(OLD)
        );
    END IF;

    	-- 根据表名发送到不同的通知通道
	CASE TG_TABLE_NAME
		WHEN 'user_traffic_logs' THEN
			PERFORM pg_notify('traffic_update', payload::text);
		WHEN 'node_traffic_logs' THEN
			PERFORM pg_notify('node_traffic_update', payload::text);
		WHEN 'subscription_nodes' THEN
			PERFORM pg_notify('node_update', payload::text);
		WHEN 'payment_records' THEN
			PERFORM pg_notify('payment_update', payload::text);
		ELSE
			-- 默认发送到通用通道
			PERFORM pg_notify('data_change', payload::text);
	END CASE;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- 为用户表创建触发器
-- 注意：用户表可能在 MongoDB 中，这里只为 PostgreSQL 版本创建
CREATE OR REPLACE FUNCTION notify_user_change()
RETURNS TRIGGER AS $$
DECLARE
    payload JSON;
    event_type TEXT;
BEGIN
    -- 确定事件类型
    IF TG_OP = 'INSERT' THEN
        event_type := 'INSERT';
        payload := json_build_object(
            'event', event_type,
            'table', 'users',
            'data', row_to_json(NEW)
        );
    ELSIF TG_OP = 'UPDATE' THEN
        event_type := 'UPDATE';
        payload := json_build_object(
            'event', event_type,
            'table', 'users',
            'data', row_to_json(NEW),
            'old_data', row_to_json(OLD)
        );
    ELSIF TG_OP = 'DELETE' THEN
        event_type := 'DELETE';
        payload := json_build_object(
            'event', event_type,
            'table', 'users',
            'data', row_to_json(OLD)
        );
    END IF;

    -- 发送用户变更通知
    PERFORM pg_notify('user_update', payload::text);

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- 为流量日志表创建触发器
CREATE TRIGGER traffic_logs_notify_trigger
    AFTER INSERT OR UPDATE OR DELETE ON user_traffic_logs
    FOR EACH ROW EXECUTE FUNCTION notify_change();

-- 为节点流量日志表创建触发器
CREATE TRIGGER node_traffic_logs_notify_trigger
    AFTER INSERT OR UPDATE OR DELETE ON node_traffic_logs
    FOR EACH ROW EXECUTE FUNCTION notify_change();

-- 为订阅节点表创建触发器
CREATE TRIGGER subscription_nodes_notify_trigger
    AFTER INSERT OR UPDATE OR DELETE ON subscription_nodes
    FOR EACH ROW EXECUTE FUNCTION notify_change();

-- 为缴费记录表创建触发器
CREATE TRIGGER payment_records_notify_trigger
    AFTER INSERT OR UPDATE OR DELETE ON payment_records
    FOR EACH ROW EXECUTE FUNCTION notify_change();

-- 如果存在用户表，为其创建触发器
-- 注意：这个触发器只有在用户表存在于 PostgreSQL 中时才需要创建
-- DO $$
-- BEGIN
--     IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
--         CREATE TRIGGER users_notify_trigger
--             AFTER INSERT OR UPDATE OR DELETE ON users
--             FOR EACH ROW EXECUTE FUNCTION notify_user_change();
--     END IF;
-- END $$;

-- 创建测试函数（可选）
CREATE OR REPLACE FUNCTION test_notification(channel_name TEXT, message TEXT)
RETURNS VOID AS $$
BEGIN
    PERFORM pg_notify(channel_name, message);
END;
$$ LANGUAGE plpgsql;

-- 使用示例：
-- SELECT test_notification('user_update', '{"event": "TEST", "table": "users", "data": {"test": true}}'); 