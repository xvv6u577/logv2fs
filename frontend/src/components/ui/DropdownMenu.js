import React, { useEffect, useRef, useState } from "react";

/**
 * 简易下拉菜单。items: [{ key, label, onClick, danger?, disabled? }]
 * align: 'left' | 'right'
 */
export default function DropdownMenu({ label, items = [], align = "right", buttonClassName = "" }) {
	const [open, setOpen] = useState(false);
	const ref = useRef(null);

	useEffect(() => {
		if (!open) return;
		const onDoc = (e) => {
			if (ref.current && !ref.current.contains(e.target)) setOpen(false);
		};
		document.addEventListener("mousedown", onDoc);
		return () => document.removeEventListener("mousedown", onDoc);
	}, [open]);

	return (
		<div className="relative inline-block" ref={ref}>
			<button
				type="button"
				onClick={() => setOpen((v) => !v)}
				className={
					buttonClassName ||
					"px-3 py-1.5 rounded-lg text-xs font-medium bg-gray-600 hover:bg-gray-500 text-white"
				}
			>
				{label}
			</button>
			{open && (
				<div
					className={`absolute z-40 mt-1 min-w-[10rem] bg-gray-700 border border-gray-600 rounded-lg shadow-lg py-1 ${
						align === "left" ? "left-0" : "right-0"
					}`}
				>
					{items.map((item) => (
						<button
							key={item.key}
							type="button"
							disabled={item.disabled}
							onClick={() => {
								setOpen(false);
								item.onClick?.();
							}}
							className={`block w-full text-left px-3 py-2 text-sm transition-colors disabled:opacity-40 ${
								item.danger
									? "text-red-300 hover:bg-red-900 hover:bg-opacity-40"
									: "text-gray-200 hover:bg-gray-600"
							}`}
						>
							{item.label}
						</button>
					))}
				</div>
			)}
		</div>
	);
}
