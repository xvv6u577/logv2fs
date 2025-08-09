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

		// 调用注册的消息处理器
		const handlers = this.messageHandlers.get(message.type) || [];
		handlers.forEach(handler => {
			try {
				handler(message);
			} catch (error) {
				console.error(`处理消息类型 ${message.type} 时出错:`, error);
			}
		});
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