import api from '../lib/axios';

/**
 * API 查询函数集合
 *
 * 这些函数被 React Query 的 useQuery / useMutation 调用。
 * 每个函数返回 Promise，解析为后端响应数据。
 *
 * 安全说明：所有请求统一通过 src/lib/axios.js 的拦截器塞入 Authorization: Bearer，
 * 因此本文件不再显式传 token 参数；保留 token 形参仅为向后兼容旧调用点，可在迁移完成后删除。
 */

// ==================== 用户相关 ====================

export const fetchUsers = async () => {
	const { data } = await api.get('n778cf');
	return data;
};

export const fetchCurrentUser = async (email) => {
	const { data } = await api.get(`user/${email}`);
	return data;
};

// ==================== 节点相关 ====================

export const fetchNodes = async () => {
	const { data } = await api.get('c47kr8');
	return data;
};

export const fetchSubscriptionNodes = async () => {
	const { data } = await api.get('subscription-nodes');
	return data;
};

// ==================== Mutation 函数 ====================

export const addUser = async ({ userData }) => {
	const { data } = await api.post('signup', userData);
	return data;
};

/**
 * 更新用户信息
 * 后端真实路由：POST /v1/edit/:name，路径段为 email_as_id。
 */
export const updateUser = async ({ userData }) => {
	const { data } = await api.post(`edit/${userData.email_as_id}`, userData, {
		headers: { 'Content-Type': 'application/json' },
	});
	return data;
};

/**
 * 删除用户
 * 后端真实路由：GET /v1/deluser/:name
 */
export const deleteUser = async ({ userData }) => {
	const { data } = await api.get(`deluser/${userData.email_as_id}`);
	return data;
};

export const disableUser = async ({ userData }) => {
	const { data } = await api.put(`disableuser/${userData.email_as_id}`, {}, {
		headers: { 'Content-Type': 'application/json' },
	});
	return data;
};

export const enableUser = async ({ userData }) => {
	const { data } = await api.put(`enableuser/${userData.email_as_id}`, {}, {
		headers: { 'Content-Type': 'application/json' },
	});
	return data;
};

export const upsertNodes = async ({ nodes }) => {
	const { data } = await api.put('upsert-nodes', nodes);
	return data;
};

/**
 * 删除节点（保留旧 API 形态以兼容历史调用点；
 * 当前后端路由表中没有 7tpxya，调用前请先在 routers/authorized.go 注册）。
 */
export const deleteNodes = async ({ nodes }) => {
	const { data } = await api.delete('7tpxya', { data: nodes });
	return data;
};
