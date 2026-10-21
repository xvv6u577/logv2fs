import React from "react";

/** 右侧抽屉。点击遮罩关闭。 */
export default function Drawer({ open, title, onClose, children, widthClass = "max-w-xl" }) {
	if (!open) return null;

	return (
		<div className="fixed inset-0 z-50 flex justify-end">
			<div className="absolute inset-0 bg-black bg-opacity-50" onClick={onClose} />
			<div
				className={`relative w-full ${widthClass} h-full bg-gray-800 shadow-2xl flex flex-col border-l border-gray-700`}
				onClick={(e) => e.stopPropagation()}
			>
				<div className="flex items-center justify-between p-4 border-b border-gray-700 flex-shrink-0">
					<h3 className="text-lg font-bold text-white truncate pr-4">{title}</h3>
					<button
						type="button"
						onClick={onClose}
						className="text-gray-400 hover:text-white p-1"
						aria-label="关闭"
					>
						<svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
				<div className="flex-1 overflow-y-auto p-4">{children}</div>
			</div>
		</div>
	);
}
