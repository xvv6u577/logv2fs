// WebSocket 客户端服务
class WebSocketService {
	constructor() {
		this.ws = null;
		this.reconnectAttempts = 0;
		this.maxReconnectAttempts = 5;
		this.reconnectDelay = 1000; // 初始重连延迟1秒
		this.maxReconnectDelay = 30000; // 最大重连延迟30秒
		this.isConnecting = false;
		this.messageHandlers = new Map();
		this.connectionStatus = 'disconnected'; // disconnected, connecting, connected, reconnecting
		this.userID = null;
		this.isAdmin = false;
		
		// 防抖机制
		this.debounceTimeout = null;
		this.pendingRefresh = false;
		this.debounceDelay = 3000; // 3秒防抖延迟
	}

	// 初始化连接
	connect(userID = null, isAdmin = false) {
		if (this.isConnecting || this.ws?.readyState === WebSocket.OPEN) {
			return;
		}

		this.userID = userID;
		this.isAdmin = isAdmin;
		this.isConnecting = true;
		this.connectionStatus = 'connecting';

		// 构建 WebSocket URL
		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const host = window.location.host;
		const wsUrl = `${protocol}//${host}/ws?user_id=${encodeURIComponent(userID || '')}&is_admin=${isAdmin}`;

		try {
			this.ws = new WebSocket(wsUrl);
			this.setupEventHandlers();
		} catch (error) {
			console.error('WebSocket 连接创建失败:', error);
			this.handleConnectionError();
		}
	}

	// 设置事件处理器
	setupEventHandlers() {
		this.ws.onopen = () => {
			console.log('WebSocket 连接已建立');
			this.isConnecting = false;
			this.connectionStatus = 'connected';
			this.reconnectAttempts = 0;
			this.reconnectDelay = 1000;
			
			// 发送连接成功消息
			this.sendMessage({
				type: 'connection_established',
				timestamp: new Date().toISOString()
			});
		};

		this.ws.onmessage = (event) => {
			try {
				const message = JSON.parse(event.data);
				this.handleMessage(message);
			} catch (error) {
				console.error('解析 WebSocket 消息失败:', error);
			}
		};

		this.ws.onclose = (event) => {
			console.log('WebSocket 连接已关闭:', event.code, event.reason);
			this.isConnecting = false;
			this.connectionStatus = 'disconnected';
			
			// 如果不是正常关闭，尝试重连
			if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
				this.scheduleReconnect();
			}
		};

