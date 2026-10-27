import React from "react";
import * as Tabs from "@radix-ui/react-tabs";
import Drawer from "../ui/Drawer";
import { formatBytes, formatDate, getCurrentMonthTraffic, getTrafficOfTodayFromArray } from "../../service/service";
import { getStatusBadgeClass, getStatusLabel, subUrl } from "./userUtils";

function SubCopyRow({ title, desc, type, user, onCopy }) {
	return (
		<div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 p-3 bg-gray-700 rounded-lg">
			<div>
				<span className="text-white font-medium text-sm">{title}</span>
				<p className="text-gray-400 text-xs mt-0.5">{desc}</p>
			</div>
			<div className="flex gap-2 flex-shrink-0">
				<button type="button" onClick={() => onCopy(subUrl(type, user))} className="px-3 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-700 text-white text-xs">
					复制链接
				</button>
				<button type="button" onClick={() => onCopy(subUrl(type, user, true))} className="px-3 py-1.5 rounded-lg bg-gray-600 hover:bg-gray-500 text-white text-xs" title="?exclude=hk">
					去 HK
				</button>
			</div>
		</div>
	);
}

function TrafficTable({ rows, labelKey, valueKey, valueClass }) {
	if (!rows?.length) {
		return <p className="text-gray-400 text-sm p-4 text-center">暂无数据</p>;
	}
	return (
		<table className="w-full text-sm">
			<thead className="bg-gray-600 sticky top-0">
				<tr>
					<th className="px-3 py-2 text-left">周期</th>
					<th className="px-3 py-2 text-right">流量</th>
				</tr>
			</thead>
			<tbody>
				{rows.map((row, i) => (
					<tr key={i} className="border-t border-gray-600">
						<td className="px-3 py-2">{row[labelKey]}</td>
						<td className={`px-3 py-2 text-right font-mono ${valueClass}`}>{formatBytes(row[valueKey])}</td>
					</tr>
				))}
			</tbody>
		</table>
	);
}

export default function UserDrawer({ user, open, onClose, onCopy, tab = "overview", onTabChange }) {
	if (!user) return null;

	const monthly = [...(user.monthly_logs || [])].sort((a, b) => b.month - a.month).slice(0, 10);
	const daily = [...(user.daily_logs || [])].sort((a, b) => b.date - a.date).slice(0, 10);

	return (
		<Drawer open={open} onClose={onClose} title={user.name || user.email_as_id || "用户详情"} widthClass="max-w-xl">
			<Tabs.Root value={tab} onValueChange={onTabChange}>
				<Tabs.List className="flex gap-1 mb-4 border-b border-gray-700 pb-2">
					{[
						["overview", "概览"],
						["subs", "订阅"],
						["traffic", "流量"],
					].map(([v, label]) => (
						<Tabs.Trigger
							key={v}
							value={v}
							className="px-3 py-1.5 text-sm rounded-t text-gray-400 data-[state=active]:text-white data-[state=active]:bg-gray-700"
						>
							{label}
						</Tabs.Trigger>
					))}
				</Tabs.List>

				<Tabs.Content value="overview" className="space-y-3 text-sm">
					<div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
						<div>
							<span className="text-gray-400">用户名：</span>
							<span className="text-white">{user.name || "—"}</span>
						</div>
						<div className="sm:col-span-2">
							<span className="text-gray-400">邮箱 ID：</span>
							<span className="text-white break-all">{user.email_as_id || "—"}</span>
							{Array.isArray(user.email_as_id_aliases) && user.email_as_id_aliases.length > 0 && (
								<div className="mt-1 text-xs text-gray-400">
									旧 ID（仍可用）：
									<span className="text-gray-300 break-all">{user.email_as_id_aliases.join(", ")}</span>
								</div>
							)}
						</div>
						<div>
							<span className="text-gray-400">状态：</span>
							<span className={`px-2 py-0.5 rounded-full text-xs font-medium ${getStatusBadgeClass(user.status)}`}>
								{getStatusLabel(user.status)}
							</span>
						</div>
						<div>
							<span className="text-gray-400">角色：</span>
							<span className={user.role === "admin" ? "text-purple-400" : "text-blue-400"}>
								{user.role === "admin" ? "管理员" : "普通用户"}
							</span>
						</div>
						<div>
							<span className="text-gray-400">今日：</span>
							<span className="text-blue-400 font-mono">{getTrafficOfTodayFromArray(user.daily_logs)}</span>
						</div>
						<div>
							<span className="text-gray-400">本月：</span>
							<span className="text-green-400 font-mono">{getCurrentMonthTraffic(user.monthly_logs)}</span>
						</div>
						<div>
							<span className="text-gray-400">总量：</span>
							<span className="text-white font-mono">{user.used ? formatBytes(user.used) : "0 B"}</span>
						</div>
						<div>
							<span className="text-gray-400">更新：</span>
							<span className="text-white">{formatDate(user.updated_at)}</span>
						</div>
					</div>
					{user.remark && (
						<div className="pt-2 border-t border-gray-700">
							<span className="text-gray-400">备注：</span>
							<span className="text-white">{user.remark}</span>
						</div>
					)}
				</Tabs.Content>

				<Tabs.Content value="subs" className="space-y-3">
					<SubCopyRow title="Shadowrocket / Surge" desc="iOS Shadowrocket、Surge" type="static" user={user} onCopy={onCopy} />
					<SubCopyRow title="Verge" desc="Verge 客户端" type="verge" user={user} onCopy={onCopy} />
					<SubCopyRow title="Sing-box" desc="Sing-box 客户端" type="singbox" user={user} onCopy={onCopy} />
				</Tabs.Content>

				<Tabs.Content value="traffic" className="space-y-4">
					<div>
						<h4 className="text-sm font-medium text-gray-300 mb-2">月度（最近 10）</h4>
						<div className="bg-gray-700 rounded-lg overflow-hidden max-h-64 overflow-y-auto">
							<TrafficTable rows={monthly} labelKey="month" valueKey="traffic" valueClass="text-green-400" />
						</div>
					</div>
					<div>
						<h4 className="text-sm font-medium text-gray-300 mb-2">日流量（最近 10）</h4>
						<div className="bg-gray-700 rounded-lg overflow-hidden max-h-64 overflow-y-auto">
							<TrafficTable rows={daily} labelKey="date" valueKey="traffic" valueClass="text-blue-400" />
						</div>
					</div>
				</Tabs.Content>
			</Tabs.Root>
		</Drawer>
	);
}
