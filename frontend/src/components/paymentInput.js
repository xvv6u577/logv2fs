import React, { useState, useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { alert } from '../store/message';
import axios from 'axios';

const PaymentInput = () => {
	const [users, setUsers] = useState([]);
	const [selectedUser, setSelectedUser] = useState('');
	const [amount, setAmount] = useState('');
	const [startDate, setStartDate] = useState(new Date().toISOString().split('T')[0]);
	const [endDate, setEndDate] = useState('');
	const [remark, setRemark] = useState('');
	const [loading, setLoading] = useState(false);
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
		errorText: "text-red-400 text-sm mt-1",
	};

	// 获取用户列表
	useEffect(() => {
		fetchUsers();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, []);

	// 当开始日期变化时，自动设置结束日期为30天后
	useEffect(() => {
		if (startDate && !endDate) {
			const start = new Date(startDate);
			const end = new Date(start);
			end.setDate(end.getDate() + 30); // 默认30天服务期
			setEndDate(end.toISOString().split('T')[0]);
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [startDate]);

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
			const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1; // +1 包含结束日期
			return diffDays;
		}
		return 0;
	};

	// 提交缴费记录
	const handleSubmit = (e) => {
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

		if (!startDate) {
			dispatch(alert({ show: true, content: "请选择服务开始日期", type: "error" }));
			return;
		}

		if (!endDate) {
			dispatch(alert({ show: true, content: "请选择服务结束日期", type: "error" }));
			return;
		}

		// 验证结束日期不能早于开始日期
		if (new Date(endDate) < new Date(startDate)) {
			dispatch(alert({ show: true, content: "服务结束日期不能早于开始日期", type: "error" }));
			return;
		}

		setLoading(true);

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
			})
			.catch((err) => {
				if (err.response) {
					dispatch(alert({ show: true, content: err.response.data.error || "添加失败", type: "error" }));
				} else {
					dispatch(alert({ show: true, content: "添加失败: " + err.toString(), type: "error" }));
				}
			})
			.finally(() => {
				setLoading(false);
			});
	};

	// 重置表单
	const handleReset = () => {
		setSelectedUser('');
		setAmount('');
		setRemark('');
		setStartDate(new Date().toISOString().split('T')[0]);
		setEndDate('');
		setSearchTerm('');
	};

	return (
		<div className={styles.container}>
			<div className="max-w-4xl mx-auto">
				{/* 页面标题 */}
				<div className="mb-8">
					<h1 className="text-3xl font-bold mb-2">添加缴费记录</h1>
					<p className="text-gray-400">为用户录入VPN服务续费信息</p>
				</div>

				{/* 表单卡片 */}
				<div className={styles.card}>
					<form onSubmit={handleSubmit} className="space-y-6">
						{/* 用户选择 */}
						<div>
							<label className={styles.label}>选择用户</label>
							{/* 搜索框 */}
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
							{/* 开始日期 */}
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

							{/* 结束日期 */}
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
								disabled={loading}
								className={`${styles.button} ${styles.buttonPrimary} flex-1 ${loading ? 'opacity-50 cursor-not-allowed' : ''}`}
							>
								{loading ? '提交中...' : '提交缴费记录'}
							</button>
							<button
								type="button"
								onClick={handleReset}
								className={`${styles.button} ${styles.buttonSecondary} flex-1`}
							>
								重置表单
							</button>
						</div>
					</form>
				</div>

				{/* 提示信息 */}
				<div className="mt-6 bg-blue-900 bg-opacity-30 border border-blue-700 rounded-lg p-4">
					<h3 className="text-blue-300 font-semibold mb-2">提示</h3>
					<ul className="text-blue-200 text-sm space-y-1 list-disc list-inside">
						<li>请确保用户信息正确，缴费记录一旦提交将记录在系统中</li>
						<li>续费金额必须大于0</li>
						<li>服务开始日期通常为今天或用户当前服务到期的第二天</li>
						<li>服务结束日期决定了VPN服务的有效期限</li>
						<li>系统会自动计算服务天数，方便核对</li>
						<li>备注信息可用于记录特殊说明或优惠信息</li>
					</ul>
				</div>
			</div>
		</div>
	);
};

export default PaymentInput; 