		this.ws.onerror = (error) => {
			console.error('WebSocket 错误:', error);
			this.handleConnectionError();
		};
	}

	// 处理连接错误
	handleConnectionError() {
		this.isConnecting = false;
		this.connectionStatus = 'disconnected';
		
		if (this.reconnectAttempts < this.maxReconnectAttempts) {
			this.scheduleReconnect();
		}
	}

	// 安排重连
	scheduleReconnect() {
		this.connectionStatus = 'reconnecting';
		this.reconnectAttempts++;
		
		const delay = Math.min(this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1), this.maxReconnectDelay);
		
		console.log(`WebSocket 将在 ${delay}ms 后尝试重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
		
		setTimeout(() => {
			if (this.connectionStatus === 'reconnecting') {
				this.connect(this.userID, this.isAdmin);
			}
		}, delay);
	}

	// 发送消息
	sendMessage(message) {
		if (this.ws && this.ws.readyState === WebSocket.OPEN) {
			try {
				this.ws.send(JSON.stringify(message));
			} catch (error) {
				console.error('发送 WebSocket 消息失败:', error);
			}
		} else {
			console.warn('WebSocket 未连接，无法发送消息');
		}
	}

	// 发送心跳
	sendPing() {
		this.sendMessage({
			type: 'ping',
			timestamp: new Date().toISOString()
		});
	}

	// 处理接收到的消息
	handleMessage(message) {
		console.log('收到 WebSocket 消息:', message);

		// 处理心跳响应
		if (message.type === 'pong') {
			return;
		}

		// 处理批量更新消息
		if (message.type === 'batch_update' && message.action === 'batch') {
			console.log(`收到批量更新消息，包含 ${message.data.count} 个事件`);
			this.handleBatchUpdate(message.data.messages);
			return;
		}

		// 调用注册的消息处理器
		const handlers = this.messageHandlers.get(message.type) || [];
		handlers.forEach(handler => {
			try {
				handler(message);
			} catch (error) {
				console.error(`处理消息类型 ${message.type} 时出错:`, error);
			}
		});

		// 触发防抖刷新
		this.debounceRefresh();
	}

	// 处理批量更新消息
	handleBatchUpdate(messages) {
		// 按消息类型分组处理
		const messagesByType = new Map();
		
		messages.forEach(msg => {
			if (!messagesByType.has(msg.type)) {
				messagesByType.set(msg.type, []);
			}
			messagesByType.get(msg.type).push(msg);
		});

		// 为每种消息类型调用处理器
		messagesByType.forEach((msgs, messageType) => {
			const handlers = this.messageHandlers.get(messageType) || [];
			handlers.forEach(handler => {
				msgs.forEach(msg => {
					try {
						handler(msg);
					} catch (error) {
						console.error(`处理批量消息类型 ${messageType} 时出错:`, error);
					}
				});
			});
		});

		// 触发防抖刷新
		this.debounceRefresh();
	}

	// 防抖刷新机制
	debounceRefresh() {
		// 如果已有待执行的刷新，清除它
		if (this.debounceTimeout) {
			clearTimeout(this.debounceTimeout);
		}

		// 标记有待刷新的状态
		this.pendingRefresh = true;

		// 设置新的防抖定时器
		this.debounceTimeout = setTimeout(() => {
			if (this.pendingRefresh) {
				console.log('执行防抖刷新');
				this.executeRefresh();
				this.pendingRefresh = false;
			}
			this.debounceTimeout = null;
		}, this.debounceDelay);
	}

	// 执行刷新操作
	executeRefresh() {
		// 触发页面刷新事件
		const refreshEvent = new CustomEvent('websocket-refresh', {
			detail: { timestamp: new Date() }
		});
		window.dispatchEvent(refreshEvent);

		// 通知所有注册的刷新处理器
		const refreshHandlers = this.messageHandlers.get('refresh') || [];
		refreshHandlers.forEach(handler => {
			try {
				handler({ type: 'refresh', timestamp: new Date() });
			} catch (error) {
				console.error('执行刷新处理器时出错:', error);
			}
		});
	}

	// 立即刷新（跳过防抖）
	immediateRefresh() {
		if (this.debounceTimeout) {
			clearTimeout(this.debounceTimeout);
			this.debounceTimeout = null;
		}
		this.pendingRefresh = false;
		this.executeRefresh();
	}

	// 注册消息处理器
	on(messageType, handler) {
		if (!this.messageHandlers.has(messageType)) {
			this.messageHandlers.set(messageType, []);
		}
		this.messageHandlers.get(messageType).push(handler);
	}

	// 移除消息处理器
	off(messageType, handler) {
		if (this.messageHandlers.has(messageType)) {
			const handlers = this.messageHandlers.get(messageType);
			const index = handlers.indexOf(handler);
			if (index > -1) {
				handlers.splice(index, 1);
			}
		}
	}

	// 获取连接状态
	getConnectionStatus() {
		return this.connectionStatus;
	}

	// 断开连接
	disconnect() {
		this.connectionStatus = 'disconnected';
		this.reconnectAttempts = this.maxReconnectAttempts; // 阻止重连
		
		// 清除防抖定时器
		if (this.debounceTimeout) {
			clearTimeout(this.debounceTimeout);
			this.debounceTimeout = null;
		}
		this.pendingRefresh = false;
		
		if (this.ws) {
			this.ws.close(1000, '用户主动断开连接');
			this.ws = null;
		}
	}

	// 启动心跳
	startHeartbeat() {
		this.heartbeatInterval = setInterval(() => {
			if (this.connectionStatus === 'connected') {
				this.sendPing();
			}
		}, 30000); // 每30秒发送一次心跳
	}

	// 停止心跳
	stopHeartbeat() {
		if (this.heartbeatInterval) {
			clearInterval(this.heartbeatInterval);
			this.heartbeatInterval = null;
		}
	}
}

// 创建全局 WebSocket 服务实例
const websocketService = new WebSocketService();

// 启动心跳
websocketService.startHeartbeat();

export default websocketService; 