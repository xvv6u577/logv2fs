import React, { useState, useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { alert } from '../store/message';
import axios from 'axios';

const PaymentRecords = () => {
	const [records, setRecords] = useState([]);
	const [loading, setLoading] = useState(false);
	const [deleteLoading, setDeleteLoading] = useState(null);
	const [showAddModal, setShowAddModal] = useState(false);
	const [pagination, setPagination] = useState({
		page: 1,
		limit: 10,
		total: 0,
	});
	const [searchEmail, setSearchEmail] = useState('');

	// 添加表单状态
	const [users, setUsers] = useState([]);
	const [selectedUser, setSelectedUser] = useState('');
	const [amount, setAmount] = useState('');
	const [startDate, setStartDate] = useState(new Date().toISOString().split('T')[0]);
	const [endDate, setEndDate] = useState('');
	const [remark, setRemark] = useState('');
	const [formLoading, setFormLoading] = useState(false);
	const [searchTerm, setSearchTerm] = useState('');

	const dispatch = useDispatch();
	const loginState = useSelector((state) => state.login);

	// 通用样式类
	const styles = {
		container: "min-h-screen bg-gray-900 text-white p-6",
		card: "bg-gray-800 rounded-lg shadow-lg p-6",
		input: "w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-transparent",
		select: "w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:ring-2 focus:ring-blue-500",
		label: "block text-sm font-medium text-gray-300 mb-2",
		button: "px-4 py-2 rounded-lg font-medium text-sm transition-colors focus:outline-none focus:ring-2",
		buttonPrimary: "bg-blue-600 hover:bg-blue-700 text-white focus:ring-blue-500",
		buttonSecondary: "bg-gray-600 hover:bg-gray-700 text-white focus:ring-gray-500",
		buttonDanger: "bg-red-600 hover:bg-red-700 text-white focus:ring-red-500",
		buttonSuccess: "bg-green-600 hover:bg-green-700 text-white focus:ring-green-500",
		modal: "fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50",
		modalContent: "bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto",
		table: "w-full text-sm text-left",
		tableHeader: "text-xs text-gray-400 uppercase bg-gray-700",
		tableRow: "border-b border-gray-700 hover:bg-gray-750",
	};

	// 格式化金额
	const formatAmount = (amount) => {
		return new Intl.NumberFormat('zh-CN', {
			style: 'currency',
			currency: 'CNY',
		}).format(amount);
	};

	// 格式化日期
	const formatDate = (dateStr) => {
		return new Date(dateStr).toLocaleDateString('zh-CN');
	};

	// 获取缴费记录列表
	const fetchRecords = () => {
		setLoading(true);

		const params = {
			page: pagination.page,
			limit: pagination.limit,
		};

		if (searchEmail) {
			params.user_email = searchEmail;
		}

		axios
			.get(process.env.REACT_APP_API_HOST + "payment/records", {
				headers: { token: loginState.token },
				params: params,
			})
			.then((response) => {
				setRecords(response.data.records || []);
				setPagination(prev => ({
					...prev,
					total: response.data.total || 0,
				}));
			})
			.catch((err) => {
				if (err.response) {
					dispatch(alert({ show: true, content: err.response.data.error || "获取缴费记录失败", type: "error" }));
				} else {
					dispatch(alert({ show: true, content: "网络错误: " + err.toString(), type: "error" }));
				}
			})
			.finally(() => {
				setLoading(false);
			});
	};

	// 获取用户列表
	const fetchUsers = () => {
		axios
			.get(process.env.REACT_APP_API_HOST + "n778cf", {
				headers: { token: loginState.token },
			})
			.then((response) => {
				setUsers(response.data);
			})
			.catch((err) => {
				if (err.response) {
					dispatch(alert({ show: true, content: err.response.data.error || "加载用户失败", type: "error" }));
				} else {
					dispatch(alert({ show: true, content: "网络错误: " + err.toString(), type: "error" }));
				}
			});
	};

	// 获取记录ID（兼容MongoDB和PostgreSQL）
	const getRecordId = (record) => {
		return record.id || record._id;
	};

	// 删除缴费记录
	const handleDelete = (recordId, recordInfo) => {
		if (!window.confirm(`确定要删除 ${recordInfo.user_name} 的缴费记录吗？\n金额：${formatAmount(recordInfo.amount)}\n服务期：${formatDate(recordInfo.start_date)} 至 ${formatDate(recordInfo.end_date)}`)) {
			return;
		}

		setDeleteLoading(recordId);

		axios
			.delete(process.env.REACT_APP_API_HOST + `payment/${recordId}`, {
				headers: { token: loginState.token },
			})
			.then((response) => {
				dispatch(alert({ show: true, content: response.data.message || "缴费记录删除成功", type: "success" }));
				fetchRecords(); // 重新加载列表
			})
			.catch((err) => {
				if (err.response) {
					dispatch(alert({ show: true, content: err.response.data.error || "删除失败", type: "error" }));
				} else {
					dispatch(alert({ show: true, content: "删除失败: " + err.toString(), type: "error" }));
				}
			})
			.finally(() => {
				setDeleteLoading(null);
			});
	};

	// 初始加载
	useEffect(() => {
		fetchRecords();
		fetchUsers();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [pagination.page, pagination.limit, searchEmail]);

	// 当开始日期变化时，自动设置结束日期为30天后
	useEffect(() => {
		if (startDate && !endDate) {
			const start = new Date(startDate);
			const end = new Date(start);
			end.setDate(end.getDate() + 30);
			setEndDate(end.toISOString().split('T')[0]);
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [startDate]);

	// 过滤用户列表
	const filteredUsers = users.filter(user => {
		const searchLower = searchTerm.toLowerCase();
		return (
			user.email_as_id?.toLowerCase().includes(searchLower) ||
			user.name?.toLowerCase().includes(searchLower)
		);
	});

	// 计算服务天数
	const calculateDays = () => {
		if (startDate && endDate) {
			const start = new Date(startDate);
			const end = new Date(endDate);
			const diffTime = Math.abs(end - start);
			const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1;
			return diffDays;
		}
		return 0;
	};

	// 提交添加缴费记录
	const handleSubmitAdd = (e) => {
		e.preventDefault();

		// 验证表单
		if (!selectedUser) {
			dispatch(alert({ show: true, content: "请选择用户", type: "error" }));
			return;
		}

		if (!amount || parseFloat(amount) <= 0) {
			dispatch(alert({ show: true, content: "请输入有效的缴费金额", type: "error" }));
			return;
		}

		if (!startDate || !endDate) {
			dispatch(alert({ show: true, content: "请选择服务日期", type: "error" }));
			return;
		}

		if (new Date(endDate) < new Date(startDate)) {
			dispatch(alert({ show: true, content: "服务结束日期不能早于开始日期", type: "error" }));
			return;
		}

		setFormLoading(true);

		// 将日期字符串转换为ISO格式
		const startDateTime = new Date(startDate + 'T00:00:00').toISOString();
		const endDateTime = new Date(endDate + 'T23:59:59').toISOString();

		const paymentData = {
			user_email_as_id: selectedUser,
			amount: parseFloat(amount),
			start_date: startDateTime,
			end_date: endDateTime,
			remark: remark,
		};

		axios
			.post(process.env.REACT_APP_API_HOST + "payment", paymentData, {
				headers: {
					token: loginState.token,
					'Content-Type': 'application/json',
				},
			})
			.then((response) => {
				dispatch(alert({ show: true, content: response.data.message || "缴费记录添加成功", type: "success" }));
				// 重置表单
				setSelectedUser('');
				setAmount('');
				setRemark('');
				setStartDate(new Date().toISOString().split('T')[0]);
				setEndDate('');
				setSearchTerm('');
				setShowAddModal(false);
				fetchRecords(); // 重新加载列表
			})
			.catch((err) => {
				if (err.response) {
					dispatch(alert({ show: true, content: err.response.data.error || "添加失败", type: "error" }));
				} else {
					dispatch(alert({ show: true, content: "添加失败: " + err.toString(), type: "error" }));
				}
			})
			.finally(() => {
				setFormLoading(false);
			});
	};

	// 重置添加表单
	const resetAddForm = () => {
		setSelectedUser('');
		setAmount('');
		setRemark('');
		setStartDate(new Date().toISOString().split('T')[0]);
		setEndDate('');
		setSearchTerm('');
	};

	// 分页处理
	const handlePageChange = (newPage) => {
		setPagination(prev => ({ ...prev, page: newPage }));
	};

	const totalPages = Math.ceil(pagination.total / pagination.limit);

	return (
		<div className={styles.container}>
			<div className="max-w-7xl mx-auto">
				{/* 页面标题和操作栏 */}
				<div className="mb-8 flex flex-col md:flex-row md:items-center md:justify-between">
					<div>
						<h1 className="text-3xl font-bold mb-2">缴费记录管理</h1>
						<p className="text-gray-400">管理用户VPN服务缴费记录</p>
					</div>
					<div className="mt-4 md:mt-0">
						<button
							onClick={() => setShowAddModal(true)}
							className={`${styles.button} ${styles.buttonSuccess}`}
						>
							+ 添加缴费记录
						</button>
					</div>
				</div>

				{/* 搜索和筛选 */}
				<div className={`${styles.card} mb-6`}>
					<div className="grid grid-cols-1 md:grid-cols-3 gap-4">
						<div>
							<label className={styles.label}>按用户邮箱搜索</label>
							<input
								type="text"
								placeholder="输入用户邮箱..."
								value={searchEmail}
								onChange={(e) => setSearchEmail(e.target.value)}
								className={styles.input}
							/>
						</div>
						<div>
							<label className={styles.label}>每页显示</label>
							<select
								value={pagination.limit}
								onChange={(e) => setPagination(prev => ({ ...prev, limit: parseInt(e.target.value), page: 1 }))}
								className={styles.select}
							>
								<option value={10}>10 条</option>
								<option value={20}>20 条</option>
								<option value={50}>50 条</option>
							</select>
						</div>
						<div className="flex items-end">
							<button
								onClick={fetchRecords}
								className={`${styles.button} ${styles.buttonPrimary} w-full`}
								disabled={loading}
							>
								{loading ? '搜索中...' : '搜索'}
							</button>
						</div>
					</div>
				</div>

				{/* 缴费记录表格 */}
				<div className={styles.card}>
					<div className="flex items-center justify-between mb-4">
						<h3 className="text-lg font-semibold">缴费记录列表</h3>
						<div className="text-sm text-gray-400">
							共 {pagination.total} 条记录
						</div>
					</div>

					{loading ? (
						<div className="text-center py-8">
							<div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-white"></div>
							<p className="mt-2 text-gray-400">加载中...</p>
						</div>
					) : (
						<>
							{records.length > 0 ? (
								<div className="overflow-x-auto">
									<table className={styles.table}>
										<thead className={styles.tableHeader}>
											<tr>
												<th className="px-6 py-3">用户</th>
												<th className="px-6 py-3">金额</th>
												<th className="px-6 py-3">服务期</th>
												<th className="px-6 py-3">服务天数</th>
												<th className="px-6 py-3">日均金额</th>
												<th className="px-6 py-3">操作员</th>
												<th className="px-6 py-3">创建时间</th>
												<th className="px-6 py-3">备注</th>
												<th className="px-6 py-3">操作</th>
											</tr>
										</thead>
										<tbody>
											{records.map((record) => (
												<tr key={getRecordId(record)} className={styles.tableRow}>
													<td className="px-6 py-4">
														<div>
															<div className="font-medium">{record.user_name}</div>
															<div className="text-sm text-gray-400">{record.user_email_as_id}</div>
														</div>
													</td>
													<td className="px-6 py-4 font-semibold text-green-400">
														{formatAmount(record.amount)}
													</td>
													<td className="px-6 py-4">
														<div className="text-sm">
															<div>{formatDate(record.start_date)}</div>
															<div className="text-gray-400">至 {formatDate(record.end_date)}</div>
														</div>
													</td>
													<td className="px-6 py-4 text-center">
														{record.service_days} 天
													</td>
													<td className="px-6 py-4">
														{formatAmount(record.daily_amount)}
													</td>
													<td className="px-6 py-4">
														<div className="text-sm">
															<div>{record.operator_name}</div>
															<div className="text-gray-400">{record.operator_email}</div>
														</div>
													</td>
													<td className="px-6 py-4 text-sm text-gray-400">
														{formatDate(record.created_at)}
													</td>
													<td className="px-6 py-4 text-sm">
														{record.remark || '-'}
													</td>
													<td className="px-6 py-4">
														<button
															onClick={() => handleDelete(getRecordId(record), record)}
															disabled={deleteLoading === getRecordId(record)}
															className={`${styles.button} ${styles.buttonDanger} text-xs ${deleteLoading === getRecordId(record) ? 'opacity-50 cursor-not-allowed' : ''}`}
														>
															{deleteLoading === getRecordId(record) ? '删除中...' : '删除'}
														</button>
													</td>
												</tr>
											))}
										</tbody>
									</table>
								</div>
							) : (
								<div className="text-center py-8 text-gray-400">
									<p>暂无缴费记录</p>
								</div>
							)}

							{/* 分页 */}
							{totalPages > 1 && (
								<div className="flex items-center justify-between mt-6">
									<div className="text-sm text-gray-400">
										第 {pagination.page} 页，共 {totalPages} 页
									</div>
									<div className="flex space-x-2">
										<button
											onClick={() => handlePageChange(pagination.page - 1)}
											disabled={pagination.page === 1}
											className={`${styles.button} ${styles.buttonSecondary} ${pagination.page === 1 ? 'opacity-50 cursor-not-allowed' : ''}`}
										>
											上一页
										</button>
										<button
											onClick={() => handlePageChange(pagination.page + 1)}
											disabled={pagination.page === totalPages}
											className={`${styles.button} ${styles.buttonSecondary} ${pagination.page === totalPages ? 'opacity-50 cursor-not-allowed' : ''}`}
										>
											下一页
										</button>
									</div>
								</div>
							)}
						</>
					)}
				</div>

				{/* 添加缴费记录模态框 */}
				{showAddModal && (
					<div className={styles.modal} onClick={(e) => e.target === e.currentTarget && setShowAddModal(false)}>
						<div className={styles.modalContent}>
							<div className="p-6">
								{/* 模态框标题 */}
								<div className="flex items-center justify-between mb-6">
									<h2 className="text-xl font-bold">添加缴费记录</h2>
									<button
										onClick={() => setShowAddModal(false)}
										className="text-gray-400 hover:text-white text-2xl"
									>
										×
									</button>
								</div>

								{/* 添加表单 */}
								<form onSubmit={handleSubmitAdd} className="space-y-6">
									{/* 用户选择 */}
									<div>
										<label className={styles.label}>选择用户</label>
										<input
											type="text"
											placeholder="搜索用户..."
											value={searchTerm}
											onChange={(e) => setSearchTerm(e.target.value)}
											className={`${styles.input} mb-2`}
										/>
										<select
											value={selectedUser}
											onChange={(e) => setSelectedUser(e.target.value)}
											className={styles.select}
											required
										>
											<option value="">请选择用户</option>
											{filteredUsers.map((user) => (
												<option key={user.email_as_id} value={user.email_as_id}>
													{user.name || user.email_as_id} ({user.email_as_id})
												</option>
											))}
										</select>
									</div>

									{/* 缴费金额 */}
									<div>
										<label className={styles.label}>续费金额（元）</label>
										<input
											type="number"
											step="0.01"
											min="0.01"
											placeholder="请输入续费金额"
											value={amount}
											onChange={(e) => setAmount(e.target.value)}
											className={styles.input}
											required
										/>
									</div>

									{/* 服务时间段 */}
									<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
										<div>
											<label className={styles.label}>服务开始日期</label>
											<input
												type="date"
												value={startDate}
												onChange={(e) => setStartDate(e.target.value)}
												className={styles.input}
												required
											/>
										</div>
										<div>
											<label className={styles.label}>服务结束日期</label>
											<input
												type="date"
												value={endDate}
												onChange={(e) => setEndDate(e.target.value)}
												min={startDate}
												className={styles.input}
												required
											/>
										</div>
									</div>

									{/* 服务天数提示 */}
									{startDate && endDate && (
										<div className="bg-blue-900 bg-opacity-30 border border-blue-700 rounded-lg p-3">
											<p className="text-blue-300 text-sm">
												服务期限：<span className="font-semibold">{calculateDays()} 天</span>
												（{startDate} 至 {endDate}）
											</p>
										</div>
									)}

									{/* 备注 */}
									<div>
										<label className={styles.label}>备注（可选）</label>
										<textarea
											placeholder="输入备注信息..."
											value={remark}
											onChange={(e) => setRemark(e.target.value)}
											className={styles.input}
											rows="3"
										/>
									</div>

									{/* 按钮区域 */}
									<div className="flex space-x-4 pt-4">
										<button
											type="submit"
											disabled={formLoading}
											className={`${styles.button} ${styles.buttonSuccess} flex-1 ${formLoading ? 'opacity-50 cursor-not-allowed' : ''}`}
										>
											{formLoading ? '提交中...' : '提交缴费记录'}
										</button>
										<button
											type="button"
											onClick={resetAddForm}
											className={`${styles.button} ${styles.buttonSecondary} flex-1`}
										>
											重置表单
										</button>
									</div>
								</form>
							</div>
						</div>
					</div>
				)}
			</div>
		</div>
	);
};

export default PaymentRecords;
