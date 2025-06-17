import { useEffect, useState, useMemo } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, reset } from "../store/message";
import axios from "axios";
import Alert from "./alert";

const Home = () => {
	const [users, setUsers] = useState([]);
	const [loading, setLoading] = useState(true);
	const [searchTerm, setSearchTerm] = useState("");
	const [sortBy, setSortBy] = useState("role");
	const [filterStatus, setFilterStatus] = useState("all");
	const [modalUser, setModalUser] = useState(null);
	const [editModalOpen, setEditModalOpen] = useState(false);
	const [editingUser, setEditingUser] = useState(null);
	const [editForm, setEditForm] = useState({ name: "", role: "" });

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
		buttonCopy: "px-2 py-1 bg-gray-600 hover:bg-gray-500 text-white text-xs rounded transition-colors",
		input: "w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-transparent",
		select: "px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:ring-2 focus:ring-blue-500",
		card: "bg-gray-800 rounded-lg shadow-lg hover:shadow-xl transition-shadow",
		badge: "px-2 py-1 rounded-full text-xs font-medium",
		badgeAdmin: "bg-purple-900 text-purple-300",
		badgeUser: "bg-blue-900 text-blue-300",
		badgeOnline: "bg-green-900 text-green-300",
		badgeOffline: "bg-red-900 text-red-300",
	};

	// 排序函数
	const sortUsers = (users, sortType) => {
		const getTraffic = (user, period) => {
			switch (period) {
				case 'daily': return user.daily_logs?.[0]?.traffic ?? 0;
				case 'monthly': return user.monthly_logs?.[0]?.traffic ?? 0;
				default: return 0;
			}
		};

		switch (sortType) {
			case "role":
				return users.sort((a, b) => {
					if (a.role === "admin" && b.role !== "admin") return -1;
					if (b.role === "admin" && a.role !== "admin") return 1;
					return getTraffic(b, 'monthly') - getTraffic(a, 'monthly');
				});
			case "online":
				return users.sort((a, b) => {
					if (a.status === "plain" && b.status !== "plain") return -1;
					if (b.status === "plain" && a.status !== "plain") return 1;
					return getTraffic(b, 'monthly') - getTraffic(a, 'monthly');
				});
			case "today":
				return users.sort((a, b) => {
					if (a.status === "plain" && b.status !== "plain") return -1;
					if (b.status === "plain" && a.status !== "plain") return 1;
					return getTraffic(b, 'daily') - getTraffic(a, 'daily');
				});
			case "monthly":
				return users.sort((a, b) => {
					if (a.status === "plain" && b.status !== "plain") return -1;
					if (b.status === "plain" && a.status !== "plain") return 1;
					return getTraffic(b, 'monthly') - getTraffic(a, 'monthly');
				});
			case "used":
				return users.sort((a, b) => {
					if (a.status === "plain" && b.status !== "plain") return -1;
					if (b.status === "plain" && a.status !== "plain") return 1;
					return (b.used || 0) - (a.used || 0);
				});
			case "name":
				return users.sort((a, b) => (a.name || "").localeCompare(b.name || ""));
			case "update_time":
				return users.sort((a, b) => new Date(b.updated_at || 0) - new Date(a.updated_at || 0));
			default:
				return users;
		}
	};

	// 过滤和搜索用户
	const filteredAndSortedUsers = useMemo(() => {
		let filtered = users;

		// 状态过滤
		if (filterStatus !== "all") {
			filtered = filtered.filter(user => user.status === filterStatus);
		}

		// 搜索过滤
		if (searchTerm) {
			filtered = filtered.filter(user =>
				user.email_as_id?.toLowerCase().includes(searchTerm.toLowerCase()) ||
				user.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
				user.role?.toLowerCase().includes(searchTerm.toLowerCase()) ||
				user.remark?.toLowerCase().includes(searchTerm.toLowerCase())
			);
		}

		// 排序
		return sortUsers([...filtered], sortBy);
	}, [users, searchTerm, sortBy, filterStatus]);

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

	// 打开编辑用户模态框
	const openEditModal = (user) => {
		setEditingUser(user);
		setEditForm({
			name: user.name || "",
			role: user.role || "normal"
		});
		setEditModalOpen(true);
	};

	// 关闭编辑模态框
	const closeEditModal = () => {
		setEditModalOpen(false);
		setEditingUser(null);
		setEditForm({ name: "", role: "" });
	};

	// 提交编辑用户
	const submitEditUser = () => {
		if (!editingUser) return;

		// 验证表单数据
		if (!editForm.name.trim()) {
			dispatch(alert({ show: true, content: "用户名不能为空", type: "error" }));
			return;
		}

		// 创建编辑用户的数据
		const editData = {
			email_as_id: editingUser.email_as_id,
			name: editForm.name.trim(),
			role: editForm.role
		};

		// 调用编辑用户API
		axios
			.post(`${process.env.REACT_APP_API_HOST}edit/${editingUser.email_as_id}`, editData, {
				headers: { 
					token: loginState.token,
					'Content-Type': 'application/json'
				},
			})
			.then((response) => {
				dispatch(alert({ show: true, content: response.data.message || "用户编辑成功", type: "success" }));
				closeEditModal();
				// 重新获取用户列表
				window.location.reload();
			})
			.catch((err) => {
				if (err.response) {
					dispatch(alert({ show: true, content: err.response.data.error || "编辑失败", type: "error" }));
				} else {
					dispatch(alert({ show: true, content: "编辑失败: " + err.toString(), type: "error" }));
				}
			});
	};

	// 删除用户
	const deleteUser = (user) => {
		if (window.confirm(`确定要删除用户 "${user.name || user.email_as_id}" 吗？此操作不可撤销。`)) {
			// 调用删除用户API
			axios
				.get(`${process.env.REACT_APP_API_HOST}deluser/${user.email_as_id}`, {
					headers: { 
						token: loginState.token,
						'Content-Type': 'application/json'
					},
				})
				.then((response) => {
					dispatch(alert({ show: true, content: response.data.message || "用户删除成功", type: "success" }));
					// 重新获取用户列表
					setUsers(users.filter(u => u.email_as_id !== user.email_as_id));
				})
				.catch((err) => {
					if (err.response) {
						dispatch(alert({ show: true, content: err.response.data.error || "删除失败", type: "error" }));
					} else {
						dispatch(alert({ show: true, content: "删除失败: " + err.toString(), type: "error" }));
					}
				});
		}
	};

	// 打开用户详情模态框
	const openUserModal = (user) => {
		setModalUser(user);
	};

	// 关闭用户详情模态框
	const closeUserModal = () => {
		setModalUser(null);
	};

	// 重置所有过滤条件
	const resetFilters = () => {
		setSearchTerm("");
		setFilterStatus("all");
		setSortBy("role");
	};

	// 处理键盘事件
	const handleKeyDown = (event) => {
		if (event.key === 'Enter') {
			resetFilters();
		}
	};

	// 用户卡片组件
	const UserCard = ({ user, index }) => {
		return (
			<div className={`${styles.card} overflow-hidden`}>
				{/* 用户基本信息 */}
				<div className="p-4">
					{/* 用户头像和基本信息 */}
					<div className="flex items-center space-x-3 mb-3">
						<div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center text-white font-bold text-sm">
							{index + 1}
						</div>
						<div className="flex-1 min-w-0">
							<h3 className="text-sm font-semibold text-white truncate">
								{user.name || user.email_as_id || "未知用户"}
							</h3>
							<p className="text-gray-400 text-xs truncate">
								{user.email_as_id || "无邮箱"}
							</p>
						</div>
					</div>

					{/* 状态标签 */}
					<div className="flex flex-wrap gap-1 mb-3">
						<span className={`${styles.badge} ${user.role === "admin" ? styles.badgeAdmin : styles.badgeUser}`}>
							{user.role === "admin" ? "管理员" : "用户"}
						</span>
						<span className={`${styles.badge} ${user.status === "plain" ? styles.badgeOnline : styles.badgeOffline}`}>
							{user.status === "plain" ? "在线" : "离线"}
						</span>
					</div>

					{/* 流量统计 */}
					<div className="space-y-3 mb-3">
						<div className="flex justify-between items-center">
							<span className="text-sm font-bold text-blue-200">今日</span>
							<span className="text-sm text-blue-400 font-bold">
								{user.daily_logs?.[0]?.traffic ? formatBytes(user.daily_logs[0].traffic) : "0 B"}
							</span>
						</div>
						<div className="flex justify-between items-center">
							<span className="text-sm font-bold text-green-200">本月</span>
							<span className="text-sm text-green-400 font-bold">
								{user.monthly_logs?.[0]?.traffic ? formatBytes(user.monthly_logs[0].traffic) : "0 B"}
							</span>
						</div>
					</div>

					{/* 更新时间 */}
					<div className="mb-3">
						<p className="text-gray-500 text-xs">
							更新于: {formatDate(user.updated_at)}
						</p>
					</div>

					{/* 备注信息 */}
					{user.remark && (
						<div className="mb-3 pt-2 border-t border-gray-700">
							<p className="text-gray-400 text-xs truncate" title={user.remark}>
								备注: {user.remark}
							</p>
						</div>
					)}

					{/* 操作按钮区域 */}
					<div className="flex space-x-2">
						<button
							onClick={() => openUserModal(user)}
							className={`${styles.button} ${styles.buttonPrimary} flex-1 text-xs`}
						>
							详情
						</button>
						{user.email_as_id && (
							<button
								onClick={() => copyToClipboard(user.email_as_id)}
								className={`${styles.button} ${styles.buttonSecondary} flex-1 text-xs`}
								title="复制邮箱"
							>
								复制
							</button>
						)}
					</div>
				</div>
			</div>
		);
	};

	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(reset({}));
			}, 5000);
		}
	}, [message, dispatch]);

	useEffect(() => {
		setLoading(true);
		axios
			.get(process.env.REACT_APP_API_HOST + "n778cf", {
				headers: { token: loginState.token },
			})
			.then((response) => {
				setUsers(response.data);
				setLoading(false);
			})
			.catch((err) => {
				setLoading(false);
				if (err.response) {
					dispatch(alert({ show: true, content: err.response.data.error || "加载用户失败", type: "error" }));
				} else {
					dispatch(alert({ show: true, content: "网络错误: " + err.toString(), type: "error" }));
				}
			});
	}, [rerenderSignal, loginState.jwt.Email, loginState.token, dispatch]);

	return (
		<div className="min-h-screen bg-gray-900 text-white p-6" onKeyDown={handleKeyDown} tabIndex={0}>
			<Alert 
				message={message.content} 
				type={message.type} 
				shown={message.show} 
				close={() => { dispatch(reset({})); }} 
			/>

						{/* 用户详情模态框 */}
			{modalUser && (
				<div 
					className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
					onClick={closeUserModal}
				>
					<div 
						className="bg-gray-800 rounded-lg w-full max-w-4xl max-h-[90vh] overflow-y-auto"
						onClick={(e) => e.stopPropagation()}
					>
						{/* 模态框头部 */}
						<div className="flex items-center justify-between p-6 border-b border-gray-700">
							<h3 className="text-xl font-bold text-white">用户详情</h3>
							<button
								onClick={closeUserModal}
								className="text-gray-400 hover:text-white transition-colors"
							>
								<svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
								</svg>
							</button>
						</div>

						{/* 模态框内容 */}
						<div className="p-6 space-y-6">
							{/* 用户基本信息 */}
							<div>
								<h4 className="text-lg font-medium text-gray-300 mb-4">基本信息</h4>
								<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
									<div>
										<span className="text-gray-400">用户名: </span>
										<span className="text-white">{modalUser.name || "未设置"}</span>
									</div>
									<div className="flex items-center space-x-2">
										<span className="text-gray-400">邮箱ID: </span>
										<span className="text-white">{modalUser.email_as_id || "无"}</span>
										{modalUser.email_as_id && (
											<button
												onClick={() => copyToClipboard(modalUser.email_as_id)}
												className={styles.buttonCopy}
												title="复制邮箱ID"
											>
												复制
											</button>
										)}
									</div>
									<div>
										<span className="text-gray-400">状态: </span>
										<span className={modalUser.status === "plain" ? "text-green-400" : "text-red-400"}>
											{modalUser.status === "plain" ? "活跃" : "非活跃"}
										</span>
									</div>
									<div>
										<span className="text-gray-400">角色: </span>
										<span className={modalUser.role === "admin" ? "text-purple-400" : "text-blue-400"}>
											{modalUser.role === "admin" ? "管理员" : "普通用户"}
										</span>
									</div>
									<div>
										<span className="text-gray-400">总使用: </span>
										<span className="text-white">{modalUser.used ? formatBytes(modalUser.used) : "0 B"}</span>
									</div>
									<div>
										<span className="text-gray-400">最后更新: </span>
										<span className="text-white">{formatDate(modalUser.updated_at)}</span>
									</div>
									{modalUser.remark && (
										<div className="md:col-span-2">
											<span className="text-gray-400">备注: </span>
											<span className="text-white">{modalUser.remark}</span>
										</div>
									)}
								</div>
							</div>

							{/* 订阅链接 */}
							<div>
								<h4 className="text-lg font-medium text-gray-300 mb-4">订阅链接</h4>
								<div className="space-y-3">
									{/* Shadowrocket/Surge 订阅 */}
									<div className="flex items-center justify-between p-3 bg-gray-700 rounded-lg">
										<div>
											<span className="text-white font-medium">Shadowrocket/Surge 订阅</span>
											<p className="text-gray-400 text-sm mt-1">适用于 iOS Shadowrocket 和 Surge</p>
										</div>
										<button
											onClick={() => copyToClipboard(process.env.REACT_APP_FILE_AND_SUB_URL + "/static/" + modalUser.email_as_id)}
											className={`${styles.button} ${styles.buttonPrimary} text-xs`}
											title="复制订阅链接"
										>
											复制链接
										</button>
									</div>

									{/* Verge 订阅 */}
									<div className="flex items-center justify-between p-3 bg-gray-700 rounded-lg">
										<div>
											<span className="text-white font-medium">Verge 订阅</span>
											<p className="text-gray-400 text-sm mt-1">适用于 Verge 客户端</p>
										</div>
										<button
											onClick={() => copyToClipboard(process.env.REACT_APP_FILE_AND_SUB_URL + "/verge/" + modalUser.email_as_id)}
											className={`${styles.button} ${styles.buttonPrimary} text-xs`}
											title="复制订阅链接"
										>
											复制链接
										</button>
									</div>

									{/* Sing-box 订阅 */}
									<div className="flex items-center justify-between p-3 bg-gray-700 rounded-lg">
										<div>
											<span className="text-white font-medium">Sing-box 订阅</span>
											<p className="text-gray-400 text-sm mt-1">适用于 Sing-box 客户端</p>
										</div>
										<button
											onClick={() => copyToClipboard(process.env.REACT_APP_FILE_AND_SUB_URL + "/singbox/" + modalUser.email_as_id)}
											className={`${styles.button} ${styles.buttonPrimary} text-xs`}
											title="复制订阅链接"
										>
											复制链接
										</button>
									</div>
								</div>
							</div>

							{/* 流量统计 */}
							<div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
								{/* 月度流量 */}
								<div>
									<h4 className="text-lg font-medium text-gray-300 mb-4">月度流量统计</h4>
									<div className="bg-gray-700 rounded-lg overflow-hidden">
										{modalUser.monthly_logs && modalUser.monthly_logs.length > 0 ? (
											<table className="w-full text-sm">
												<thead className="bg-gray-600">
													<tr>
														<th className="px-4 py-3 text-left">月份</th>
														<th className="px-4 py-3 text-right">流量</th>
													</tr>
												</thead>
												<tbody>
													{modalUser.monthly_logs
														.sort((a, b) => b.month - a.month)
														.slice(0, 10)
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

								{/* 日度流量 */}
								<div>
									<h4 className="text-lg font-medium text-gray-300 mb-4">近期日流量统计</h4>
									<div className="bg-gray-700 rounded-lg overflow-hidden">
										{modalUser.daily_logs && modalUser.daily_logs.length > 0 ? (
											<table className="w-full text-sm">
												<thead className="bg-gray-600">
													<tr>
														<th className="px-4 py-3 text-left">日期</th>
														<th className="px-4 py-3 text-right">流量</th>
													</tr>
												</thead>
												<tbody>
													{modalUser.daily_logs
														.sort((a, b) => b.date - a.date)
														.slice(0, 10)
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

							{/* 操作按钮区域 */}
							<div className="flex space-x-4 pt-4 border-t border-gray-700">
								<button 
									className={`${styles.button} ${styles.buttonPrimary}`}
									onClick={() => {
										closeUserModal();
										openEditModal(modalUser);
									}}
								>
									编辑用户
								</button>
								{modalUser.role !== "admin" && (
									<button 
										className={`${styles.button} ${styles.buttonDanger}`}
										onClick={() => {
											closeUserModal();
											deleteUser(modalUser);
										}}
									>
										删除用户
									</button>
								)}
								<button 
									className={`${styles.button} ${styles.buttonSecondary}`}
									onClick={closeUserModal}
								>
									关闭
								</button>
							</div>
						</div>
					</div>
				</div>
			)}

			{/* 编辑用户模态框 */}
			{editModalOpen && (
				<div 
					className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
					onClick={closeEditModal}
				>
					<div 
						className="bg-gray-800 rounded-lg p-6 w-full max-w-md"
						onClick={(e) => e.stopPropagation()}
					>
						<h3 className="text-xl font-bold mb-4">编辑用户</h3>
						
						<div className="space-y-4">
							{/* 邮箱ID（只读） */}
							<div>
								<label className="block text-sm font-medium text-gray-300 mb-2">
									邮箱ID
								</label>
								<input
									type="text"
									value={editingUser?.email_as_id || ""}
									disabled
									className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-gray-400 cursor-not-allowed"
								/>
							</div>

							{/* 用户名 */}
							<div>
								<label className="block text-sm font-medium text-gray-300 mb-2">
									用户名
								</label>
								<input
									type="text"
									value={editForm.name}
									onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
									className={styles.input}
									placeholder="请输入用户名"
								/>
							</div>

							{/* 角色 */}
							<div>
								<label className="block text-sm font-medium text-gray-300 mb-2">
									角色
								</label>
								<select
									value={editForm.role}
									onChange={(e) => setEditForm({ ...editForm, role: e.target.value })}
									className={styles.select}
								>
									<option value="normal">普通用户</option>
									<option value="admin">管理员</option>
								</select>
							</div>
						</div>

						{/* 按钮区域 */}
						<div className="flex space-x-4 mt-6">
				<button
								onClick={submitEditUser}
								className={`${styles.button} ${styles.buttonPrimary} flex-1`}
							>
								保存
				</button>
				<button
								onClick={closeEditModal}
								className={`${styles.button} ${styles.buttonSecondary} flex-1`}
							>
								取消
				</button>
						</div>
					</div>
				</div>
			)}

			{/* 页面标题 */}
			<div className="mb-8">
				<h1 className="text-3xl font-bold mb-2">用户管理</h1>
				<p className="text-gray-400">管理和监控用户状态</p>
			</div>

			{/* 搜索和过滤控制栏 */}
			<div className="mb-6 grid grid-cols-1 md:grid-cols-4 gap-4">
				{/* 搜索框 */}
				<div className="relative">
					<input
						type="text"
						placeholder="搜索用户（用户名、邮箱ID、角色、备注）..."
						value={searchTerm}
						onChange={(e) => setSearchTerm(e.target.value)}
						className={styles.input}
					/>
					<div className="absolute inset-y-0 right-0 pr-3 flex items-center">
						<svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
						</svg>
					</div>
				</div>

				{/* 排序选择 */}
				<select
					value={sortBy}
					onChange={(e) => setSortBy(e.target.value)}
					className={styles.select}
				>
					<option value="role">按角色排序</option>
					<option value="name">按用户名排序</option>
					<option value="update_time">按更新时间排序</option>
					<option value="online">按在线状态</option>
					<option value="today">按今日流量</option>
					<option value="monthly">按月流量</option>
					<option value="used">按总使用量</option>
				</select>

				{/* 状态过滤 */}
				<select
					value={filterStatus}
					onChange={(e) => setFilterStatus(e.target.value)}
					className={styles.select}
				>
					<option value="all">所有状态</option>
					<option value="plain">在线</option>
					<option value="offline">离线</option>
				</select>

				{/* 清空过滤按钮 */}
				<button
					onClick={resetFilters}
					className={`${styles.button} ${styles.buttonSecondary} flex items-center justify-center space-x-2`}
					title="重置所有过滤条件 (按Enter键)"
				>
					<svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
					</svg>
					<span>重置</span>
				</button>
			</div>

			{/* 用户列表 */}
			{loading ? (
				<div className={`${styles.card} p-8 text-center`}>
					<div className="flex flex-col items-center">
						<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mb-4"></div>
						<h3 className="text-lg font-medium text-gray-300 mb-2">加载中...</h3>
						<p className="text-gray-400">正在获取用户数据</p>
					</div>
				</div>
			) : filteredAndSortedUsers.length === 0 ? (
				<div className={`${styles.card} p-8 text-center`}>
					<svg className="mx-auto h-12 w-12 text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2M4 13h2m13-8V4a1 1 0 00-1-1H7a1 1 0 00-1 1v1m13 0H5" />
					</svg>
					<h3 className="text-lg font-medium text-gray-300 mb-2">未找到用户</h3>
					<p className="text-gray-400">尝试调整搜索条件或过滤器</p>
				</div>
			) : (
				<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
					{filteredAndSortedUsers.map((user, index) => (
						<UserCard key={user.id || index} user={user} index={index} />
					))}
				</div>
			)}
		</div>
	);
};

export default Home;
