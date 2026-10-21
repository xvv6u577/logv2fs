import React from "react";
import { useState } from "react";

/**
 * 危险操作确认框。requireText 非空时，必须输入完全一致的字符串才能确认。
 */
export default function ConfirmDialog({
	open,
	title,
	message,
	confirmLabel = "确认",
	cancelLabel = "取消",
	danger = true,
	requireText = "",
	requireHint = "",
	onConfirm,
	onCancel,
}) {
	const [typed, setTyped] = useState("");
	if (!open) return null;

	const needMatch = Boolean(requireText);
	const matched = !needMatch || typed === requireText;

	const close = () => {
		setTyped("");
		onCancel?.();
	};

	const confirm = () => {
		if (!matched) return;
		setTyped("");
		onConfirm?.();
	};

	return (
		<div
			className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-[60] p-4"
			onClick={close}
		>
			<div
				className="bg-gray-800 rounded-lg w-full max-w-md shadow-xl"
				onClick={(e) => e.stopPropagation()}
			>
				<div className="p-6 border-b border-gray-700">
					<h3 className="text-lg font-bold text-white">{title}</h3>
				</div>
				<div className="p-6 space-y-4">
					<p className="text-gray-300 text-sm whitespace-pre-wrap">{message}</p>
					{needMatch && (
						<div>
							<label className="block text-sm text-gray-400 mb-2">
								{requireHint || `请输入「${requireText}」以确认`}
							</label>
							<input
								type="text"
								value={typed}
								onChange={(e) => setTyped(e.target.value)}
								className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:ring-2 focus:ring-red-500"
								autoFocus
							/>
						</div>
					)}
				</div>
				<div className="flex gap-3 p-6 pt-0">
					<button
						type="button"
						disabled={!matched}
						onClick={confirm}
						className={`flex-1 px-4 py-2 rounded-lg text-sm font-medium text-white transition-colors ${
							danger
								? "bg-red-600 hover:bg-red-700 disabled:opacity-40"
								: "bg-blue-600 hover:bg-blue-700 disabled:opacity-40"
						} disabled:cursor-not-allowed`}
					>
						{confirmLabel}
					</button>
					<button
						type="button"
						onClick={close}
						className="flex-1 px-4 py-2 rounded-lg text-sm font-medium text-white bg-gray-600 hover:bg-gray-700"
					>
						{cancelLabel}
					</button>
				</div>
			</div>
		</div>
	);
}
