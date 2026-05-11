import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useSelector } from 'react-redux';
import * as api from '../api/queries';

/**
 * 自定义 React Query Hooks
 *
 * - 鉴权 token 由全局 axios 拦截器注入，无需在此显式传递。
 * - 查询是否启用仍依赖 token 是否存在，避免未登录态下空跑请求。
 */

// ==================== Query Keys ====================
export const queryKeys = {
	users: ['users'],
	currentUser: (email) => ['currentUser', email],
	nodes: ['nodes'],
	subscriptionNodes: ['subscriptionNodes'],
};

// ==================== 查询 Hooks ====================

export const useUsers = () => {
	const token = useSelector((state) => state.login.token);

	return useQuery({
		queryKey: queryKeys.users,
		queryFn: () => api.fetchUsers(),
		enabled: !!token,
	});
};

export const useCurrentUser = () => {
	const loginState = useSelector((state) => state.login);
	const { jwt, token } = loginState;
	const email = jwt?.email;

	return useQuery({
		queryKey: queryKeys.currentUser(email),
		queryFn: () => api.fetchCurrentUser(email),
		enabled: !!token && !!email,
	});
};

export const useNodes = () => {
	const token = useSelector((state) => state.login.token);

	return useQuery({
		queryKey: queryKeys.nodes,
		queryFn: () => api.fetchNodes(),
		enabled: !!token,
	});
};

export const useSubscriptionNodes = () => {
	const token = useSelector((state) => state.login.token);

	return useQuery({
		queryKey: queryKeys.subscriptionNodes,
		queryFn: () => api.fetchSubscriptionNodes(),
		enabled: !!token,
	});
};

// ==================== 变更 Hooks ====================

export const useAddUser = (options = {}) => {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (userData) => api.addUser({ userData }),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.users });
		},
		...options,
	});
};

export const useUpdateUser = (options = {}) => {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (userData) => api.updateUser({ userData }),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.users });
		},
		...options,
	});
};

export const useDeleteUser = (options = {}) => {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (userData) => api.deleteUser({ userData }),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.users });
		},
		...options,
	});
};

export const useDisableUser = (options = {}) => {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (userData) => api.disableUser({ userData }),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.users });
		},
		...options,
	});
};

export const useEnableUser = (options = {}) => {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (userData) => api.enableUser({ userData }),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.users });
		},
		...options,
	});
};

export const useUpsertNodes = (options = {}) => {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (nodes) => api.upsertNodes({ nodes }),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.nodes });
		},
		...options,
	});
};

export const useDeleteNodes = (options = {}) => {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (nodes) => api.deleteNodes({ nodes }),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.nodes });
		},
		...options,
	});
};
