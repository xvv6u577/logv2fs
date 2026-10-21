import React from "react";
import DropdownMenu from "../ui/DropdownMenu";
import {
	formatBytes,
	getCurrentMonthTraffic,
	getTrafficOfTodayFromArray,
} from "../../service/service";
import { getStatusBadgeClass, getStatusLabel, subUrl } from "./userUtils";

function SortHeader({ label, col, sortBy, sortDir, onSort }) {
	const active = sortBy === col;
	return (
		<th className="px-3 py-2 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
			<button
				type="button"
				onClick={() => onSort(col)}
				className={`inline-flex items-center gap-1 hover:text-white ${active ? "text-white" : ""}`}
			>
				{label}
				{active && <span className="text-blue-400">{sortDir === "asc" ? "↑" : "↓"}</span>}
			</button>
		</th>
	);
}

function SubDropdown({ user, onCopy }) {
	const items = [];
	[
		["static", "Shadowrocket"],
		["verge", "Verge"],
		["singbox", "Sing-box"],
	].forEach(([type, name]) => {
		items.push({
			key: `${type}-full`,
			label: `${name} · 完整`,
			onClick: () => onCopy(subUrl(type, user)),
		});
		items.push({
			key: `${type}-hk`,
			label: `${name} · 去 HK`,
			onClick: () => onCopy(subUrl(type, user, true)),
		});
	});
	return (
		<DropdownMenu
			label="复制订阅"
			align="right"
			buttonClassName="px-2.5 py-1 rounded-lg text-xs font-medium bg-blue-600 hover:bg-blue-700 text-white"
			items={items}
		/>
	);
}

const ghostBtn =
	"px-2 py-1 rounded-md text-xs text-gray-300 border border-gray-600 hover:bg-gray-700 hover:text-white whitespace-nowrap";

function RowActions({ user, onCopy, onOpenTab, onEdit, onEnable, onDisable, onDelete }) {
	return (
		<div className="flex flex-wrap gap-1.5 justify-end items-center">
			<SubDropdown user={user} onCopy={onCopy} />
			<button type="button" className={ghostBtn} onClick={() => onOpenTab(user, "overview")}>概览</button>
			<button type="button" className={ghostBtn} onClick={() => onOpenTab(user, "subs")}>订阅</button>
			<button type="button" className={ghostBtn} onClick={() => onOpenTab(user, "traffic")}>流量</button>
			<button type="button" className={ghostBtn} onClick={() => onEdit(user)}>编辑</button>
			<MoreMenu user={user} onEnable={onEnable} onDisable={onDisable} onDelete={onDelete} />
		</div>
	);
}

function MoreMenu({ user, onEnable, onDisable, onDelete }) {
	if (user.role === "admin") return null;
	const inactive = user.status === "disabled" || user.status === "overdue";
	const items = [
		inactive
			? { key: "enable", label: "启用", onClick: () => onEnable(user) }
			: { key: "disable", label: "禁用", onClick: () => onDisable(user) },
		{ key: "delete", label: "删除", danger: true, onClick: () => onDelete(user) },
	];
	return (
		<DropdownMenu
			label="⋯"
			align="right"
			buttonClassName="px-2 py-1 rounded-md text-xs text-gray-300 border border-gray-600 hover:bg-gray-700"
			items={items}
		/>
	);
}

function UserCardMobile({ user, selected, onToggle, ...actions }) {
	return (
		<div className={`bg-gray-800 rounded-lg p-3 border ${selected ? "border-blue-500" : "border-gray-700"}`}>
			<div className="flex items-start gap-2 mb-2">
				{user.role !== "admin" && (
					<input type="checkbox" checked={selected} onChange={() => onToggle(user)} className="mt-1" />
				)}
				<div className="flex-1 min-w-0">
					<div className="flex items-center gap-2 flex-wrap">
						<span className="text-white font-medium text-sm truncate">{user.name || user.email_as_id}</span>
						<span className={`px-2 py-0.5 rounded-full text-xs font-medium ${getStatusBadgeClass(user.status)}`}>
							{getStatusLabel(user.status)}
						</span>
						{user.role === "admin" && (
							<span className="px-2 py-0.5 rounded-full text-xs font-medium bg-purple-900 text-purple-300">管理员</span>
						)}
					</div>
					{user.remark && <p className="text-gray-500 text-xs truncate mt-0.5">{user.remark}</p>}
				</div>
			</div>
			<div className="grid grid-cols-3 gap-2 text-center text-xs mb-3">
				<div>
					<p className="text-gray-500">今日</p>
					<p className="text-blue-400 font-mono">{getTrafficOfTodayFromArray(user.daily_logs)}</p>
				</div>
				<div>
					<p className="text-gray-500">本月</p>
					<p className="text-green-400 font-mono">{getCurrentMonthTraffic(user.monthly_logs)}</p>
				</div>
				<div>
					<p className="text-gray-500">总量</p>
					<p className="text-white font-mono">{user.used ? formatBytes(user.used) : "0 B"}</p>
				</div>
			</div>
			<RowActions user={user} {...actions} />
		</div>
	);
}

