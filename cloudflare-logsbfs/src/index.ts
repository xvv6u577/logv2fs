import { createClient } from '@supabase/supabase-js';

// 定义环境变量接口，添加Supabase相关配置
interface Env {
	SUPABASE_URL: string;
	SUPABASE_ANON_KEY: string;
	ASSETS: { fetch: typeof fetch };
}

// 定义待办事项类型
interface Todo {
	id?: number;
	title: string;
	completed: boolean;
	created_at?: string;
}

export default {
	async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
		// 获取请求URL和路径
		const url = new URL(request.url);
		const path = url.pathname;
		
		// 处理静态资产请求
		if (path === '/' || path === '/index.html' || !path.startsWith('/api/')) {
			return env.ASSETS.fetch(request);
		}
		
		// 初始化Supabase客户端
		const supabase = createClient(env.SUPABASE_URL, env.SUPABASE_ANON_KEY);
		
		// CORS 头部设置
		const corsHeaders = {
			'Access-Control-Allow-Origin': '*',
			'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
			'Access-Control-Allow-Headers': 'Content-Type, Authorization',
			'Content-Type': 'application/json'
		};
		
		// 处理 OPTIONS 请求（CORS预检请求）
		if (request.method === 'OPTIONS') {
			return new Response(null, {
				headers: corsHeaders,
				status: 204
			});
		}
		
		// 待办事项API路由
		if (path.startsWith('/api/todos')) {
			try {
				// 获取所有待办事项
				if (path === '/api/todos' && request.method === 'GET') {
					const { data, error } = await supabase
						.from('todos')
						.select('*')
						.order('created_at', { ascending: false });
					
					if (error) throw error;
					
					return new Response(JSON.stringify({ success: true, data }), {
						headers: corsHeaders
					});
				}
				
				// 创建新的待办事项
				if (path === '/api/todos' && request.method === 'POST') {
					const body = await request.json() as Todo;
					
					if (!body.title) {
						return new Response(JSON.stringify({ success: false, error: '标题不能为空' }), {
							status: 400,
							headers: corsHeaders
						});
					}
					
					const { data, error } = await supabase
						.from('todos')
						.insert([
							{ title: body.title, completed: body.completed || false }
						])
						.select();
					
					if (error) throw error;
					
					return new Response(JSON.stringify({ success: true, data }), {
						status: 201,
						headers: corsHeaders
					});
				}
				
				// 更新待办事项
				if (path.match(/^\/api\/todos\/\d+$/) && request.method === 'PUT') {
					const id = parseInt(path.split('/').pop() || '0');
					const body = await request.json() as Todo;
					
					const { data, error } = await supabase
						.from('todos')
						.update({ 
							title: body.title, 
							completed: body.completed 
						})
						.eq('id', id)
						.select();
					
					if (error) throw error;
					
					if (!data || data.length === 0) {
						return new Response(JSON.stringify({ success: false, error: '未找到该待办事项' }), {
							status: 404,
							headers: corsHeaders
						});
					}
					
					return new Response(JSON.stringify({ success: true, data }), {
						headers: corsHeaders
					});
				}
				
				// 删除待办事项
				if (path.match(/^\/api\/todos\/\d+$/) && request.method === 'DELETE') {
					const id = parseInt(path.split('/').pop() || '0');
					
					const { error } = await supabase
						.from('todos')
						.delete()
						.eq('id', id);
					
					if (error) throw error;
					
					return new Response(JSON.stringify({ success: true }), {
						headers: corsHeaders
					});
				}
				
				// 获取单个待办事项
				if (path.match(/^\/api\/todos\/\d+$/) && request.method === 'GET') {
					const id = parseInt(path.split('/').pop() || '0');
					
					const { data, error } = await supabase
						.from('todos')
						.select('*')
						.eq('id', id)
						.single();
					
					if (error) {
						if (error.code === 'PGRST116') {
							return new Response(JSON.stringify({ success: false, error: '未找到该待办事项' }), {
								status: 404,
								headers: corsHeaders
							});
						}
						throw error;
					}
					
					return new Response(JSON.stringify({ success: true, data }), {
						headers: corsHeaders
					});
				}
				
				// 如果没有匹配的路由
				return new Response(JSON.stringify({ success: false, error: '无效的请求' }), {
					status: 400,
					headers: corsHeaders
				});
				
			} catch (error: any) {
				console.error('Error:', error);
				
				return new Response(JSON.stringify({ 
					success: false, 
					error: error.message || '服务器错误',
					details: error.details || null
				}), {
					status: 500,
					headers: corsHeaders
				});
			}
		}
		
		// 默认响应，应该不会到达这里，因为非API请求会被静态资产处理
		return new Response('Hello World! Supabase Worker is running!', {
			headers: {
				'Content-Type': 'text/plain'
			}
		});
	},
} satisfies ExportedHandler<Env>;
