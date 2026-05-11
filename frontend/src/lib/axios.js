import axios from 'axios';

/**
 * 全局 axios 单例。
 *
 * 设计目的：
 *   1. 统一注入 Authorization: Bearer 头，避免每个组件手写 headers。
 *   2. 统一处理 401/403，触发自动登出与跳转。
 *   3. 设置统一的超时与 baseURL，便于一处调整。
 *
 * 使用示例：
 *   import api from '@/lib/axios';
 *   const { data } = await api.get('user/me');
 */

const baseURL = process.env.REACT_APP_API_HOST || '/v1/';

const api = axios.create({
	baseURL,
	timeout: 15000,
});

/**
 * readToken 从 localStorage 读取 token 并安全反序列化。
 * 兼容历史上把 JWT 当字符串再 JSON.stringify 一遍的写法。
 */
function readToken() {
	const raw = localStorage.getItem('token');
	if (!raw) return '';
	try {
		const parsed = JSON.parse(raw);
		return typeof parsed === 'string' ? parsed : '';
	} catch {
		return raw;
	}
}

api.interceptors.request.use((config) => {
	const token = readToken();
	if (token) {
		config.headers = config.headers || {};
		config.headers.Authorization = `Bearer ${token}`;
	}
	return config;
});

api.interceptors.response.use(
	(response) => response,
	(error) => {
		const status = error?.response?.status;
		if (status === 401 || status === 403) {
			localStorage.removeItem('token');
			if (window.location.pathname !== '/login') {
				window.location.assign('/login');
			}
		}
		return Promise.reject(error);
	}
);

export default api;