export default function UserTable({
	users,
	sortBy,
	sortDir,
	onSort,
	selectedIds,
	onToggle,
	onToggleAll,
	onCopy,
	onOpenTab,
	onEdit,
	onEnable,
	onDisable,
	onDelete,
}) {
	const selectable = users.filter((u) => u.role !== "admin");
	const allSelected = selectable.length > 0 && selectable.every((u) => selectedIds.has(u.email_as_id));
	const actions = { onCopy, onOpenTab, onEdit, onEnable, onDisable, onDelete };

	return (
		<>
			{/* 桌面表格 */}
			<div className="hidden md:block overflow-x-auto bg-gray-800 rounded-lg border border-gray-700">
				<table className="min-w-full divide-y divide-gray-700">
					<thead className="bg-gray-750 bg-gray-900">
						<tr>
							<th className="px-3 py-2 w-10">
								<input
									type="checkbox"
									checked={allSelected}
									onChange={onToggleAll}
									disabled={selectable.length === 0}
									title="全选（跳过管理员）"
								/>
							</th>
							<SortHeader label="用户" col="name" sortBy={sortBy} sortDir={sortDir} onSort={onSort} />
							<SortHeader label="状态" col="status" sortBy={sortBy} sortDir={sortDir} onSort={onSort} />
							<SortHeader label="今日" col="today" sortBy={sortBy} sortDir={sortDir} onSort={onSort} />
							<SortHeader label="本月" col="monthly" sortBy={sortBy} sortDir={sortDir} onSort={onSort} />
							<SortHeader label="总量" col="used" sortBy={sortBy} sortDir={sortDir} onSort={onSort} />
							<th className="px-3 py-2 text-right text-xs font-medium text-gray-400 uppercase w-[22rem]">操作</th>
						</tr>
					</thead>
					<tbody className="divide-y divide-gray-700">
						{users.map((user) => (
							<tr key={user.email_as_id || user.id} className="hover:bg-gray-750 hover:bg-opacity-50">
								<td className="px-3 py-2">
									{user.role !== "admin" ? (
										<input
											type="checkbox"
											checked={selectedIds.has(user.email_as_id)}
											onChange={() => onToggle(user)}
										/>
									) : (
										<span className="text-gray-600 text-xs">—</span>
									)}
								</td>
								<td className="px-3 py-2">
									<div className="text-sm text-white font-medium">{user.name || user.email_as_id}</div>
									{user.remark && <div className="text-xs text-gray-500 truncate max-w-xs">{user.remark}</div>}
									{user.role === "admin" && (
										<span className="inline-block mt-0.5 px-1.5 py-0.5 rounded text-[10px] bg-purple-900 text-purple-300">管理员</span>
									)}
								</td>
								<td className="px-3 py-2">
									<span className={`px-2 py-0.5 rounded-full text-xs font-medium ${getStatusBadgeClass(user.status)}`}>
										{getStatusLabel(user.status)}
									</span>
								</td>
								<td className="px-3 py-2 text-sm font-mono text-blue-400">{getTrafficOfTodayFromArray(user.daily_logs)}</td>
								<td className="px-3 py-2 text-sm font-mono text-green-400">{getCurrentMonthTraffic(user.monthly_logs)}</td>
								<td className="px-3 py-2 text-sm font-mono text-gray-300">{user.used ? formatBytes(user.used) : "0 B"}</td>
								<td className="px-3 py-2 w-[22rem]">
									<RowActions user={user} {...actions} />
								</td>
							</tr>
						))}
					</tbody>
				</table>
			</div>

			{/* 手机紧凑卡片 */}
			<div className="md:hidden space-y-3">
				{users.map((user) => (
					<UserCardMobile
						key={user.email_as_id || user.id}
						user={user}
						selected={selectedIds.has(user.email_as_id)}
						onToggle={onToggle}
						{...actions}
					/>
				))}
			</div>
		</>
	);
}
