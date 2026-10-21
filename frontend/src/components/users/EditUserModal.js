import React, { useState, useEffect } from "react";

const inputClass =
	"w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-transparent";

export default function EditUserModal({ open, user, onClose, onSubmit, submitting }) {
	const [form, setForm] = useState({
		name: "",
		role: "normal",
		password: "",
		confirmPassword: "",
		remark: "",
	});
	const [error, setError] = useState("");

	useEffect(() => {
		if (open && user) {
			setForm({
				name: user.name || "",
				role: user.role || "normal",
				password: "",
				confirmPassword: "",
				remark: user.remark || "",
			});
			setError("");
		}
	}, [open, user]);

	if (!open || !user) return null;

	const handleSave = () => {
		if (!form.name.trim()) {
			setError("用户名不能为空");
			return;
		}
		if (form.password) {
			if (form.password.length < 8) {
				setError("密码长度至少需要 8 个字符");
				return;
			}
			if (!/[A-Za-z]/.test(form.password) || !/[0-9]/.test(form.password)) {
				setError("密码必须同时包含字母和数字");
				return;
			}
			if (form.password !== form.confirmPassword) {
				setError("两次输入的密码不一致");
				return;
			}
		}
		const data = {
			email_as_id: user.email_as_id,
			name: form.name.trim(),
			role: form.role,
			remark: form.remark.trim(),
		};
		if (form.password) data.password = form.password;
		onSubmit(data);
	};

	return (
		<div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4" onClick={onClose}>
			<div className="bg-gray-800 rounded-lg p-6 w-full max-w-md" onClick={(e) => e.stopPropagation()}>
				<h3 className="text-xl font-bold mb-4 text-white">编辑用户</h3>
				<div className="space-y-4">
					<div>
						<label className="block text-sm text-gray-300 mb-1">邮箱 ID</label>
						<input type="text" value={user.email_as_id || ""} disabled className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-gray-400 cursor-not-allowed" />
					</div>
					<div>
						<label className="block text-sm text-gray-300 mb-1">用户名</label>
						<input type="text" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} className={inputClass} />
					</div>
					<div>
						<label className="block text-sm text-gray-300 mb-1">角色</label>
						<select value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value })} className={inputClass}>
							<option value="normal">普通用户</option>
							<option value="admin">管理员</option>
						</select>
					</div>
					<div>
						<label className="block text-sm text-gray-300 mb-1">备注</label>
						<textarea value={form.remark} onChange={(e) => setForm({ ...form, remark: e.target.value })} className={inputClass} rows={2} />
					</div>
					<div className="border-t border-gray-600 pt-4">
						<p className="text-sm text-gray-400 mb-2">修改密码（可选，留空则不变）</p>
						<input type="password" placeholder="新密码（≥8 位，字母+数字）" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} className={`${inputClass} mb-2`} />
						<input type="password" placeholder="确认新密码" value={form.confirmPassword} onChange={(e) => setForm({ ...form, confirmPassword: e.target.value })} className={inputClass} />
					</div>
					{error && <p className="text-sm text-red-400">{error}</p>}
				</div>
				<div className="flex gap-3 mt-6">
					<button type="button" disabled={submitting} onClick={handleSave} className="flex-1 px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 text-white text-sm disabled:opacity-50">
						{submitting ? "保存中…" : "保存"}
					</button>
					<button type="button" onClick={onClose} className="flex-1 px-4 py-2 rounded-lg bg-gray-600 hover:bg-gray-700 text-white text-sm">
						取消
					</button>
				</div>
			</div>
		</div>
	);
}
