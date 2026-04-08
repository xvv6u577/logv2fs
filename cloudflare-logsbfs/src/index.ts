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

// 定义 SSL 证书类型
interface SSLCertificate {
	id?: number;
	domain: string;
	expiry_date?: string;
	issuer?: string;
	tags?: string[];
	last_checked?: string;
	status?: 'valid' | 'expiring' | 'expired' | 'error' | 'pending';
	error_message?: string;
	created_at?: string;
	updated_at?: string;
}

// SSL Labs API 响应类型
interface SSLLabsResponse {
	host: string;
	status: string;
	endpoints?: Array<{
		statusMessage: string;
		grade: string;
		hasWarnings: boolean;
		details?: {
			cert?: {
				subject: string;
				notAfter: number; // Unix timestamp in milliseconds
				notBefore: number;
				issuerSubject: string;
			};
		};
	}>;
}

/**
 * 使用备用方案检查证书（简化版，使用 crt.sh）
 * 注意：这个方案更快但信息较少
 */
async function checkSSLCertificateSimple(domain: string, supabase: any): Promise<void> {
	try {
		// 使用 crt.sh API 获取证书信息
		const crtshUrl = `https://crt.sh/?q=${encodeURIComponent(domain)}&output=json`;
		const response = await fetch(crtshUrl);
		
		if (!response.ok) {
			throw new Error(`crt.sh API 返回错误: ${response.status}`);
		}
		
		const certificates = await response.json() as Array<{
			not_after: string;
			issuer_name: string;
			common_name: string;
		}>;
		
		if (!certificates || certificates.length === 0) {
			throw new Error('未找到证书信息');
		}
		
		// 找到最新的证书
		const latestCert = certificates.sort((a, b) => 
			new Date(b.not_after).getTime() - new Date(a.not_after).getTime()
		)[0];

		const expiryDate = new Date(latestCert.not_after);
		const now = new Date();
		const daysRemaining = Math.floor((expiryDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
		
		// 确定状态
		let status: 'valid' | 'expiring' | 'expired';
		if (daysRemaining < 0) {
			status = 'expired';
		} else if (daysRemaining <= 7) {
			status = 'expiring';
		} else {
			status = 'valid';
		}
		
		// 更新数据库
		const { error: updateError } = await supabase
			.from('ssl_certificates')
			.update({
				expiry_date: expiryDate.toISOString(),
				issuer: latestCert.issuer_name || 'Unknown',
				status: status,
				last_checked: new Date().toISOString(),
				error_message: null
			})
			.eq('domain', domain);
		
		if (updateError) {
			console.error(`更新数据库失败 (${domain}):`, updateError);
		} else {
			console.log(`成功检查域名 ${domain}, 状态: ${status}, 剩余天数: ${daysRemaining}`);
		}
		
	} catch (error: any) {
		console.error(`简化方案检查域名 ${domain} 失败:`, error.message);
		
		// 更新数据库，标记为错误
		await supabase
			.from('ssl_certificates')
			.update({
				status: 'error',
				error_message: error.message || '检查失败',
				last_checked: new Date().toISOString()
			})
			.eq('domain', domain);
	}
}

export default {
	/**
	 * 处理定时任务（Cron Triggers）
	 */
	async scheduled(controller: ScheduledController, env: Env, ctx: ExecutionContext): Promise<void> {
		console.log('开始执行定时检查任务...');
		
		try {
			// 初始化 Supabase 客户端
			const supabase = createClient(env.SUPABASE_URL, env.SUPABASE_ANON_KEY);
			
			// 获取所有需要检查的域名
			const { data: certificates, error } = await supabase
				.from('ssl_certificates')
				.select('domain');
			
			if (error) {
				console.error('获取域名列表失败:', error);
				return;
			}
			
			if (!certificates || certificates.length === 0) {
				console.log('没有需要检查的域名');
				return;
			}
			
			console.log(`找到 ${certificates.length} 个域名需要检查`);
			
			// 批量检查所有域名（使用简化方案更快）
			const checkPromises = certificates.map((cert: { domain: string }) =>
				checkSSLCertificateSimple(cert.domain, supabase)
			);
			
			await Promise.allSettled(checkPromises);
			
			console.log('定时检查任务完成');
		} catch (error) {
			console.error('定时任务执行失败:', error);
		}
	},
	
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
		
		// SSL 证书监控 API 路由
		if (path.startsWith('/api/certificates')) {
			try {
				// 获取所有证书信息
				if (path === '/api/certificates' && request.method === 'GET') {
					const { data, error } = await supabase
						.from('ssl_certificates')
						.select('*')
						.order('created_at', { ascending: false });
					
					if (error) throw error;
					
					return new Response(JSON.stringify({ success: true, data }), {
						headers: corsHeaders
					});
				}
				
				// 添加新域名
				if (path === '/api/certificates' && request.method === 'POST') {
					const body = await request.json() as SSLCertificate;
					
					if (!body.domain) {
						return new Response(JSON.stringify({ success: false, error: '域名不能为空' }), {
							status: 400,
							headers: corsHeaders
						});
					}
					
					// 验证域名格式
					const domainRegex = /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$/i;
					if (!domainRegex.test(body.domain)) {
						return new Response(JSON.stringify({ success: false, error: '域名格式不正确' }), {
							status: 400,
							headers: corsHeaders
						});
					}
					
					// 插入数据库，状态为 pending
					const { data, error } = await supabase
						.from('ssl_certificates')
						.insert([
							{ 
								domain: body.domain.toLowerCase(),
								tags: body.tags || [],
								status: 'pending'
							}
						])
						.select();
					
					if (error) {
						// 检查是否是重复域名错误
						if (error.code === '23505') {
							return new Response(JSON.stringify({ success: false, error: '该域名已存在' }), {
								status: 409,
								headers: corsHeaders
							});
						}
						throw error;
					}
					
					// 立即触发证书检查（异步执行，使用简化方案更快）
					if (data && data[0]) {
						ctx.waitUntil(checkSSLCertificateSimple(data[0].domain, supabase));
					}
					
					return new Response(JSON.stringify({ success: true, data }), {
						status: 201,
						headers: corsHeaders
					});
				}
				
				// 删除域名
				if (path.match(/^\/api\/certificates\/\d+$/) && request.method === 'DELETE') {
					const id = parseInt(path.split('/').pop() || '0');
					
					const { error } = await supabase
						.from('ssl_certificates')
						.delete()
						.eq('id', id);

					console.log('删除域名:', id);
					
					if (error) {
						console.error('删除域名失败:', error);
						throw error;
					}
					
					return new Response(JSON.stringify({ success: true }), {
						headers: corsHeaders
					});
				}
				
				// 更新标签
				if (path.match(/^\/api\/certificates\/\d+\/tags$/) && request.method === 'PUT') {
					const id = parseInt(path.split('/')[3]);
					const body = await request.json() as { tags: string[] };
					
					const { data, error } = await supabase
						.from('ssl_certificates')
						.update({ tags: body.tags || [] })
						.eq('id', id)
						.select();
					
					if (error) throw error;
					
					if (!data || data.length === 0) {
						return new Response(JSON.stringify({ success: false, error: '未找到该域名' }), {
							status: 404,
							headers: corsHeaders
						});
					}
					
					return new Response(JSON.stringify({ success: true, data }), {
						headers: corsHeaders
					});
				}
				
				// 手动触发检查单个域名
				if (path.match(/^\/api\/certificates\/\d+\/check$/) && request.method === 'POST') {
					const id = parseInt(path.split('/')[3]);
					
					// 获取域名信息
					const { data: cert, error: fetchError } = await supabase
						.from('ssl_certificates')
						.select('domain')
						.eq('id', id)
						.single();
					
					if (fetchError || !cert) {
						return new Response(JSON.stringify({ success: false, error: '未找到该域名' }), {
							status: 404,
							headers: corsHeaders
						});
					}
					
					// 异步执行检查（使用简化方案）
					ctx.waitUntil(checkSSLCertificateSimple(cert.domain, supabase));
					
					return new Response(JSON.stringify({ success: true, message: '正在检查证书...' }), {
						headers: corsHeaders
					});
				}
				
				// 手动触发检查所有域名
				if (path === '/api/certificates/check-all' && request.method === 'POST') {
					// 获取所有域名
					const { data: certificates, error: fetchError } = await supabase
						.from('ssl_certificates')
						.select('domain');
					
					if (fetchError) throw fetchError;
					
					// 异步执行所有检查（使用简化方案）
					if (certificates && certificates.length > 0) {
						ctx.waitUntil(
							Promise.allSettled(
								certificates.map((cert: { domain: string }) => 
									checkSSLCertificateSimple(cert.domain, supabase)
								)
							)
						);
					}
					
					return new Response(JSON.stringify({ 
						success: true, 
						message: `正在检查 ${certificates?.length || 0} 个域名的证书...` 
					}), {
						headers: corsHeaders
					});
				}
				
				// 获取所有可用标签（从现有证书中提取）
				if (path === '/api/tags' && request.method === 'GET') {
					const { data, error } = await supabase
						.from('ssl_certificates')
						.select('tags');
					
					if (error) throw error;
					
					// 提取并去重所有标签
					const allTags = new Set<string>();
					data?.forEach((cert: { tags: string[] }) => {
						cert.tags?.forEach(tag => allTags.add(tag));
					});
					
					return new Response(JSON.stringify({ 
						success: true, 
						data: Array.from(allTags).sort() 
					}), {
						headers: corsHeaders
					});
				}
				
				// 如果没有匹配的路由
				return new Response(JSON.stringify({ success: false, error: '无效的请求' }), {
					status: 400,
					headers: corsHeaders
				});
				
			} catch (error: any) {
				console.error('SSL Certificates API Error:', error);
				
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
