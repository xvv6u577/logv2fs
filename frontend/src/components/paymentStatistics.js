import React, { useState, useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { alert } from '../store/message';
import axios from 'axios';

const PaymentStatistics = () => {
	const [statistics, setStatistics] = useState(null);
	const [loading, setLoading] = useState(false);
	const [statType, setStatType] = useState('monthly');
	const [dateRange, setDateRange] = useState({
		startDate: new Date(new Date().getFullYear(), 0, 1).toISOString().split('T')[0], // 今年第一天
		endDate: new Date().toISOString().split('T')[0], // 今天
	});

	const dispatch = useDispatch();
	const loginState = useSelector((state) => state.login);

	// 通用样式类
	const styles = {
		container: "min-h-screen bg-gray-900 text-white p-6",
		card: "bg-gray-800 rounded-lg shadow-lg p-6",
		input: "px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-transparent",
		select: "px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:ring-2 focus:ring-blue-500",
		label: "block text-sm font-medium text-gray-300 mb-2",
		button: "px-4 py-2 rounded-lg font-medium text-sm transition-colors focus:outline-none focus:ring-2",
		buttonPrimary: "bg-blue-600 hover:bg-blue-700 text-white focus:ring-blue-500",
		buttonSecondary: "bg-gray-600 hover:bg-gray-700 text-white focus:ring-gray-500",
		badge: "px-2 py-1 rounded-full text-xs font-medium",
		statCard: "bg-gray-700 rounded-lg p-4 text-center",
		table: "w-full text-sm text-left",
		tableHeader: "text-xs text-gray-400 uppercase bg-gray-700",
		tableRow: "border-b border-gray-700 hover:bg-gray-700",
	};

	// 格式化金额
	const formatAmount = (amount) => {
		return new Intl.NumberFormat('zh-CN', {
			style: 'currency',
			currency: 'CNY',
		}).format(amount);
	};

	// 格式化日期
	const formatDate = (dateStr, type) => {
		if (type === 'daily') {
			// 20240101 -> 2024-01-01
			return `${dateStr.slice(0, 4)}-${dateStr.slice(4, 6)}-${dateStr.slice(6, 8)}`;
		} else if (type === 'monthly') {
			// 202401 -> 2024年1月
			return `${dateStr.slice(0, 4)}年${parseInt(dateStr.slice(4, 6))}月`;
		} else if (type === 'yearly') {
			// 2024 -> 2024年
			return `${dateStr}年`;
		}
		return dateStr;
	};

	// 获取统计数据
	const fetchStatistics = () => {
		setLoading(true);

		const params = {
			type: statType,
			start_date: dateRange.startDate,
			end_date: dateRange.endDate,
		};

		axios
			.get(process.env.REACT_APP_API_HOST + "payment/statistics", {
				headers: { token: loginState.token },
				params: params,
			})
			.then((response) => {
				setStatistics(response.data);
			})
			.catch((err) => {
				if (err.response) {
					dispatch(alert({ show: true, content: err.response.data.error || "获取统计数据失败" }));
				} else {
					dispatch(alert({ show: true, content: "网络错误: " + err.toString() }));
				}
			})
			.finally(() => {
				setLoading(false);
			});
	};

	// 初始加载
	useEffect(() => {
		fetchStatistics();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [statType, dateRange]);

	// 统计卡片组件
	const StatCard = ({ title, value, subValue, color }) => (
		<div className={`${styles.statCard} ${color}`}>
			<h3 className="text-gray-400 text-sm mb-2">{title}</h3>
			<p className="text-2xl font-bold">{value}</p>
			{subValue && <p className="text-gray-400 text-sm mt-1">{subValue}</p>}
		</div>
	);

	return (
		<div className={styles.container}>
			<div className="max-w-7xl mx-auto">
				{/* 页面标题 */}
				<div className="mb-8">
					<h1 className="text-3xl font-bold mb-2">费用统计</h1>
					<p className="text-gray-400">查看收费统计数据</p>
				</div>

				{/* 控制栏 */}
				<div className={`${styles.card} mb-6`}>
					<div className="grid grid-cols-1 md:grid-cols-4 gap-4">
						{/* 统计类型 */}
						<div>
							<label className={styles.label}>统计类型</label>
							<select
								value={statType}
								onChange={(e) => setStatType(e.target.value)}
								className={styles.select}
							>
								<option value="daily">按日统计</option>
								<option value="monthly">按月统计</option>
								<option value="yearly">按年统计</option>
								<option value="overall">综合统计</option>
							</select>
						</div>

						{/* 开始日期 */}
						<div>
							<label className={styles.label}>开始日期</label>
							<input
								type="date"
								value={dateRange.startDate}
								onChange={(e) => setDateRange({ ...dateRange, startDate: e.target.value })}
								className={styles.input}
							/>
						</div>

						{/* 结束日期 */}
						<div>
							<label className={styles.label}>结束日期</label>
							<input
								type="date"
								value={dateRange.endDate}
								onChange={(e) => setDateRange({ ...dateRange, endDate: e.target.value })}
								className={styles.input}
							/>
						</div>

						{/* 刷新按钮 */}
						<div className="flex items-end">
							<button
								onClick={fetchStatistics}
								className={`${styles.button} ${styles.buttonPrimary} w-full`}
								disabled={loading}
							>
								{loading ? '加载中...' : '刷新数据'}
							</button>
						</div>
					</div>
				</div>

				{/* 统计概览 */}
				{statistics && (
					<div>
						{/* 总体统计卡片 */}
						<div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
							<StatCard
								title="总收入"
								value={formatAmount(statistics.total_amount || 0)}
								color="bg-gradient-to-br from-green-800 to-green-700"
							/>
							<StatCard
								title="缴费次数"
								value={statistics.payment_count || 0}
								subValue="次"
								color="bg-gradient-to-br from-blue-800 to-blue-700"
							/>
							<StatCard
								title="日期范围"
								value={statistics.date_range || ''}
								color="bg-gradient-to-br from-purple-800 to-purple-700"
							/>
						</div>

						{/* 综合统计 */}
						{statType === 'overall' && (
							<div className="p-6 space-y-6">
								{/* 最近日统计 */}
								{statistics.daily_stats && statistics.daily_stats.length > 0 && (
									<div>
										<h4 className="text-md font-medium text-gray-300 mb-3">最近日统计</h4>
										<div className="grid grid-cols-1 md:grid-cols-3 gap-4">
											{statistics.daily_stats.slice(0, 6).map((stat, index) => (
												<div key={index} className="bg-gray-700 rounded-lg p-3">
													<p className="text-sm text-gray-400">{formatDate(stat.date, 'daily')}</p>
													<p className="text-lg font-semibold text-green-400 mt-1">
														{formatAmount(stat.total_amount)}
													</p>
													<p className="text-xs text-gray-500">
														{stat.payment_count} 笔 / {stat.user_count} 人
													</p>
												</div>
											))}
										</div>
									</div>
								)}

								{/* 最近月统计 */}
								{statistics.monthly_stats && statistics.monthly_stats.length > 0 && (
									<div>
										<h4 className="text-md font-medium text-gray-300 mb-3">最近月统计</h4>
										<div className="grid grid-cols-1 md:grid-cols-3 gap-4">
											{statistics.monthly_stats.slice(0, 6).map((stat, index) => (
												<div key={index} className="bg-gray-700 rounded-lg p-3">
													<p className="text-sm text-gray-400">{formatDate(stat.month, 'monthly')}</p>
													<p className="text-lg font-semibold text-green-400 mt-1">
														{formatAmount(stat.total_amount)}
													</p>
													<p className="text-xs text-gray-500">
														{stat.payment_count} 笔 / {stat.user_count} 人
													</p>
												</div>
											))}
										</div>
									</div>
								)}
							</div>
						)}

						{/* 详细数据表格 */}
						{statType && (
							<div className={styles.card}>
								<h3 className="text-lg font-semibold mb-4">
									{statType === 'daily' ? '每日统计' : 
									 statType === 'monthly' ? '每月统计' : 
									 statType === 'yearly' ? '每年统计' : ''}
								</h3>

								{/* 日统计表 */}
								{statType === 'daily' && statistics.daily_stats && statistics.daily_stats.length > 0 && (
									<div className="overflow-x-auto">
										<table className={styles.table}>
											<thead className={styles.tableHeader}>
												<tr>
													<th className="px-6 py-3">日期</th>
													<th className="px-6 py-3">收入金额</th>
													<th className="px-6 py-3">缴费次数</th>
													<th className="px-6 py-3">缴费用户数</th>
													<th className="px-6 py-3">平均金额</th>
												</tr>
											</thead>
											<tbody>
												{statistics.daily_stats.map((stat) => (
													<tr key={stat.date} className={styles.tableRow}>
														<td className="px-6 py-4">{formatDate(stat.date, 'daily')}</td>
														<td className="px-6 py-4 font-semibold text-green-400">
															{formatAmount(stat.total_amount)}
														</td>
														<td className="px-6 py-4">{stat.payment_count}</td>
														<td className="px-6 py-4">{stat.user_count}</td>
														<td className="px-6 py-4">
															{formatAmount(stat.total_amount / stat.payment_count)}
														</td>
													</tr>
												))}
											</tbody>
										</table>
									</div>
								)}

								{/* 月统计表 */}
								{statType === 'monthly' && statistics.monthly_stats && statistics.monthly_stats.length > 0 && (
									<div className="overflow-x-auto">
										<table className={styles.table}>
											<thead className={styles.tableHeader}>
												<tr>
													<th className="px-6 py-3">月份</th>
													<th className="px-6 py-3">收入金额</th>
													<th className="px-6 py-3">缴费次数</th>
													<th className="px-6 py-3">缴费用户数</th>
													<th className="px-6 py-3">平均金额</th>
												</tr>
											</thead>
											<tbody>
												{statistics.monthly_stats.map((stat) => (
													<tr key={stat.month} className={styles.tableRow}>
														<td className="px-6 py-4">{formatDate(stat.month, 'monthly')}</td>
														<td className="px-6 py-4 font-semibold text-green-400">
															{formatAmount(stat.total_amount)}
														</td>
														<td className="px-6 py-4">{stat.payment_count}</td>
														<td className="px-6 py-4">{stat.user_count}</td>
														<td className="px-6 py-4">
															{formatAmount(stat.total_amount / stat.payment_count)}
														</td>
													</tr>
												))}
											</tbody>
										</table>
									</div>
								)}

								{/* 年统计表 */}
								{statType === 'yearly' && statistics.yearly_stats && statistics.yearly_stats.length > 0 && (
									<div className="overflow-x-auto">
										<table className={styles.table}>
											<thead className={styles.tableHeader}>
												<tr>
													<th className="px-6 py-3">年份</th>
													<th className="px-6 py-3">收入金额</th>
													<th className="px-6 py-3">缴费次数</th>
													<th className="px-6 py-3">缴费用户数</th>
													<th className="px-6 py-3">平均金额</th>
												</tr>
											</thead>
											<tbody>
												{statistics.yearly_stats.map((stat) => (
													<tr key={stat.year} className={styles.tableRow}>
														<td className="px-6 py-4">{formatDate(stat.year, 'yearly')}</td>
														<td className="px-6 py-4 font-semibold text-green-400">
															{formatAmount(stat.total_amount)}
														</td>
														<td className="px-6 py-4">{stat.payment_count}</td>
														<td className="px-6 py-4">{stat.user_count}</td>
														<td className="px-6 py-4">
															{formatAmount(stat.total_amount / stat.payment_count)}
														</td>
													</tr>
												))}
											</tbody>
										</table>
									</div>
								)}

								{/* 无数据提示 */}
								{((statType === 'daily' && (!statistics.daily_stats || statistics.daily_stats.length === 0)) ||
								  (statType === 'monthly' && (!statistics.monthly_stats || statistics.monthly_stats.length === 0)) ||
								  (statType === 'yearly' && (!statistics.yearly_stats || statistics.yearly_stats.length === 0))) && (
									<div className="text-center py-8 text-gray-400">
										<p>暂无数据</p>
									</div>
								)}
							</div>
						)}
					</div>
				)}
			</div>
		</div>
	);
};

export default PaymentStatistics; 