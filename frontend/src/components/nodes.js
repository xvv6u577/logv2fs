import { useEffect, useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import axios from "axios";
import { alert, reset, success } from "../store/message";
import Alert from "./alert";
import { doRerender } from "../store/rerender";
import { formatBytes } from "../service/service";
import websocketService from "../service/websocket";

function Nodes() {
	const [singboxNodes, setSingboxNodes] = useState([]);
	const [monitoredDomains, setMonitoredDomains] = useState([]);
	const [loading, setLoading] = useState(true); // 添加加载状态
	const [newDomain, setNewDomain] = useState("");
	const [newRemark, setNewRemark] = useState("");
	const [activeSection, setActiveSection] = useState("nodes"); // 'nodes' or 'domains' 
	const [selectedNode, setSelectedNode] = useState(null); // 用于控制模态框显示的节点
	const [wsStatus, setWsStatus] = useState('disconnected'); // WebSocket 连接状态
	const [customDates, setCustomDates] = useState({}); // 存储每个节点的自定义日期

	const dispatch = useDispatch();
	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);
	const rerenderSignal = useSelector((state) => state.rerender);

	// 通用样式类
	const styles = {
		button: "px-4 py-2 rounded-lg font-medium text-sm transition-colors focus:outline-none focus:ring-2",
		buttonPrimary: "bg-blue-600 hover:bg-blue-700 text-white focus:ring-blue-500",
		buttonSecondary: "bg-gray-600 hover:bg-gray-700 text-white focus:ring-gray-500",
		buttonDanger: "bg-red-600 hover:bg-red-700 text-white focus:ring-red-500",
		input: "w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-transparent",
		select: "px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:ring-2 focus:ring-blue-500",
		card: "bg-gray-800 rounded-lg shadow-lg hover:shadow-xl transition-all duration-200",
		badge: "px-2 py-1 rounded-full text-xs font-medium",
		badgeGreen: "bg-green-900 text-green-300",
		badgeRed: "bg-red-900 text-red-300",
		badgeBlue: "bg-blue-900 text-blue-300",
	};

	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(reset({}));
			}, 5000);
		}
	}, [message, dispatch]);

	// WebSocket 实时数据更新
	useEffect(() => {
		// 连接 WebSocket
		const userID = loginState.jwt?.Email;
		const isAdmin = loginState.jwt?.Role === "admin";
		
		websocketService.connect(userID, isAdmin);
		
		// 监听连接状态变化
		const checkStatus = () => {
			setWsStatus(websocketService.getConnectionStatus());
		};
		
		// 定期检查连接状态
		const statusInterval = setInterval(checkStatus, 1000);
		checkStatus(); // 立即检查一次
		
		// 注册消息处理器
		const handleNodeTrafficUpdate = (message) => {
			console.log('收到节点流量更新:', message);
			
			// 更新节点的流量数据
			if (message.data && message.data.domain_as_id) {
				
				console.log('收到节点流量更新:', message);
			}
		};
		
		// 注册消息处理器
		websocketService.on('node_traffic_update', handleNodeTrafficUpdate);
		
		// 清理函数
		return () => {
			websocketService.off('node_traffic_update', handleNodeTrafficUpdate);
			clearInterval(statusInterval);
		};
	}, [loginState.jwt]);
	
	// 初始加载数据
	useEffect(() => {
		setLoading(true); // 开始加载
		
		// 使用Promise.all同时获取节点数据和域名监控数据
		Promise.all([
			axios.get(process.env.REACT_APP_API_HOST + "c47kr8", {
				headers: { token: loginState.token },
			}),
			axios.get(process.env.REACT_APP_API_HOST + "681p32", {
				headers: { token: loginState.token },
			})
		])
		.then(async ([nodesResponse, domainsResponse]) => {
			setSingboxNodes(nodesResponse.data || []);
			setMonitoredDomains(domainsResponse.data || []);
			
			// 初始化自定义日期
			await initializeCustomDates(nodesResponse.data || []);
			
			setLoading(false); // 加载完成
		})
		.catch((err) => {
			setLoading(false); // 加载完成（即使出错）
			dispatch(alert({ show: true, content: err.toString() }));
		});
	}, [loginState, dispatch, rerenderSignal]);

	const handleAddDomain = (e) => {
		e.preventDefault();
		axios({
			method: "put",
			url: process.env.REACT_APP_API_HOST + "g7302b",
			headers: { token: loginState.token },
			data: monitoredDomains,
		})
			.then((response) => {
				dispatch(success({ show: true, content: response.data.message }));
				dispatch(doRerender({ rerender: !rerenderSignal.rerender }));
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	};

	const addNewDomain = () => {
		if (newDomain.length > 0 && newRemark.length > 0) {
			const tempDomains = monitoredDomains?.filter(item => item.domain === newDomain) || [];
			if (tempDomains.length === 0) {
				setMonitoredDomains([...(monitoredDomains || []), { 
					domain: newDomain, 
					remark: newRemark, 
					days_to_expire: -1, 
					expired_date: "" 
				}]);
			}
			setNewDomain("");
			setNewRemark("");
		} else {
			dispatch(alert({ show: true, content: "域名和备注不能为空" }));
		}
	};

	const removeDomain = (domainToRemove) => {
		setMonitoredDomains(monitoredDomains?.filter(item => item.domain !== domainToRemove) || []);
	};

	// 计算自定义日期流量
	const calculateCustomDateTraffic = (node, customDate) => {
		if (!customDate || !node?.daily_logs) return 0;
		
		const startDate = new Date(customDate);
		const today = new Date();
		let totalTraffic = 0;
		
		// 遍历所有日流量记录
		node.daily_logs.forEach(log => {
			const logDate = new Date(
				log.date.substring(0, 4), // 年
				log.date.substring(4, 6) - 1, // 月 (需要减1，因为Date的月份从0开始)
				log.date.substring(6, 8) // 日
			);
			
			// 如果日志日期在自定义日期之后且在今天之前或等于今天
			if (logDate >= startDate && logDate <= today) {
				totalTraffic += log.traffic || 0;
			}
		});
		
		return totalTraffic;
	};

	// 获取当月首日
	const getFirstDayOfCurrentMonth = () => {
		const today = new Date();
		const firstDay = new Date(today.getFullYear(), today.getMonth(), 1);
		return firstDay.toISOString().split('T')[0];
	};

	// 处理自定义日期变化
	const handleCustomDateChange = (nodeIndex, date) => {
		const newCustomDates = {
			...customDates,
			[nodeIndex]: date
		};
		setCustomDates(newCustomDates);
		
		// 保存到数据库
		saveCustomDateToDatabase(nodeIndex, date);
	};

	// 保存自定义日期到数据库
	const saveCustomDateToDatabase = async (nodeIndex, date) => {
		try {
			const node = singboxNodes[nodeIndex];
			if (!node) return;

			await axios.put(
				process.env.REACT_APP_API_HOST + "custom-date",
				{
					domain_as_id: node.domain_as_id,
					custom_date: date
				},
				{
					headers: { token: loginState.token }
				}
			);
		} catch (error) {
			console.error('保存自定义日期失败:', error);
			dispatch(alert({ show: true, content: "保存自定义日期失败" }));
		}
	};

	// 从数据库加载自定义日期
	const loadCustomDatesFromDatabase = async () => {
		try {
			const response = await axios.get(
				process.env.REACT_APP_API_HOST + "custom-dates",
				{
					headers: { token: loginState.token }
				}
			);
			
			if (response.data) {
				setCustomDates(response.data);
			}
		} catch (error) {
			console.error('加载自定义日期失败:', error);
			// 如果加载失败，使用默认日期
			const defaultDate = getFirstDayOfCurrentMonth();
			const defaultDates = {};
			singboxNodes.forEach((_, index) => {
				defaultDates[index] = defaultDate;
			});
			setCustomDates(defaultDates);
		}
	};

	// 初始化自定义日期
	const initializeCustomDates = async (nodes) => {
		try {
			// 先尝试从数据库加载
			const response = await axios.get(
				process.env.REACT_APP_API_HOST + "custom-dates",
				{
					headers: { token: loginState.token }
				}
			);
			
			if (response.data && Object.keys(response.data).length > 0) {
				// 将domain_as_id映射转换为索引映射
				const indexMapping = {};
				nodes.forEach((node, index) => {
					if (response.data[node.domain_as_id]) {
						indexMapping[index] = response.data[node.domain_as_id];
					} else {
						indexMapping[index] = getFirstDayOfCurrentMonth();
					}
				});
				setCustomDates(indexMapping);
			} else {
				// 如果没有保存的日期，使用默认日期（当月首日）
				const defaultDate = getFirstDayOfCurrentMonth();
				const defaultDates = {};
				nodes.forEach((_, index) => {
					defaultDates[index] = defaultDate;
				});
				setCustomDates(defaultDates);
			}
		} catch (error) {
			console.error('初始化自定义日期失败:', error);
			// 如果加载失败，使用默认日期
			const defaultDate = getFirstDayOfCurrentMonth();
			const defaultDates = {};
			nodes.forEach((_, index) => {
				defaultDates[index] = defaultDate;
			});
			setCustomDates(defaultDates);
		}
	};

	// 节点卡片组件
	const NodeCard = ({ node, index }) => {
		const handleCardClick = () => {
			setSelectedNode({ node, index });
		};

		return (
			<div 
				className={`${styles.card} p-4 cursor-pointer transform transition-all duration-200 hover:scale-105 hover:bg-gray-750 hover:shadow-2xl border border-transparent hover:border-blue-500/20`}
				onClick={handleCardClick}
			>
				<div className="flex items-start justify-between mb-3">
					<div className="flex items-center space-x-2">
						<span className="bg-gray-700 text-gray-300 px-2 py-1 rounded text-xs font-mono">
							#{index + 1}
						</span>
						<span className={`${styles.badge} ${node.status === "active" ? styles.badgeGreen : styles.badgeRed}`}>
							{node.status === "active" ? "活跃" : "离线"}
						</span>
					</div>
				</div>

				<h3 className="text-xl font-semibold text-white mb-2 truncate">{node.remark}</h3>
				<p className="text-gray-400 mb-2 text-xs truncate">{node.domain_as_id}</p>
				<p className="text-xs text-blue-400 mb-3 opacity-70 hover:opacity-100 transition-opacity">
					💡 点击查看详细数据
				</p>

				<div className="grid grid-cols-1 gap-3">
					<div className="text-center">
						<p className="text-xs text-blue-200 mb-1">今日流量</p>
						<h3 className="font-extrabold text-blue-400 text-lg">
							{(() => {
								const today = new Date();
								const todayStr = today.getFullYear().toString() + 
												(today.getMonth() + 1).toString().padStart(2, '0') + 
												today.getDate().toString().padStart(2, '0');
								const todayLog = node?.daily_logs?.find(log => log.date === todayStr);
								return todayLog ? formatBytes(todayLog.traffic) : "0";
							})()}
						</h3>
					</div>
					<div className="text-center">
						<p className="text-xs text-green-200 mb-1">本月流量</p>
						<h3 className="font-extrabold text-green-400 text-lg">
							{(() => {
								const today = new Date();
								const currentMonth = today.getFullYear().toString() + 
													(today.getMonth() + 1).toString().padStart(2, '0');
								const monthLog = node?.monthly_logs?.find(log => log.month === currentMonth);
								return monthLog ? formatBytes(monthLog.traffic) : "0";
							})()}
						</h3>
					</div>
					{/* 自定义日期流量 - 只读显示 */}
					<div className="text-center">
						<p className="text-xs text-purple-200 mb-1">
							{customDates[index] ? `自从 ${customDates[index]} 流量` : '自定义日期流量'}
						</p>
						<h3 className="font-extrabold text-purple-400 text-lg">
							{formatBytes(calculateCustomDateTraffic(node, customDates[index]))}
						</h3>
					</div>
				</div>
			</div>
		);
	};

	// 悬浮式节点详情模态框
	const NodeDetailModal = ({ nodeData, onClose }) => {
		if (!nodeData) return null;
		
		const { node, index } = nodeData;
		
		const handleOverlayClick = (e) => {
			if (e.target === e.currentTarget) {
				onClose();
			}
		};

		return (
			<div 
				className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
				onClick={handleOverlayClick}
			>
				<div className="bg-gray-800 rounded-lg shadow-2xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
					{/* 模态框头部 */}
					<div className="flex items-center justify-between p-6 border-b border-gray-700">
						<div className="flex items-center space-x-3">
							<span className="bg-gray-700 text-gray-300 px-3 py-1 rounded text-sm font-mono">
								#{index + 1}
							</span>
							<span className={`${styles.badge} ${node.status === "active" ? styles.badgeGreen : styles.badgeRed}`}>
								{node.status === "active" ? "活跃" : "离线"}
							</span>
							<h2 className="text-xl font-bold text-white">{node.domain_as_id}</h2>
						</div>
						<button
							onClick={onClose}
							className="text-gray-400 hover:text-white transition-colors p-2"
						>
							<svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
							</svg>
						</button>
					</div>

					{/* 模态框内容 */}
					<div className="p-6">
						<p className="text-gray-400 mb-6">{node.remark}</p>

						{/* 流量概览 */}
						<div className="grid grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
							<div className="text-center bg-gray-700 rounded-lg p-4">
								<p className="text-base font-extrabold text-blue-200 mb-2">今日流量</p>
								<p className="font-bold text-blue-400 text-2xl">
									{(() => {
										const today = new Date();
										const todayStr = today.getFullYear().toString() + 
														(today.getMonth() + 1).toString().padStart(2, '0') + 
														today.getDate().toString().padStart(2, '0');
										const todayLog = node?.daily_logs?.find(log => log.date === todayStr);
										return todayLog ? formatBytes(todayLog.traffic) : "0";
									})()}
								</p>
							</div>
							<div className="text-center bg-gray-700 rounded-lg p-4">
								<p className="text-base font-extrabold text-green-200 mb-2">本月流量</p>
								<p className="font-bold text-green-400 text-2xl">
									{(() => {
										const today = new Date();
										const currentMonth = today.getFullYear().toString() + 
															(today.getMonth() + 1).toString().padStart(2, '0');
										const monthLog = node?.monthly_logs?.find(log => log.month === currentMonth);
										return monthLog ? formatBytes(monthLog.traffic) : "0";
									})()}
								</p>
							</div>
							<div className="text-center bg-gray-700 rounded-lg p-4">
								<p className="text-base font-extrabold text-purple-200 mb-2">本年流量</p>
								<p className="font-bold text-purple-400 text-2xl">
									{(() => {
										const currentYear = new Date().getFullYear().toString();
										const yearLog = node?.yearly_logs?.find(log => log.year === currentYear);
										return yearLog ? formatBytes(yearLog.traffic) : "0";
									})()}
								</p>
							</div>
							{/* 自定义日期流量 */}
							<div className="text-center bg-gray-700 rounded-lg p-4">
								<div className="mb-3">
									<input
										type="date"
										value={customDates[index] || ''}
										onChange={(e) => handleCustomDateChange(index, e.target.value)}
										className="w-full px-3 py-2 text-sm bg-gray-600 border border-gray-500 rounded text-white focus:ring-2 focus:ring-purple-500 focus:border-transparent"
										placeholder="选择起始日期"
									/>
								</div>
								<p className="text-base font-extrabold text-orange-200 mb-2">
									{customDates[index] ? `自从 ${customDates[index]} 流量` : '自定义日期流量'}
								</p>
								<p className="font-bold text-orange-400 text-2xl">
									{formatBytes(calculateCustomDateTraffic(node, customDates[index]))}
								</p>
							</div>
						</div>

						{/* 详细流量统计 */}
						<div className="grid grid-cols-2 gap-6">
							{/* 月度流量统计 */}
							<div>
								<h4 className="text-lg font-medium text-gray-300 mb-4">月度流量统计（过去12个月）</h4>
								<div className="bg-gray-700 rounded-lg overflow-hidden max-h-80 overflow-y-auto">
									{node?.monthly_logs && (node.monthly_logs?.length || 0) > 0 ? (
										<table className="w-full text-sm">
											<thead className="bg-gray-600 sticky top-0">
												<tr>
													<th className="px-4 py-3 text-left">月份</th>
													<th className="px-4 py-3 text-right">流量</th>
												</tr>
											</thead>
											<tbody>
												{node.monthly_logs
													?.sort((a, b) => b.month - a.month)
													?.slice(0, 12)
													?.map((item, idx) => (
														<tr key={idx} className="border-t border-gray-600 hover:bg-gray-650">
															<td className="px-4 py-3">{item.month}</td>
															<td className="px-4 py-3 text-right font-mono text-green-400">
																{formatBytes(item.traffic)}
															</td>
														</tr>
													))}
											</tbody>
										</table>
									) : (
										<div className="p-6 text-center text-gray-400">
											暂无月度流量数据
										</div>
									)}
								</div>
							</div>

							{/* 日度流量统计 */}
							<div>
								<h4 className="text-lg font-medium text-gray-300 mb-4">日流量统计（过去30天）</h4>
								<div className="bg-gray-700 rounded-lg overflow-hidden max-h-80 overflow-y-auto">
									{node?.daily_logs && (node.daily_logs?.length || 0) > 0 ? (
										<table className="w-full text-sm">
											<thead className="bg-gray-600 sticky top-0">
												<tr>
													<th className="px-4 py-3 text-left">日期</th>
													<th className="px-4 py-3 text-right">流量</th>
												</tr>
											</thead>
											<tbody>
												{node.daily_logs
													?.sort((a, b) => b.date - a.date)
													?.slice(0, 30)
													?.map((item, idx) => (
														<tr key={idx} className="border-t border-gray-600 hover:bg-gray-650">
															<td className="px-4 py-3">{item.date}</td>
															<td className="px-4 py-3 text-right font-mono text-blue-400">
																{formatBytes(item.traffic)}
															</td>
														</tr>
													))}
											</tbody>
										</table>
									) : (
										<div className="p-6 text-center text-gray-400">
											暂无日流量数据
										</div>
									)}
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		);
	};

	// 域名卡片组件
	const DomainCard = ({ domain, index }) => (
		<div className={`${styles.card} p-6 relative`}>
			<button 
				className="absolute top-4 right-4 text-gray-400 hover:text-red-400 transition-colors"
				onClick={() => removeDomain(domain.domain)}
			>
				<svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>

			<div className="mb-4">
				<h2 className="text-xl font-bold text-blue-300 mb-2">{domain.remark}</h2>
				<h3 className="text-lg font-semibold text-white mb-1">{domain.domain}</h3>
			</div>

			<div className="text-center">
				<div className="text-3xl font-bold text-white mb-2">
					{domain.days_to_expire}天
				</div>
				<p className="text-gray-400 text-sm">
					到期时间: {domain.expired_date}
				</p>
			</div>
		</div>
	);

	return (
		<div className="min-h-screen bg-gray-900 text-white p-6">
			<Alert 
				message={message.content} 
				type={message.type} 
				shown={message.show} 
				close={() => { dispatch(reset({})); }} 
			/>

			{/* 节点详情模态框 */}
			<NodeDetailModal 
				nodeData={selectedNode} 
				onClose={() => setSelectedNode(null)} 
			/>

			{/* 页面标题 */}
			<div className="mb-8">
				<h1 className="text-3xl font-bold mb-2">节点管理</h1>
				<p className="text-gray-400">管理节点状态和域名监控</p>
				{/* WebSocket 连接状态指示器 */}
				<div className="flex items-center space-x-2 mt-2">
					<div className={`w-2 h-2 rounded-full ${
						wsStatus === 'connected' ? 'bg-green-500' :
						wsStatus === 'connecting' ? 'bg-yellow-500' :
						wsStatus === 'reconnecting' ? 'bg-orange-500' :
						'bg-red-500'
					}`}></div>
					<span className={`text-xs ${
						wsStatus === 'connected' ? 'text-green-400' :
						wsStatus === 'connecting' ? 'text-yellow-400' :
						wsStatus === 'reconnecting' ? 'text-orange-400' :
						'text-red-400'
					}`}>
						{wsStatus === 'connected' ? '实时流量监控已连接' :
						 wsStatus === 'connecting' ? '正在连接...' :
						 wsStatus === 'reconnecting' ? '正在重连...' :
						 '连接已断开'}
					</span>
				</div>
			</div>

			{/* 导航标签 */}
			<div className="flex space-x-4 mb-8">
				<button
					onClick={() => setActiveSection("nodes")}
					className={`px-6 py-2 rounded-lg font-medium transition-colors ${
						activeSection === "nodes" 
							? "bg-blue-600 text-white" 
							: "bg-gray-700 text-gray-300 hover:bg-gray-600"
					}`}
				>
					节点监控 ({singboxNodes?.length || 0})
				</button>
				<button
					onClick={() => setActiveSection("domains")}
					className={`px-6 py-2 rounded-lg font-medium transition-colors ${
						activeSection === "domains" 
							? "bg-blue-600 text-white" 
							: "bg-gray-700 text-gray-300 hover:bg-gray-600"
					}`}
				>
					域名监控 ({monitoredDomains?.length || 0})
				</button>
			</div>

			{/* 节点管理部分 */}
			{activeSection === "nodes" && (
				<div>
					{/* 节点列表 */}
					{loading ? (
						// 加载中状态
						<div className={`${styles.card} p-8 text-center`}>
							<div className="flex flex-col items-center">
								<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mb-4"></div>
								<h3 className="text-lg font-medium text-gray-300 mb-2">加载中...</h3>
								<p className="text-gray-400">正在获取节点数据</p>
							</div>
						</div>
					) : (
						<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
							{(singboxNodes?.length || 0) === 0 ? (
								<div className={`${styles.card} p-8 text-center col-span-full`}>
									<svg className="mx-auto h-12 w-12 text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
									</svg>
									<h3 className="text-lg font-medium text-gray-300 mb-2">暂无节点</h3>
									<p className="text-gray-400">等待节点数据加载</p>
								</div>
							) : (
								singboxNodes?.map((node, index) => (
									<NodeCard key={index} node={node} index={index} />
								)) || []
							)}
						</div>
					)}
				</div>
			)}

			{/* 域名监控部分 */}
			{activeSection === "domains" && (
				<div>
					{/* 添加域名表单 */}
					<div className={`${styles.card} p-6 mb-8`}>
						<h3 className="text-lg font-semibold text-white mb-4">添加域名监控</h3>
						<form onSubmit={handleAddDomain}>
							<div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
								<input
									type="text"
									placeholder="域名"
									value={newDomain}
									onChange={(e) => setNewDomain(e.target.value.replace(/\s/g, ""))}
									className={styles.input}
								/>
								<input
									type="text"
									placeholder="备注"
									value={newRemark}
									onChange={(e) => setNewRemark(e.target.value.replace(/\s/g, ""))}
									className={styles.input}
								/>
								<button
									type="button"
									onClick={addNewDomain}
									className={`${styles.button} ${styles.buttonPrimary}`}
								>
									添加域名
								</button>
							</div>
							<button
								type="submit"
								className={`${styles.button} ${styles.buttonSecondary}`}
							>
								更新域名监控
							</button>
						</form>
					</div>

					{/* 域名列表 */}
					{loading ? (
						// 加载中状态
						<div className={`${styles.card} p-8 text-center`}>
							<div className="flex flex-col items-center">
								<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mb-4"></div>
								<h3 className="text-lg font-medium text-gray-300 mb-2">加载中...</h3>
								<p className="text-gray-400">正在获取域名监控数据</p>
							</div>
						</div>
					) : (monitoredDomains?.length || 0) > 0 ? (
						<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
							{monitoredDomains?.map((domain, index) => (
								<DomainCard key={index} domain={domain} index={index} />
							)) || []}
						</div>
					) : (
						<div className={`${styles.card} p-8 text-center`}>
							<svg className="mx-auto h-12 w-12 text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9v-9m0-9v9" />
							</svg>
							<h3 className="text-lg font-medium text-gray-300 mb-2">暂无域名监控</h3>
							<p className="text-gray-400">添加域名开始监控到期时间</p>
						</div>
					)}
				</div>
			)}
		</div>
	);
}

export default Nodes;