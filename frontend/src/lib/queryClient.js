import { QueryClient } from '@tanstack/react-query';

/**
 * React Query 全局配置
 * 
 * 默认配置说明：
 * - staleTime: 数据被认为"陈旧"的时间（5分钟）
 * - cacheTime: 数据在缓存中保留的时间（10分钟）
 * - refetchOnWindowFocus: 窗口重新获得焦点时是否自动刷新（false，避免频繁请求）
 * - retry: 失败后重试次数（1次）
 * - refetchOnMount: 组件挂载时是否刷新（true）
 */
export const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			staleTime: 5 * 60 * 1000, // 5分钟
			gcTime: 10 * 60 * 1000, // 10分钟（原 cacheTime）
			refetchOnWindowFocus: false,
			retry: 1,
			refetchOnMount: true,
		},
		mutations: {
			retry: 1,
		},
	},
});
