import { useEffect, useState } from "react";
import { alert, reset } from "../store/message";
import { useSelector, useDispatch } from "react-redux";
import axios from "axios";
import Alert from "./alert";

function Mypanel() {
	const [user, setUser] = useState({});

	const dispatch = useDispatch();
	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);
	const rerenderSignal = useSelector((state) => state.rerender);

	// 通用样式类
	const styles = {
		button: "px-4 py-2 rounded-lg font-medium text-sm transition-colors focus:outline-none focus:ring-2",
		buttonPrimary: "bg-blue-600 hover:bg-blue-700 text-white focus:ring-blue-500",
		buttonCopy: "px-2 py-1 bg-gray-600 hover:bg-gray-500 text-white text-xs rounded transition-colors",
		card: "bg-gray-800 rounded-lg shadow-lg hover:shadow-xl transition-shadow",
		badge: "px-2 py-1 rounded-full text-xs font-medium",
		badgeOnline: "bg-green-900 text-green-300",
		badgeOffline: "bg-red-900 text-red-300",
	};

	// 格式化字节数
	const formatBytes = (bytes) => {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	};

	// 格式化时间
	const formatDate = (dateString) => {
		if (!dateString) return "未知";
		try {
			const date = new Date(dateString);
			return date.toLocaleString('zh-CN', {
				year: 'numeric',
				month: '2-digit',
				day: '2-digit',
				hour: '2-digit',
				minute: '2-digit'
			});
		} catch (error) {
			return "无效日期";
		}
	};

	// 复制到剪贴板
	const copyToClipboard = (text) => {
		navigator.clipboard.writeText(text).then(() => {
			dispatch(alert({ show: true, content: "已复制到剪贴板", type: "success" }));
		}).catch(() => {
			dispatch(alert({ show: true, content: "复制失败", type: "error" }));
		});
	};

	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(reset({}));
			}, 5000);
		}
	}, [message, dispatch]);

	useEffect(() => {
		axios
			.get(process.env.REACT_APP_API_HOST + "user/" + loginState.jwt.Email, {
				headers: { token: loginState.token },
			})
			.then((response) => {
				setUser(response.data);
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString(), type: "error" }));
			});
	}, [loginState, dispatch, rerenderSignal]);

	return (
		<div className="min-h-screen bg-gray-900 text-white p-6">
			<Alert 
				message={message.content} 
				type={message.type} 
				shown={message.show} 
				close={() => { dispatch(reset({})); }} 
			/>

			{/* 页面标题 */}
			<div className="mb-8">
				<h1 className="text-3xl font-bold mb-2">我的面板</h1>
				<p className="text-gray-400">查看您的使用情况和订阅信息</p>
			</div>

			{/* 流量统计卡片 */}
			<div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
				{/* 今日流量 */}
				<div className={`${styles.card} p-6`}>
					<div className="flex items-center justify-between mb-4">
						<h3 className="text-lg font-medium text-gray-300">今日流量</h3>
						<div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-full flex items-center justify-center">
							<svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
							</svg>
						</div>
					</div>
					<div className="text-3xl font-bold text-blue-400 mb-2">
						{user?.daily_logs?.length > 0 ? formatBytes(user?.daily_logs?.slice(-1)[0].traffic) : "0 B"}
					</div>
					<p className="text-gray-400 text-sm">今日已使用流量</p>
				</div>

				{/* 本月流量 */}
				<div className={`${styles.card} p-6`}>
					<div className="flex items-center justify-between mb-4">
						<h3 className="text-lg font-medium text-gray-300">本月流量</h3>
						<div className="w-12 h-12 bg-gradient-to-br from-green-500 to-green-600 rounded-full flex items-center justify-center">
							<svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
							</svg>
						</div>
					</div>
					<div className="text-3xl font-bold text-green-400 mb-2">
						{user?.monthly_logs?.length > 0 ? formatBytes(user?.monthly_logs?.slice(-1)[0].traffic) : "0 B"}
					</div>
					<p className="text-gray-400 text-sm">本月已使用流量</p>
				</div>

				{/* 本年流量 */}
				<div className={`${styles.card} p-6`}>
					<div className="flex items-center justify-between mb-4">
						<h3 className="text-lg font-medium text-gray-300">本年流量</h3>
						<div className="w-12 h-12 bg-gradient-to-br from-purple-500 to-purple-600 rounded-full flex items-center justify-center">
							<svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 8v8m-4-5v5m-4-2v2m-2 4h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
							</svg>
						</div>
					</div>
					<div className="text-3xl font-bold text-purple-400 mb-2">
						{user?.yearly_logs?.length > 0 ? formatBytes(user?.yearly_logs?.slice(-1)[0].traffic) : "0 B"}
					</div>
					<p className="text-gray-400 text-sm">本年已使用流量</p>
				</div>
			</div>

			{/* 用户信息和订阅链接 */}
			<div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
				{/* 基本信息 */}
				<div className={`${styles.card} p-6`}>
					<h3 className="text-xl font-bold text-white mb-6">基本信息</h3>
					<div className="space-y-4">
						<div className="flex items-center justify-between">
							<span className="text-gray-400">邮箱ID:</span>
							<div className="flex items-center space-x-2">
								<span className="text-white">{user.email_as_id || "无"}</span>
								{user.email_as_id && (
									<button
										onClick={() => copyToClipboard(user.email_as_id)}
										className={styles.buttonCopy}
										title="复制邮箱ID"
									>
										复制
									</button>
								)}
							</div>
						</div>
						<div className="flex items-center justify-between">
							<span className="text-gray-400">用户名:</span>
							<span className="text-white">{user.name || "未设置"}</span>
						</div>
						<div className="flex items-center justify-between">
							<span className="text-gray-400">状态:</span>
							<span className={`${styles.badge} ${user.status === "plain" ? styles.badgeOnline : styles.badgeOffline}`}>
								{user.status === "plain" ? "正常" : "异常"}
							</span>
						</div>
						<div className="flex items-center justify-between">
							<span className="text-gray-400">总使用:</span>
							<span className="text-white">{user.used ? formatBytes(user.used) : "0 B"}</span>
						</div>
						<div className="flex items-center justify-between">
							<span className="text-gray-400">最后更新:</span>
							<span className="text-white">{formatDate(user.updated_at)}</span>
						</div>
					</div>
				</div>

				{/* 订阅链接 */}
				<div className={`${styles.card} p-6`}>
					<h3 className="text-xl font-bold text-white mb-6">订阅链接</h3>
					<div className="space-y-4">
						{/* Shadowrocket/Surge 订阅 */}
						<div className="flex items-center justify-between p-3 bg-gray-700 rounded-lg">
							<div>
								<span className="text-white font-medium">Shadowrocket/Surge</span>
								<p className="text-gray-400 text-sm">适用于 iOS 客户端</p>
							</div>
							<button
								onClick={() => copyToClipboard(process.env.REACT_APP_FILE_AND_SUB_URL + "/static/" + user.email_as_id)}
								className={`${styles.button} ${styles.buttonPrimary} text-xs`}
								title="复制订阅链接"
							>
								复制链接
							</button>
						</div>

						{/* Verge 订阅 */}
						<div className="flex items-center justify-between p-3 bg-gray-700 rounded-lg">
							<div>
								<span className="text-white font-medium">Verge</span>
								<p className="text-gray-400 text-sm">适用于 Verge 客户端</p>
							</div>
							<button
								onClick={() => copyToClipboard(process.env.REACT_APP_FILE_AND_SUB_URL + "/verge/" + user.email_as_id)}
								className={`${styles.button} ${styles.buttonPrimary} text-xs`}
								title="复制订阅链接"
							>
								复制链接
							</button>
						</div>

						{/* Sing-box 订阅 */}
						<div className="flex items-center justify-between p-3 bg-gray-700 rounded-lg">
							<div>
								<span className="text-white font-medium">Sing-box</span>
								<p className="text-gray-400 text-sm">适用于 Sing-box 客户端</p>
							</div>
							<button
								onClick={() => copyToClipboard(process.env.REACT_APP_FILE_AND_SUB_URL + "/singbox/" + user.email_as_id)}
								className={`${styles.button} ${styles.buttonPrimary} text-xs`}
								title="复制订阅链接"
							>
								复制链接
							</button>
						</div>
					</div>
				</div>
			</div>

			{/* 流量统计表格 */}
			{user?.daily_logs?.length > 0 && (
				<div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
					{/* 月度流量统计 - 显示24个月 */}
					<div className={`${styles.card} p-6`}>
						<h3 className="text-xl font-bold text-white mb-6">月度流量统计（过去24个月）</h3>
						<div className="bg-gray-700 rounded-lg overflow-hidden">
							{user.monthly_logs && user.monthly_logs.length > 0 ? (
								<table className="w-full text-sm">
									<thead className="bg-gray-600">
										<tr>
											<th className="px-4 py-3 text-left">月份</th>
											<th className="px-4 py-3 text-right">流量</th>
										</tr>
									</thead>
									<tbody>
										{user.monthly_logs
											.sort((a, b) => b.month - a.month)
											.slice(0, 24)
											.map((log, idx) => (
												<tr key={idx} className="border-t border-gray-600">
													<td className="px-4 py-3">{log.month}</td>
													<td className="px-4 py-3 text-right font-mono text-green-400">
														{formatBytes(log.traffic)}
													</td>
												</tr>
											))}
									</tbody>
								</table>
							) : (
								<div className="p-4 text-center text-gray-400">
									暂无月度流量数据
								</div>
							)}
						</div>
					</div>

					{/* 日度流量统计 - 显示180天 */}
					<div className={`${styles.card} p-6`}>
						<h3 className="text-xl font-bold text-white mb-6">日流量统计（过去180天）</h3>
						<div className="bg-gray-700 rounded-lg overflow-hidden max-h-96 overflow-y-auto">
							{user.daily_logs && user.daily_logs.length > 0 ? (
								<table className="w-full text-sm">
									<thead className="bg-gray-600 sticky top-0">
										<tr>
											<th className="px-4 py-3 text-left">日期</th>
											<th className="px-4 py-3 text-right">流量</th>
										</tr>
									</thead>
									<tbody>
										{user.daily_logs
											.sort((a, b) => b.date - a.date)
											.slice(0, 180)
											.map((log, idx) => (
												<tr key={idx} className="border-t border-gray-600">
													<td className="px-4 py-3">{log.date}</td>
													<td className="px-4 py-3 text-right font-mono text-blue-400">
														{formatBytes(log.traffic)}
													</td>
												</tr>
											))}
									</tbody>
								</table>
							) : (
								<div className="p-4 text-center text-gray-400">
									暂无日流量数据
								</div>
							)}
						</div>
					</div>
				</div>
			)}
		</div>
	);
}

export default Mypanel;
