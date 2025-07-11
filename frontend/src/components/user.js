import { useEffect, useState, useMemo } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, reset } from "../store/message";
import axios from "axios";
import Alert from "./alert";
import AddUser from "./adduser";

const User = () => {
	const [users, setUsers] = useState([]);
	const [loading, setLoading] = useState(true);
	const [searchTerm, setSearchTerm] = useState("");
	const [sortBy, setSortBy] = useState("role");
	const [filterStatus, setFilterStatus] = useState("all");
	const [modalUser, setModalUser] = useState(null);
	const [editModalOpen, setEditModalOpen] = useState(false);
	const [editingUser, setEditingUser] = useState(null);
	const [editForm, setEditForm] = useState({ 
		name: "", 
		role: "", 
		password: "", 
		confirmPassword: "",
		remark: ""
	});
	const [userPayments, setUserPayments] = useState([]);
	const [paymentLoading, setPaymentLoading] = useState(false);

	// 添加缴费记录相关状态
	const [paymentModalOpen, setPaymentModalOpen] = useState(false);
	const [paymentForm, setPaymentForm] = useState({
		selectedUser: "",
		amount: "",
		startDate: new Date().toISOString().split('T')[0],
		endDate: "",
		remark: ""
	});
	const [paymentFormLoading, setPaymentFormLoading] = useState(false);

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
		buttonWarning: "bg-orange-600 hover:bg-orange-700 text-white focus:ring-orange-500",
		buttonSuccess: "bg-green-600 hover:bg-green-700 text-white focus:ring-green-500",
		buttonCopy: "px-2 py-1 bg-gray-600 hover:bg-gray-500 text-white text-xs rounded transition-colors",
		input: "w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-transparent",
		select: "px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:ring-2 focus:ring-blue-500",
		card: "bg-gray-800 rounded-lg shadow-lg hover:shadow-xl transition-shadow",
		badge: "px-2 py-1 rounded-full text-xs font-medium",
		badgeAdmin: "bg-purple-900 text-purple-300",
		badgeUser: "bg-blue-900 text-blue-300",
		badgeOnline: "bg-green-900 text-green-300",
		badgeOffline: "bg-red-900 text-red-300",
		badgeDisabled: "bg-gray-900 text-gray-300",
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
			if (filterStatus === "offline") {
				// 离线状态包括所有非"plain"和非"deleted"的状态
				filtered = filtered.filter(user => user.status !== "plain" && user.status !== "deleted");
			} else {
				filtered = filtered.filter(user => user.status === filterStatus);
			}
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
			role: user.role || "normal",
			password: "",
			confirmPassword: "",
			remark: user.remark || ""
		});
		setEditModalOpen(true);
	};

	// 关闭编辑模态框
	const closeEditModal = () => {
		setEditModalOpen(false);
		setEditingUser(null);
		setEditForm({ 
			name: "", 
			role: "", 
			password: "", 
			confirmPassword: "",
			remark: ""
		});
	};

	// 提交编辑用户
	const submitEditUser = () => {
		if (!editingUser) return;

		// 验证表单数据
		if (!editForm.name.trim()) {
			dispatch(alert({ show: true, content: "用户名不能为空", type: "error" }));
			return;
		}

		// 验证密码（如果输入了密码）
		if (editForm.password) {
			if (editForm.password.length < 6) {
				dispatch(alert({ show: true, content: "密码至少需要6个字符", type: "error" }));
				return;
			}
			if (editForm.password !== editForm.confirmPassword) {
				dispatch(alert({ show: true, content: "两次输入的密码不一致", type: "error" }));
				return;
			}
		}

		// 创建编辑用户的数据
		const editData = {
			email_as_id: editingUser.email_as_id,
			name: editForm.name.trim(),
			role: editForm.role,
			remark: editForm.remark.trim()
		};

		// 只有在输入了密码时才添加密码字段
		if (editForm.password) {
			editData.password = editForm.password;
		}

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

	// 禁用用户
	const disableUser = (user) => {
		if (window.confirm(`确定要禁用用户 "${user.name || user.email_as_id}" 吗？禁用后用户将无法使用服务。`)) {
			axios
				.put(`${process.env.REACT_APP_API_HOST}disableuser/${user.email_as_id}`, {}, {
					headers: { 
						token: loginState.token,
						'Content-Type': 'application/json'
					},
				})
				.then((response) => {
					dispatch(alert({ show: true, content: response.data.message || "用户已禁用", type: "success" }));
					// 更新用户列表中的状态
					setUsers(users.map(u => 
						u.email_as_id === user.email_as_id 
							? { ...u, status: "deleted" }
							: u
					));
				})
				.catch((err) => {
					if (err.response) {
						dispatch(alert({ show: true, content: err.response.data.error || "禁用失败", type: "error" }));
					} else {
						dispatch(alert({ show: true, content: "禁用失败: " + err.toString(), type: "error" }));
					}
				});
		}
	};

	// 启用用户
	const enableUser = (user) => {
		if (window.confirm(`确定要启用用户 "${user.name || user.email_as_id}" 吗？`)) {
			axios
				.put(`${process.env.REACT_APP_API_HOST}enableuser/${user.email_as_id}`, {}, {
					headers: { 
						token: loginState.token,
						'Content-Type': 'application/json'
					},
				})
				.then((response) => {
					dispatch(alert({ show: true, content: response.data.message || "用户已启用", type: "success" }));
					// 更新用户列表中的状态
					setUsers(users.map(u => 
						u.email_as_id === user.email_as_id 
							? { ...u, status: "plain" }
							: u
					));
				})
				.catch((err) => {
					if (err.response) {
						dispatch(alert({ show: true, content: err.response.data.error || "启用失败", type: "error" }));
					} else {
						dispatch(alert({ show: true, content: "启用失败: " + err.toString(), type: "error" }));
					}
				});
		}
	};

	// 获取用户缴费记录
	const fetchUserPayments = (userEmail) => {
		setPaymentLoading(true);
		axios
			.get(`${process.env.REACT_APP_API_HOST}payment/user/${userEmail}`, {
				headers: { token: loginState.token },
			})
			.then((response) => {
				setUserPayments(response.data.payments || []);
			})
			.catch((err) => {
				setUserPayments([]);
				if (err.response && err.response.status !== 404) {
					dispatch(alert({ show: true, content: err.response.data.error || "获取缴费记录失败", type: "error" }));
				}
			})
			.finally(() => {
				setPaymentLoading(false);
			});
	};

	// 格式化金额
	const formatCurrency = (amount) => {
		return new Intl.NumberFormat('zh-CN', {
			style: 'currency',
			currency: 'CNY',
		}).format(amount);
	};

	// 打开用户详情模态框
	const openUserModal = (user) => {
		setModalUser(user);
		setUserPayments([]);
		fetchUserPayments(user.email_as_id);
	};

	// 关闭用户详情模态框
	const closeUserModal = () => {
		setModalUser(null);
		setUserPayments([]);
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

	// 打开添加缴费记录模态框
	const openPaymentModal = (user) => {
		setPaymentForm({
			selectedUser: user.email_as_id,
			amount: "",
			startDate: new Date().toISOString().split('T')[0],
			endDate: "",
			remark: ""
		});
		setPaymentModalOpen(true);
	};

	// 关闭添加缴费记录模态框
	const closePaymentModal = () => {
		setPaymentModalOpen(false);
		setPaymentForm({
			selectedUser: "",
			amount: "",
			startDate: new Date().toISOString().split('T')[0],
			endDate: "",
			remark: ""
		});
	};

	// 计算服务天数
	const calculateDays = () => {
		if (paymentForm.startDate && paymentForm.endDate) {
			const start = new Date(paymentForm.startDate);
			const end = new Date(paymentForm.endDate);
			const diffTime = Math.abs(end - start);
			const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) + 1;
			return diffDays;
		}
		return 0;
	};

	// 提交添加缴费记录
	const handleSubmitPayment = (e) => {
		e.preventDefault();

		// 验证表单
		if (!paymentForm.selectedUser) {
			dispatch(alert({ show: true, content: "用户未选择", type: "error" }));
			return;
		}

		if (!paymentForm.amount || parseFloat(paymentForm.amount) <= 0) {
			dispatch(alert({ show: true, content: "请输入有效的缴费金额", type: "error" }));
			return;
		}

		if (!paymentForm.startDate || !paymentForm.endDate) {
			dispatch(alert({ show: true, content: "请选择服务日期", type: "error" }));
			return;
		}

		if (new Date(paymentForm.endDate) < new Date(paymentForm.startDate)) {
			dispatch(alert({ show: true, content: "服务结束日期不能早于开始日期", type: "error" }));
			return;
		}

		setPaymentFormLoading(true);

		// 将日期字符串转换为ISO格式
		const startDateTime = new Date(paymentForm.startDate + 'T00:00:00').toISOString();
		const endDateTime = new Date(paymentForm.endDate + 'T23:59:59').toISOString();

		const paymentData = {
			user_email_as_id: paymentForm.selectedUser,
			amount: parseFloat(paymentForm.amount),
			start_date: startDateTime,
			end_date: endDateTime,
			remark: paymentForm.remark,
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
				closePaymentModal();
				// 如果当前正在查看该用户的详情，刷新缴费记录
				if (modalUser && modalUser.email_as_id === paymentForm.selectedUser) {
					fetchUserPayments(modalUser.email_as_id);
				}
			})
			.catch((err) => {
				if (err.response) {
					dispatch(alert({ show: true, content: err.response.data.error || "添加失败", type: "error" }));
				} else {
					dispatch(alert({ show: true, content: "添加失败: " + err.toString(), type: "error" }));
				}
			})
			.finally(() => {
				setPaymentFormLoading(false);
			});
	};

	// 当开始日期变化时，自动设置结束日期为30天后
	useEffect(() => {
		if (paymentForm.startDate && !paymentForm.endDate) {
			const start = new Date(paymentForm.startDate);
			const end = new Date(start);
			end.setDate(end.getDate() + 30);
			setPaymentForm(prev => ({
				...prev,
				endDate: end.toISOString().split('T')[0]
			}));
		}
	}, [paymentForm.startDate, paymentForm.endDate]);

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
							<div className="flex items-center space-x-2">
								<p className="text-gray-400 text-xs truncate flex-1">
									{user.email_as_id || "无邮箱"}
								</p>
								{user.email_as_id && (
									<button
										onClick={() => copyToClipboard(user.email_as_id)}
										className="px-1 py-0.5 bg-gray-600 hover:bg-gray-500 text-white text-xs rounded transition-colors flex-shrink-0"
										title="复制邮箱ID"
									>
										<svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
										</svg>
									</button>
								)}
							</div>
						</div>
					</div>

					{/* 状态标签 */}
					<div className="flex flex-wrap gap-1 mb-3">
						<span className={`${styles.badge} ${user.role === "admin" ? styles.badgeAdmin : styles.badgeUser}`}>
							{user.role === "admin" ? "管理员" : "用户"}
						</span>
						<span className={`${styles.badge} ${
							user.status === "plain" ? styles.badgeOnline : 
							user.status === "deleted" ? styles.badgeDisabled : 
							styles.badgeOffline
						}`}>
							{user.status === "plain" ? "在线" : 
							 user.status === "deleted" ? "已禁用" : 
							 "离线"}
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
						<button
							onClick={() => openPaymentModal(user)}
							className={`${styles.button} ${styles.buttonSuccess} flex-1 text-xs`}
							title="添加缴费记录"
						>
							缴费
						</button>
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
										<span className={
											modalUser.status === "plain" ? "text-green-400" : 
											modalUser.status === "deleted" ? "text-gray-400" : 
											"text-red-400"
										}>
											{modalUser.status === "plain" ? "活跃" : 
											 modalUser.status === "deleted" ? "已禁用" : 
											 "非活跃"}
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

							{/* 缴费记录部分 */}
							<div>
								<h4 className="text-lg font-medium text-gray-300 mb-4">缴费记录</h4>
								<div className="bg-gray-700 rounded-lg overflow-hidden">
									{paymentLoading ? (
										<div className="p-4 text-center text-gray-400">
											<div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-2"></div>
											<p>加载缴费记录中...</p>
										</div>
									) : userPayments.length > 0 ? (
										<div>
											{/* 缴费汇总 */}
											<div className="p-4 bg-gray-600">
												<div className="grid grid-cols-3 gap-4 text-center">
													<div>
														<p className="text-gray-400 text-sm">总缴费金额</p>
														<p className="text-xl font-bold text-green-400">
															{formatCurrency(userPayments.reduce((sum, p) => sum + p.amount, 0))}
														</p>
													</div>
													<div>
														<p className="text-gray-400 text-sm">缴费次数</p>
														<p className="text-xl font-bold text-blue-400">{userPayments.length}</p>
													</div>
													<div>
														<p className="text-gray-400 text-sm">最近缴费</p>
														<p className="text-sm text-white">
															{userPayments[0] ? new Date(userPayments[0].start_date).toLocaleDateString('zh-CN') : '-'}
														</p>
													</div>
												</div>
											</div>
											{/* 缴费记录列表 */}
											<div className="max-h-64 overflow-y-auto">
												<table className="w-full text-sm">
													<thead className="bg-gray-600 sticky top-0">
														<tr>
															<th className="px-4 py-2 text-left">服务期间</th>
															<th className="px-4 py-2 text-right">金额</th>
															<th className="px-4 py-2 text-left">备注</th>
														</tr>
													</thead>
													<tbody>
														{userPayments.map((payment, idx) => {
															const startDate = new Date(payment.start_date);
															const endDate = new Date(payment.end_date);
															return (
																<tr key={payment._id || payment.id || idx} className="border-t border-gray-600">
																	<td className="px-4 py-2">
																		{startDate.toLocaleDateString('zh-CN')} ~ {endDate.toLocaleDateString('zh-CN')}
																	</td>
																	<td className="px-4 py-2 text-right font-mono text-green-400">
																		{formatCurrency(payment.amount)}
																	</td>
																	<td className="px-4 py-2 text-gray-400">
																		{payment.remark || '-'}
																	</td>
																</tr>
															);
														})}
													</tbody>
												</table>
											</div>
										</div>
									) : (
										<div className="p-4 text-center text-gray-400">
											暂无缴费记录
										</div>
									)}
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
							<div className="flex flex-wrap gap-3 pt-4 border-t border-gray-700">
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
									<>
										{modalUser.status === "deleted" ? (
											<button 
												className={`${styles.button} ${styles.buttonSuccess}`}
												onClick={() => {
													closeUserModal();
													enableUser(modalUser);
												}}
											>
												启用用户
											</button>
										) : (
											<button 
												className={`${styles.button} ${styles.buttonWarning}`}
												onClick={() => {
													closeUserModal();
													disableUser(modalUser);
												}}
											>
												禁用用户
											</button>
										)}
										<button 
											className={`${styles.button} ${styles.buttonDanger}`}
											onClick={() => {
												closeUserModal();
												deleteUser(modalUser);
											}}
										>
											删除用户
										</button>
									</>
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
						<h3 className="text-xl font-bold mb-4">编辑用户信息</h3>
						<p className="text-sm text-gray-400 mb-4">修改用户名、角色或密码</p>
						
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

							{/* 备注 */}
							<div>
								<label className="block text-sm font-medium text-gray-300 mb-2">
									用户备注
								</label>
								<textarea
									value={editForm.remark}
									onChange={(e) => setEditForm({ ...editForm, remark: e.target.value })}
									className={styles.input}
									placeholder="输入用户备注信息（可选）"
									rows="3"
								/>
								<p className="text-xs text-gray-500 mt-1">用于记录用户的特殊说明或备注信息</p>
							</div>

							{/* 密码编辑区域 */}
							<div className="border-t pt-4 border-gray-600">
								<h4 className="text-lg font-medium text-gray-300 mb-3">修改密码（可选）</h4>
								<p className="text-sm text-gray-400 mb-4">留空则保持当前密码不变</p>
								
								{/* 新密码 */}
								<div className="mb-4">
									<label className="block text-sm font-medium text-gray-300 mb-2">
										新密码（至少6个字符）
									</label>
									<input
										type="password"
										value={editForm.password}
										onChange={(e) => setEditForm({ ...editForm, password: e.target.value })}
										className={styles.input}
										placeholder="输入新密码"
									/>
								</div>

								{/* 确认密码 */}
								<div>
									<label className="block text-sm font-medium text-gray-300 mb-2">
										确认新密码
									</label>
									<input
										type="password"
										value={editForm.confirmPassword}
										onChange={(e) => setEditForm({ ...editForm, confirmPassword: e.target.value })}
										className={styles.input}
										placeholder="再次输入新密码"
									/>
									{editForm.password && editForm.confirmPassword && editForm.password !== editForm.confirmPassword && (
										<p className="mt-1 text-sm text-red-400">密码不一致</p>
									)}
									{editForm.password && editForm.password.length > 0 && editForm.password.length < 6 && (
										<p className="mt-1 text-sm text-red-400">密码至少需要6个字符</p>
									)}
								</div>
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

			{/* 添加缴费记录模态框 */}
			{paymentModalOpen && (
				<div 
					className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
					onClick={closePaymentModal}
				>
					<div 
						className="bg-gray-800 rounded-lg w-full max-w-md max-h-[90vh] overflow-y-auto"
						onClick={(e) => e.stopPropagation()}
					>
						{/* 模态框头部 */}
						<div className="flex items-center justify-between p-6 border-b border-gray-700">
							<h3 className="text-xl font-bold text-white">添加缴费记录</h3>
							<button
								onClick={closePaymentModal}
								className="text-gray-400 hover:text-white transition-colors"
							>
								<svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
								</svg>
							</button>
						</div>

						{/* 模态框内容 */}
						<div className="p-6">
							<form onSubmit={handleSubmitPayment} className="space-y-6">
								{/* 用户信息（只读） */}
								<div>
									<label className="block text-sm font-medium text-gray-300 mb-2">
										用户
									</label>
									<input
										type="text"
										value={users.find(u => u.email_as_id === paymentForm.selectedUser)?.name || paymentForm.selectedUser}
										disabled
										className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-gray-400 cursor-not-allowed"
									/>
								</div>

								{/* 缴费金额 */}
								<div>
									<label className="block text-sm font-medium text-gray-300 mb-2">
										续费金额（元）
									</label>
									<input
										type="number"
										step="0.01"
										min="0.01"
										placeholder="请输入续费金额"
										value={paymentForm.amount}
										onChange={(e) => setPaymentForm({...paymentForm, amount: e.target.value})}
										className={styles.input}
										required
									/>
								</div>

								{/* 服务时间段 */}
								<div className="grid grid-cols-1 gap-4">
									<div>
										<label className="block text-sm font-medium text-gray-300 mb-2">
											服务开始日期
										</label>
										<input
											type="date"
											value={paymentForm.startDate}
											onChange={(e) => setPaymentForm({...paymentForm, startDate: e.target.value})}
											className={styles.input}
											required
										/>
									</div>
									<div>
										<label className="block text-sm font-medium text-gray-300 mb-2">
											服务结束日期
										</label>
										<input
											type="date"
											value={paymentForm.endDate}
											onChange={(e) => setPaymentForm({...paymentForm, endDate: e.target.value})}
											min={paymentForm.startDate}
											className={styles.input}
											required
										/>
									</div>
								</div>

								{/* 服务天数提示 */}
								{paymentForm.startDate && paymentForm.endDate && (
									<div className="bg-blue-900 bg-opacity-30 border border-blue-700 rounded-lg p-3">
										<p className="text-blue-300 text-sm">
											服务期限：<span className="font-semibold">{calculateDays()} 天</span>
											（{paymentForm.startDate} 至 {paymentForm.endDate}）
										</p>
									</div>
								)}

								{/* 备注 */}
								<div>
									<label className="block text-sm font-medium text-gray-300 mb-2">
										备注（可选）
									</label>
									<textarea
										placeholder="输入备注信息..."
										value={paymentForm.remark}
										onChange={(e) => setPaymentForm({...paymentForm, remark: e.target.value})}
										className={styles.input}
										rows="3"
									/>
								</div>

								{/* 按钮区域 */}
								<div className="flex space-x-4 pt-4">
									<button
										type="submit"
										disabled={paymentFormLoading}
										className={`${styles.button} ${styles.buttonSuccess} flex-1 ${paymentFormLoading ? 'opacity-50 cursor-not-allowed' : ''}`}
									>
										{paymentFormLoading ? '提交中...' : '提交缴费记录'}
									</button>
									<button
										type="button"
										onClick={closePaymentModal}
										className={`${styles.button} ${styles.buttonSecondary} flex-1`}
									>
										取消
									</button>
								</div>
							</form>
						</div>
					</div>
				</div>
			)}

			{/* 页面标题 */}
			<div className="mb-8 flex flex-col md:flex-row md:items-center md:justify-between">
				<div>
					<h1 className="text-3xl font-bold mb-2">用户管理</h1>
					<p className="text-gray-400">管理和监控用户状态</p>
				</div>
				{loginState.jwt.Role === "admin" && (
					<div className="mt-4 md:mt-0">
						<AddUser btnName="添加用户" />
					</div>
				)}
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
					<option value="deleted">已禁用</option>
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

export default User;
