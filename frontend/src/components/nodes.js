import { useEffect, useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, reset } from "../store/message";
import axios from "axios";
import Alert from "./alert";
import { formatBytes, getCurrentMonthTraffic, getCurrentYearTraffic, getTrafficOfTodayFromArray } from "../service/service";
import { useNodes } from "../hooks/useQueries";
import { useWebSocket } from "../hooks/useWebSocket";

function Nodes() {
	// 使用 React Query 获取数据
	const { data: singboxNodes = [], isLoading: loading, error: nodesError } = useNodes();
	const [selectedNode, setSelectedNode] = useState(null); // 用于控制模态框显示的节点
	const [customDates, setCustomDates] = useState({}); // 存储每个节点的自定义日期
	
	// 使用 WebSocket hook
	const { status: wsStatus } = useWebSocket();

	const dispatch = useDispatch();
	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);

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

	// 消息自动隐藏
	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(reset({}));
			}, 5000);
		}
	}, [message, dispatch]);
	
	// 错误处理
	useEffect(() => {
		if (nodesError) {
			dispatch(alert({ show: true, content: nodesError.toString() }));
		}
	}, [nodesError, dispatch]);

	// 初始化自定义日期（当节点数据加载完成后）
	// 直接将逻辑内联进 useEffect，避免外部函数引用导致的依赖循环
	useEffect(() => {
		if (singboxNodes.length === 0) return;

		// 计算当月首日，作为默认起始日期
		const getDefaultDate = () => {
			const today = new Date();
			const firstDay = new Date(today.getFullYear(), today.getMonth(), 1);
			return firstDay.toISOString().split('T')[0];
		};

		const initialize = async () => {
			try {
				// 先尝试从后端拉取已保存的自定义日期映射
				const response = await axios.get(
					process.env.REACT_APP_API_HOST + "custom-dates",
					{ headers: { token: loginState.token } }
				);

				const defaultDate = getDefaultDate();
				const indexMapping = {};

				if (response.data && Object.keys(response.data).length > 0) {
					// 后端返回的是 domain_as_id -> date 映射，转换为 index -> date
					singboxNodes.forEach((node, index) => {
						indexMapping[index] = response.data[node.domain_as_id] || defaultDate;
					});
				} else {
					// 没有任何保存记录，所有节点使用默认日期
					singboxNodes.forEach((_, index) => {
						indexMapping[index] = defaultDate;
					});
				}
				setCustomDates(indexMapping);
			} catch (error) {
				console.error('初始化自定义日期失败:', error);
				// 出错时也回退到默认日期，保证 UI 可用
				const defaultDate = getDefaultDate();
				const fallback = {};
				singboxNodes.forEach((_, index) => {
					fallback[index] = defaultDate;
				});
				setCustomDates(fallback);
			}
		};

		initialize();
	}, [singboxNodes, loginState.token]);

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
							{getTrafficOfTodayFromArray(node?.daily_logs)}
						</h3>
					</div>
					<div className="text-center">
						<p className="text-xs text-green-200 mb-1">本月流量</p>
						<h3 className="font-extrabold text-green-400 text-lg">
							{getCurrentMonthTraffic(node?.monthly_logs)}
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
									{getTrafficOfTodayFromArray(node?.daily_logs)}
								</p>
							</div>
							<div className="text-center bg-gray-700 rounded-lg p-4">
								<p className="text-base font-extrabold text-green-200 mb-2">本月流量</p>
								<p className="font-bold text-green-400 text-2xl">
									{getCurrentMonthTraffic(node?.monthly_logs)}
								</p>
							</div>
							<div className="text-center bg-gray-700 rounded-lg p-4">
								<p className="text-base font-extrabold text-purple-200 mb-2">本年流量</p>
								<p className="font-bold text-purple-400 text-2xl">
									{getCurrentYearTraffic(node?.yearly_logs)}
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
				<h1 className="text-3xl font-bold mb-2">节点监控</h1>
				<p className="text-gray-400">管理和监控节点状态</p>
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

			{/* 节点管理部分 */}
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
		</div>
	);
}

export default Nodes;