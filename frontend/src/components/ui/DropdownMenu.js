import React, { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

/**
 * 简易下拉菜单。items: [{ key, label, onClick, danger?, disabled? }]
 * align: 'left' | 'right'
 * 菜单经 portal 挂到 body（fixed），避免被表格 overflow 裁切。
 */
export default function DropdownMenu({ label, items = [], align = "right", buttonClassName = "" }) {
	const [open, setOpen] = useState(false);
	const [pos, setPos] = useState(null);
	const btnRef = useRef(null);
	const menuRef = useRef(null);

	const updatePos = useCallback(() => {
		const btn = btnRef.current;
		if (!btn) return;
		const r = btn.getBoundingClientRect();
		const menuW = menuRef.current?.offsetWidth || 160;
		const menuH = menuRef.current?.offsetHeight || 0;
		let left = align === "left" ? r.left : r.right - menuW;
		let top = r.bottom + 4;
		left = Math.max(8, Math.min(left, window.innerWidth - menuW - 8));
		if (menuH > 0 && top + menuH > window.innerHeight - 8 && r.top > menuH + 8) {
			top = r.top - menuH - 4;
		}
		setPos({ top, left });
	}, [align]);

	const toggle = () => {
		if (open) {
			setOpen(false);
			setPos(null);
			return;
		}
		const btn = btnRef.current;
		if (btn) {
			const r = btn.getBoundingClientRect();
			const approxW = 160;
			let left = align === "left" ? r.left : r.right - approxW;
			left = Math.max(8, Math.min(left, window.innerWidth - approxW - 8));
			setPos({ top: r.bottom + 4, left });
		}
		setOpen(true);
	};

	useLayoutEffect(() => {
		if (!open) return;
		updatePos();
	}, [open, updatePos, items.length]);

	useEffect(() => {
		if (!open) return;
		const onDoc = (e) => {
			const t = e.target;
			if (btnRef.current?.contains(t) || menuRef.current?.contains(t)) return;
			setOpen(false);
			setPos(null);
		};
		const onScrollOrResize = () => updatePos();
		document.addEventListener("mousedown", onDoc);
		window.addEventListener("scroll", onScrollOrResize, true);
		window.addEventListener("resize", onScrollOrResize);
		return () => {
			document.removeEventListener("mousedown", onDoc);
			window.removeEventListener("scroll", onScrollOrResize, true);
			window.removeEventListener("resize", onScrollOrResize);
		};
	}, [open, updatePos]);

	return (
		<>
			<button
				ref={btnRef}
				type="button"
				onClick={toggle}
				className={
					buttonClassName ||
					"px-3 py-1.5 rounded-lg text-xs font-medium bg-gray-600 hover:bg-gray-500 text-white"
				}
			>
				{label}
			</button>
			{open &&
				pos &&
				createPortal(
					<div
						ref={menuRef}
						style={{ position: "fixed", top: pos.top, left: pos.left, zIndex: 9999 }}
						className="min-w-[10rem] bg-gray-700 border border-gray-600 rounded-lg shadow-lg py-1"
					>
						{items.map((item) => (
							<button
								key={item.key}
								type="button"
								disabled={item.disabled}
								onClick={() => {
									setOpen(false);
									setPos(null);
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
					</div>,
					document.body
				)}
		</>
	);
}
