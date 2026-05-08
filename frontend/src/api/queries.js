import axios from 'axios';

/**
 * API 查询函数集合
 * 
 * 这些函数被 React Query 的 useQuery 调用
 * 每个函数都返回一个 Promise，解析为实际的数据
 */

// ==================== 用户相关 ====================

/**
 * 获取所有用户列表（管理员）
 * @param {string} token - JWT token
 * @returns {Promise<Array>} 用户列表
 */
export const fetchUsers = async (token) => {
	const response = await axios.get(
		`${process.env.REACT_APP_API_HOST}n778cf`,
		{ headers: { token } }
	);
	return response.data;
};

/**
 * 获取当前用户信息
 * @param {string} email - 用户邮箱
 * @param {string} token - JWT token
 * @returns {Promise<Object>} 用户信息
 */
export const fetchCurrentUser = async (email, token) => {
	const response = await axios.get(
		`${process.env.REACT_APP_API_HOST}user/${email}`,
		{ headers: { token } }
	);
	return response.data;
};

// ==================== 节点相关 ====================

/**
 * 获取所有 Sing-box 节点
 * @param {string} token - JWT token
 * @returns {Promise<Array>} 节点列表
 */
export const fetchNodes = async (token) => {
	const response = await axios.get(
		`${process.env.REACT_APP_API_HOST}c47kr8`,
		{ headers: { token } }
	);
	return response.data;
};

/**
 * 获取所有节点列表
 * @param {string} token - JWT token
 * @returns {Promise<Array>} 节点列表
 */
export const fetchSubscriptionNodes = async (token) => {
	const response = await axios.get(
		`${process.env.REACT_APP_API_HOST}subscription-nodes`,
		{ headers: { token } }
	);

	return response.data;
};

// ==================== Mutation 函数 ====================

/**
 * 添加新用户
 * @param {Object} params
 * @param {Object} params.userData - 用户数据
 * @param {string} params.token - JWT token
 * @returns {Promise<Object>} 响应数据
 */
export const addUser = async ({ userData, token }) => {
	const response = await axios.post(
		`${process.env.REACT_APP_API_HOST}signup`,
		userData,
		{ headers: { token } }
	);
	return response.data;
};

/**
 * 更新用户信息
 * 后端真实路由：POST /v1/edit/:name，路径段为 email_as_id
 * @param {Object} params
 * @param {Object} params.userData - 用户数据，必须包含 email_as_id
 * @param {string} params.token - JWT token
 * @returns {Promise<Object>} 响应数据
 */
export const updateUser = async ({ userData, token }) => {
	const response = await axios.post(
		`${process.env.REACT_APP_API_HOST}edit/${userData.email_as_id}`,
		userData,
		{ headers: { token, 'Content-Type': 'application/json' } }
	);
	return response.data;
};

/**
 * 删除用户
 * 后端真实路由：GET /v1/deluser/:name
 * @param {Object} params
 * @param {Object} params.userData - 必须包含 email_as_id
 * @param {string} params.token - JWT token
 * @returns {Promise<Object>} 响应数据
 */
export const deleteUser = async ({ userData, token }) => {
	const response = await axios.get(
		`${process.env.REACT_APP_API_HOST}deluser/${userData.email_as_id}`,
		{ headers: { token } }
	);
	return response.data;
};

/**
 * 禁用用户
 * 后端真实路由：PUT /v1/disableuser/:name
 * @param {Object} params
 * @param {Object} params.userData - 必须包含 email_as_id
 * @param {string} params.token - JWT token
 * @returns {Promise<Object>} 响应数据
 */
export const disableUser = async ({ userData, token }) => {
	const response = await axios.put(
		`${process.env.REACT_APP_API_HOST}disableuser/${userData.email_as_id}`,
		{},
		{ headers: { token, 'Content-Type': 'application/json' } }
	);
	return response.data;
};

/**
 * 启用用户
 * 后端真实路由：PUT /v1/enableuser/:name
 * @param {Object} params
 * @param {Object} params.userData - 必须包含 email_as_id
 * @param {string} params.token - JWT token
 * @returns {Promise<Object>} 响应数据
 */
export const enableUser = async ({ userData, token }) => {
	const response = await axios.put(
		`${process.env.REACT_APP_API_HOST}enableuser/${userData.email_as_id}`,
		{},
		{ headers: { token, 'Content-Type': 'application/json' } }
	);
	return response.data;
};

/**
 * 添加节点
 * @param {Object} params
 * @param {Array} params.nodes - 节点数据
 * @param {string} params.token - JWT token
 * @returns {Promise<Object>} 响应数据
 */
export const upsertNodes = async ({ nodes, token }) => {
	const response = await axios.put(
		`${process.env.REACT_APP_API_HOST}upsert-nodes`,
		nodes,
		{ headers: { token } }
	);
	return response.data;
};

/**
 * 删除节点
 * @param {Object} params
 * @param {Array} params.nodes - 节点数据
 * @param {string} params.token - JWT token
 * @returns {Promise<Object>} 响应数据
 */
export const deleteNodes = async ({ nodes, token }) => {
	const response = await axios.delete(
		`${process.env.REACT_APP_API_HOST}7tpxya`,
		{
			headers: { token },
			data: nodes,
		}
	);
	return response.data;
};
