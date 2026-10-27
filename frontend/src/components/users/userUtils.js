/** 订阅链接；excludeHK 时追加 ?exclude=hk */
export function subUrl(type, user, excludeHK = false) {
	const base = process.env.REACT_APP_FILE_AND_SUB_URL || "";
	return `${base}/${type}/${user.email_as_id}${excludeHK ? "?exclude=hk" : ""}`;
}

export function getStatusLabel(status) {
	if (status === "plain") return "在用";
	if (status === "disabled") return "已禁用";
	if (status === "overdue") return "欠费停用"; // 历史遗留，运维迁完后应消失
	return status || "未知";
}

export function getStatusBadgeClass(status) {
	if (status === "plain") return "bg-green-900 text-green-300";
	if (status === "overdue") return "bg-orange-900 text-orange-300";
	return "bg-gray-700 text-gray-400";
}

/** 禁用类：disabled 或历史 overdue（标签「禁用」计数/过滤都算进去） */
export function isInactiveStatus(status) {
	return status === "disabled" || status === "overdue";
}

export function getTrafficNumber(user, period) {
	switch (period) {
		case "daily":
			return user.daily_logs?.[0]?.traffic ?? 0;
		case "monthly":
			return user.monthly_logs?.[0]?.traffic ?? 0;
		case "used":
			return user.used || 0;
		default:
			return 0;
	}
}

export function sortUsers(list, sortBy, sortDir = "desc") {
	const dir = sortDir === "asc" ? 1 : -1;
	const arr = [...list];
	arr.sort((a, b) => {
		let cmp = 0;
		switch (sortBy) {
			case "name":
				cmp = (a.name || a.email_as_id || "").localeCompare(b.name || b.email_as_id || "");
				break;
			case "status":
				cmp = (a.status || "").localeCompare(b.status || "");
				break;
			case "today":
				cmp = getTrafficNumber(a, "daily") - getTrafficNumber(b, "daily");
				break;
			case "monthly":
				cmp = getTrafficNumber(a, "monthly") - getTrafficNumber(b, "monthly");
				break;
			case "used":
				cmp = getTrafficNumber(a, "used") - getTrafficNumber(b, "used");
				break;
			case "update_time":
				cmp = new Date(a.updated_at || 0) - new Date(b.updated_at || 0);
				break;
			case "role":
			default:
				if (a.role === "admin" && b.role !== "admin") cmp = -1;
				else if (b.role === "admin" && a.role !== "admin") cmp = 1;
				else cmp = getTrafficNumber(b, "monthly") - getTrafficNumber(a, "monthly");
				// role 默认已经是「admin 优先 + 月流量降序」，不再乘 dir
				return cmp;
		}
		return cmp * dir;
	});
	return arr;
}

export function filterUsers(users, { searchTerm, filterStatus }) {
	let filtered = users;
	if (filterStatus === "plain") {
		filtered = filtered.filter((u) => u.status === "plain");
	} else if (filterStatus === "disabled") {
		// 历史 overdue 归入「禁用」页签
		filtered = filtered.filter((u) => isInactiveStatus(u.status));
	}
	if (searchTerm) {
		const q = searchTerm.toLowerCase();
		filtered = filtered.filter(
			(u) =>
				u.email_as_id?.toLowerCase().includes(q) ||
				(Array.isArray(u.email_as_id_aliases) && u.email_as_id_aliases.some((a) => a?.toLowerCase().includes(q))) ||
				u.name?.toLowerCase().includes(q) ||
				u.role?.toLowerCase().includes(q) ||
				u.remark?.toLowerCase().includes(q)
		);
	}
	return filtered;
}

export function countByStatus(users) {
	let plain = 0;
	let disabled = 0;
	for (const u of users) {
		if (u.status === "plain") plain += 1;
		else if (isInactiveStatus(u.status)) disabled += 1;
	}
	return { all: users.length, plain, disabled };
}
