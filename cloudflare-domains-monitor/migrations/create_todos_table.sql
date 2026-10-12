-- 创建todos表
CREATE TABLE IF NOT EXISTS todos (
  id SERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  completed BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 启用RLS（行级安全）
ALTER TABLE todos ENABLE ROW LEVEL SECURITY;

-- 创建一个用于更新updated_at字段的触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_todos_updated_at
BEFORE UPDATE ON todos
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- 创建RLS策略（此处为示例，允许所有用户访问所有数据）
-- 在实际应用中，你可能希望基于用户ID限制访问
CREATE POLICY "允许所有用户查看todos" ON todos
  FOR SELECT USING (true);

CREATE POLICY "允许所有用户插入自己的todos" ON todos
  FOR INSERT WITH CHECK (true);

CREATE POLICY "允许所有用户更新自己的todos" ON todos
  FOR UPDATE USING (true);

CREATE POLICY "允许所有用户删除自己的todos" ON todos
  FOR DELETE USING (true);

-- 插入一些示例数据
INSERT INTO todos (title, completed)
VALUES
  ('完成Cloudflare Worker项目', false),
  ('学习Supabase', false),
  ('部署应用', false); 