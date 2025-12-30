import { useEffect, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useSelector } from 'react-redux';
import websocketService from '../service/websocket';
import { queryKeys } from './useQueries';

/**
 * WebSocket Hook
 * 
 * 自动管理 WebSocket 连接，并在收到流量更新时触发 React Query 缓存失效
 * 
 * @returns {Object} { status: string, isConnected: boolean }
 */
export const useWebSocket = () => {
	const [status, setStatus] = useState('disconnected');
	const queryClient = useQueryClient();
	const loginState = useSelector((state) => state.login);

	useEffect(() => {
		const userID = loginState.jwt?.Email;
		const isAdmin = loginState.jwt?.Role === 'admin';

		if (!userID) {
			return;
		}

		// 设置 queryClient 到 WebSocket 服务
		websocketService.setQueryClient(queryClient);

		// 连接 WebSocket
		websocketService.connect(userID, isAdmin);

		// 监听 WebSocket 消息并触发查询失效
		const handleTrafficUpdate = (message) => {
			console.log('收到流量更新，触发数据刷新:', message);
			
			// 根据消息类型触发不同的查询失效
			if (message.type === 'traffic_update') {
				// 刷新所有与流量相关的数据
				websocketService.invalidateQueries([
					queryKeys.users,
					queryKeys.nodes,
				]);
				
				// 如果有具体的 email，也刷新该用户的数据
				if (message.email) {
					websocketService.invalidateQueries([
						queryKeys.currentUser(message.email),
					]);
				}
			} else if (message.type === 'node_update') {
				// 节点更新
				websocketService.invalidateQueries([queryKeys.nodes]);
			} else if (message.type === 'user_update') {
				// 用户更新
				websocketService.invalidateQueries([queryKeys.users]);
			}
		};

		// 注册流量更新处理器
		websocketService.on('traffic_update', handleTrafficUpdate);
		websocketService.on('node_update', handleTrafficUpdate);
		websocketService.on('user_update', handleTrafficUpdate);

		// 定期检查连接状态
		const statusInterval = setInterval(() => {
			const currentStatus = websocketService.getConnectionStatus();
			setStatus(currentStatus);
		}, 1000);

		// 立即检查一次
		setStatus(websocketService.getConnectionStatus());

		// 清理函数
		return () => {
			clearInterval(statusInterval);
			websocketService.off('traffic_update', handleTrafficUpdate);
			websocketService.off('node_update', handleTrafficUpdate);
			websocketService.off('user_update', handleTrafficUpdate);
		};
	}, [loginState.jwt, queryClient]);

	return {
		status,
		isConnected: status === 'connected',
	};
};
