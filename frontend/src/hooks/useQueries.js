import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useSelector } from 'react-redux';
import * as api from '../api/queries';

/**
 * 自定义 React Query Hooks
 * 
 * 这些 hooks 封装了常用的查询和变更操作
 * 自动处理 token、错误处理、缓存失效等
 */

// ==================== Query Keys ====================
// 集中管理查询键，避免硬编码

export const queryKeys = {
	users: ['users'],
	currentUser: (email) => ['currentUser', email],
	nodes: ['nodes'],
	subscriptionNodes: ['subscriptionNodes'],
};

// ==================== 查询 Hooks ====================

/**
 * 获取所有用户列表（管理员）
 */
export const useUsers = () => {
	const token = useSelector((state) => state.login.token);
	
	return useQuery({
		queryKey: queryKeys.users,
		queryFn: () => api.fetchUsers(token),
		enabled: !!token, // 只有当 token 存在时才执行查询
	});
};

/**
 * 获取当前用户信息
 */
export const useCurrentUser = () => {
	const loginState = useSelector((state) => state.login);
	const { jwt, token } = loginState;
	
	return useQuery({
		queryKey: queryKeys.currentUser(jwt.Email),
		queryFn: () => api.fetchCurrentUser(jwt.Email, token),
		enabled: !!token && !!jwt.Email, // 只有当 token 和 email 都存在时才执行
	});
};

/**
 * 获取所有节点
 */
export const useNodes = () => {
	const token = useSelector((state) => state.login.token);
	
	return useQuery({
		queryKey: queryKeys.nodes,
		queryFn: () => api.fetchNodes(token),
		enabled: !!token,
	});
};

/**
 * 获取所有活跃的全局节点
 */
export const useSubscriptionNodes = () => {
	const token = useSelector((state) => state.login.token);
	
	return useQuery({
		queryKey: queryKeys.subscriptionNodes,
		queryFn: () => api.fetchSubscriptionNodes(token),
		enabled: !!token,
	});
};

// ==================== 变更 Hooks ====================

/**
 * 添加用户
 * @param {Object} options - useMutation 选项
 * @returns {Object} mutation 对象
 */
export const useAddUser = (options = {}) => {
	const token = useSelector((state) => state.login.token);
	const queryClient = useQueryClient();
	
	return useMutation({
		mutationFn: (userData) => api.addUser({ userData, token }),
		onSuccess: () => {
			// 成功后刷新用户列表
			queryClient.invalidateQueries({ queryKey: queryKeys.users });
		},
		...options,
	});
};

/**
 * 更新用户
 */
export const useUpdateUser = (options = {}) => {
	const token = useSelector((state) => state.login.token);
	const queryClient = useQueryClient();
	
	return useMutation({
		mutationFn: (userData) => api.updateUser({ userData, token }),
		onSuccess: () => {
			// 成功后刷新用户列表
			queryClient.invalidateQueries({ queryKey: queryKeys.users });
		},
		...options,
	});
};

/**
 * 删除用户
 */
export const useDeleteUser = (options = {}) => {
	const token = useSelector((state) => state.login.token);
	const queryClient = useQueryClient();
	
	return useMutation({
		mutationFn: (userData) => api.deleteUser({ userData, token }),
		onSuccess: () => {
			// 成功后刷新用户列表
			queryClient.invalidateQueries({ queryKey: queryKeys.users });
		},
		...options,
	});
};

/**
 * 禁用用户
 */
export const useDisableUser = (options = {}) => {
	const token = useSelector((state) => state.login.token);
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (userData) => api.disableUser({ userData, token }),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.users });
		},
		...options,
	});
};

/**
 * 启用用户
 */
export const useEnableUser = (options = {}) => {
	const token = useSelector((state) => state.login.token);
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (userData) => api.enableUser({ userData, token }),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.users });
		},
		...options,
	});
};

/**
 * 添加/更新节点
 */
export const useUpsertNodes = (options = {}) => {
	const token = useSelector((state) => state.login.token);
	const queryClient = useQueryClient();
	
	return useMutation({
		mutationFn: (nodes) => api.upsertNodes({ nodes, token }),
		onSuccess: () => {
			// 成功后刷新节点列表
			queryClient.invalidateQueries({ queryKey: queryKeys.nodes });
		},
		...options,
	});
};

/**
 * 删除节点
 */
export const useDeleteNodes = (options = {}) => {
	const token = useSelector((state) => state.login.token);
	const queryClient = useQueryClient();
	
	return useMutation({
		mutationFn: (nodes) => api.deleteNodes({ nodes, token }),
		onSuccess: () => {
			// 成功后刷新节点列表
			queryClient.invalidateQueries({ queryKey: queryKeys.nodes });
		},
		...options,
	});
};
