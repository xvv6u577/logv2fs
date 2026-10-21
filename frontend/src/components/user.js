import { useEffect, useMemo, useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, reset, success } from "../store/message";
import Alert from "./alert";
import AddUser from "./adduser";
import {
	useUsers,
	useUpdateUser,
	useDeleteUser,
	useDisableUser,
	useEnableUser,
} from "../hooks/useQueries";
import { useWebSocket } from "../hooks/useWebSocket";
import ConfirmDialog from "./ui/ConfirmDialog";
import EditUserModal from "./users/EditUserModal";
import UserDrawer from "./users/UserDrawer";
import UserTable from "./users/UserTable";
import { countByStatus, filterUsers, sortUsers } from "./users/userUtils";

const User = () => {
	const { data: users = [], isLoading: loading, error: usersError } = useUsers();
	const [searchTerm, setSearchTerm] = useState("");
	const [sortBy, setSortBy] = useState("role");
	const [sortDir, setSortDir] = useState("desc");
	const [filterStatus, setFilterStatus] = useState("all");
	const [drawerUser, setDrawerUser] = useState(null);
	const [drawerTab, setDrawerTab] = useState("overview");
	const [editUser, setEditUser] = useState(null);
	const [confirm, setConfirm] = useState(null); // { type, user }
	const [selectedIds, setSelectedIds] = useState(() => new Set());

	const { status: wsStatus } = useWebSocket();
	const updateUserMutation = useUpdateUser();
	const deleteUserMutation = useDeleteUser();
	const disableUserMutation = useDisableUser();
	const enableUserMutation = useEnableUser();

	const dispatch = useDispatch();
	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);

	const counts = useMemo(() => countByStatus(users), [users]);

	const filteredAndSortedUsers = useMemo(() => {
		const filtered = filterUsers(users, { searchTerm, filterStatus });
		return sortUsers(filtered, sortBy, sortDir);
	}, [users, searchTerm, sortBy, sortDir, filterStatus]);

	const copyToClipboard = (text) => {
		navigator.clipboard.writeText(text).then(
			() => dispatch(success({ show: true, content: "已复制到剪贴板" })),
			() => dispatch(alert({ show: true, content: "复制失败" }))
		);
	};

	const buildMutationCallbacks = (successText, errorText) => ({
		onSuccess: (data) => {
			dispatch(success({ show: true, content: data.message || successText }));
		},
		onError: (err) => {
			dispatch(
				alert({
					show: true,
					content: err.response?.data?.error || `${errorText}: ${err.toString()}`,
				})
			);
		},
	});

	const resetFilters = () => {
		setSearchTerm("");
		setFilterStatus("all");
		setSortBy("role");
		setSortDir("desc");
	};

	const handleKeyDown = (event) => {
		if (event.key === "Escape") resetFilters();
	};

	const handleSort = (col) => {
		if (sortBy === col) {
			setSortDir((d) => (d === "asc" ? "desc" : "asc"));
		} else {
			setSortBy(col);
			setSortDir(col === "name" || col === "status" ? "asc" : "desc");
		}
	};

	const toggleSelect = (user) => {
		if (user.role === "admin") return;
		setSelectedIds((prev) => {
			const next = new Set(prev);
			if (next.has(user.email_as_id)) next.delete(user.email_as_id);
			else next.add(user.email_as_id);
			return next;
		});
	};

	const toggleSelectAll = () => {
		const selectable = filteredAndSortedUsers.filter((u) => u.role !== "admin");
		const allOn = selectable.length > 0 && selectable.every((u) => selectedIds.has(u.email_as_id));
		if (allOn) {
			setSelectedIds(new Set());
		} else {
			setSelectedIds(new Set(selectable.map((u) => u.email_as_id)));
		}
	};

	const runBatch = async (action) => {
		const targets = filteredAndSortedUsers.filter(
			(u) => selectedIds.has(u.email_as_id) && u.role !== "admin"
		);
		if (targets.length === 0) {
			dispatch(alert({ show: true, content: "请先勾选普通用户" }));
			return;
		}
		let ok = 0;
		let fail = 0;
		for (const u of targets) {
			try {
				if (action === "disable") {
					await disableUserMutation.mutateAsync({ email_as_id: u.email_as_id });
				} else {
					await enableUserMutation.mutateAsync({ email_as_id: u.email_as_id });
				}
				ok += 1;
			} catch {
				fail += 1;
			}
		}
		setSelectedIds(new Set());
		dispatch(
			success({
				show: true,
				content: `批量${action === "disable" ? "禁用" : "启用"}完成：成功 ${ok}，失败 ${fail}`,
			})
		);
	};

	const submitEdit = (editData) => {
		updateUserMutation.mutate(editData, {
			...buildMutationCallbacks("用户编辑成功", "编辑失败"),
			onSuccess: (data) => {
				dispatch(success({ show: true, content: data.message || "用户编辑成功" }));
				setEditUser(null);
			},
		});
	};

	const confirmRequireText = confirm?.user
		? confirm.user.name || confirm.user.email_as_id
		: "";

	useEffect(() => {
		if (message.show === true) {
			const t = setTimeout(() => dispatch(reset({})), 5000);
			return () => clearTimeout(t);
		}
	}, [message, dispatch]);

	useEffect(() => {
		if (usersError) {
			dispatch(alert({ show: true, content: usersError.toString() }));
		}
	}, [usersError, dispatch]);

	const tabs = [
		{ key: "all", label: "全部", count: counts.all },
		{ key: "plain", label: "在用", count: counts.plain },
		{ key: "disabled", label: "禁用", count: counts.disabled },
	];

	return (
		<div className="min-h-screen bg-gray-900 text-white p-4 md:p-6" onKeyDown={handleKeyDown} tabIndex={0}>
			<Alert
				message={message.content}
				type={message.type}
				shown={message.show}
				close={() => dispatch(reset({}))}
			/>

			<UserDrawer
				user={drawerUser}
				open={Boolean(drawerUser)}
				onClose={() => setDrawerUser(null)}
				onCopy={copyToClipboard}
				tab={drawerTab}
				onTabChange={setDrawerTab}
			/>

			<EditUserModal
				open={Boolean(editUser)}
				user={editUser}
				onClose={() => setEditUser(null)}
				onSubmit={submitEdit}
				submitting={updateUserMutation.isPending}
			/>

			<ConfirmDialog
				open={Boolean(confirm)}
				title={
					confirm?.type === "delete"
						? "删除用户"
						: confirm?.type === "disable"
							? "禁用用户"
							: "启用用户"
				}
				message={
					confirm?.type === "delete"
						? `确定删除用户「${confirm?.user?.name || confirm?.user?.email_as_id}」？此操作不可撤销。`
						: confirm?.type === "disable"
							? `确定禁用「${confirm?.user?.name || confirm?.user?.email_as_id}」？禁用后将无法使用服务。`
							: `确定启用「${confirm?.user?.name || confirm?.user?.email_as_id}」？`
				}
				danger={confirm?.type === "delete" || confirm?.type === "disable"}
				requireText={confirm?.type === "delete" ? confirmRequireText : ""}
				requireHint={
					confirm?.type === "delete"
						? `请输入用户名或邮箱 ID（${confirmRequireText}）以确认删除`
						: ""
				}
				confirmLabel={
					confirm?.type === "delete" ? "删除" : confirm?.type === "disable" ? "禁用" : "启用"
				}
				onCancel={() => setConfirm(null)}
				onConfirm={() => {
					const u = confirm?.user;
					const type = confirm?.type;
					setConfirm(null);
					if (!u) return;
					if (type === "delete") {
						deleteUserMutation.mutate(
							{ email_as_id: u.email_as_id },
							buildMutationCallbacks("用户删除成功", "删除失败")
						);
					} else if (type === "disable") {
						disableUserMutation.mutate(
							{ email_as_id: u.email_as_id },
							buildMutationCallbacks("用户已禁用", "禁用失败")
						);
					} else if (type === "enable") {
						enableUserMutation.mutate(
							{ email_as_id: u.email_as_id },
							buildMutationCallbacks("用户已启用", "启用失败")
						);
					}
				}}
			/>

			{/* 标题 */}
			<div className="mb-6 flex flex-col md:flex-row md:items-center md:justify-between gap-4">
				<div>
					<h1 className="text-2xl md:text-3xl font-bold mb-1">用户管理</h1>
					<p className="text-gray-400 text-sm">管理和监控用户状态</p>
					<div className="flex items-center space-x-2 mt-2">
						<div
							className={`w-2 h-2 rounded-full ${
								wsStatus === "connected"
									? "bg-green-500"
									: wsStatus === "connecting"
										? "bg-yellow-500"
										: wsStatus === "reconnecting"
											? "bg-orange-500"
											: "bg-red-500"
							}`}
						/>
						<span className="text-xs text-gray-400">
							{wsStatus === "connected"
								? "实时连接已建立"
								: wsStatus === "connecting"
									? "正在连接…"
									: wsStatus === "reconnecting"
										? "正在重连…"
										: "连接已断开"}
						</span>
					</div>
				</div>
				{loginState.jwt.role === "admin" && <AddUser btnName="添加用户" />}
			</div>

			{/* 状态页签 */}
			<div className="flex flex-wrap gap-2 mb-4">
				{tabs.map((t) => (
					<button
						key={t.key}
						type="button"
						onClick={() => setFilterStatus(t.key)}
						className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
							filterStatus === t.key
								? "bg-blue-600 text-white"
								: "bg-gray-800 text-gray-300 hover:bg-gray-700"
						}`}
					>
						{t.label}
						<span className="ml-1.5 text-xs opacity-80">{t.count}</span>
					</button>
				))}
			</div>

			{/* 搜索 + 批量 */}
			<div className="mb-4 flex flex-col md:flex-row gap-3">
				<div className="relative flex-1">
					<input
						type="text"
						placeholder="搜索用户名、邮箱 ID、角色、备注…"
						value={searchTerm}
						onChange={(e) => setSearchTerm(e.target.value)}
						className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:ring-2 focus:ring-blue-500"
					/>
				</div>
				<button
					type="button"
					onClick={resetFilters}
					className="px-4 py-2 rounded-lg bg-gray-600 hover:bg-gray-700 text-sm"
					title="重置过滤条件（Esc）"
				>
					重置
				</button>
				{selectedIds.size > 0 && (
					<>
						<button
							type="button"
							onClick={() => runBatch("enable")}
							className="px-4 py-2 rounded-lg bg-green-700 hover:bg-green-600 text-sm"
						>
							批量启用 ({selectedIds.size})
						</button>
						<button
							type="button"
							onClick={() => runBatch("disable")}
							className="px-4 py-2 rounded-lg bg-orange-700 hover:bg-orange-600 text-sm"
						>
							批量禁用 ({selectedIds.size})
						</button>
					</>
				)}
			</div>

			{loading ? (
				<div className="bg-gray-800 rounded-lg p-8 text-center text-gray-400">加载中…</div>
			) : filteredAndSortedUsers.length === 0 ? (
				<div className="bg-gray-800 rounded-lg p-8 text-center text-gray-400">未找到用户</div>
			) : (
				<UserTable
					users={filteredAndSortedUsers}
					sortBy={sortBy}
					sortDir={sortDir}
					onSort={handleSort}
					selectedIds={selectedIds}
					onToggle={toggleSelect}
					onToggleAll={toggleSelectAll}
					onCopy={copyToClipboard}
					onOpenTab={(u, t) => {
						setDrawerTab(t);
						setDrawerUser(u);
					}}
					onEdit={setEditUser}
					onEnable={(u) => setConfirm({ type: "enable", user: u })}
					onDisable={(u) => setConfirm({ type: "disable", user: u })}
					onDelete={(u) => setConfirm({ type: "delete", user: u })}
				/>
			)}
		</div>
	);
};

export default User;
