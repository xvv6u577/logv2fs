import React from 'react';
import { useSelector } from 'react-redux';
import { Navigate, useLocation } from 'react-router-dom';

/**
 * PrivateRoute 用于"任意已登录用户均可访问"的路由。
 * 未登录则跳转到 /login，并把当前路径塞进 state，便于登录后回跳。
 */
export function PrivateRoute({ children }) {
	const loginState = useSelector((state) => state.login);
	const location = useLocation();

	if (loginState.isLogin !== true) {
		return <Navigate to="/login" state={{ from: location }} replace />;
	}
	return children;
}

/**
 * AdminRoute 用于"仅管理员可访问"的路由。
 * 安全模型：前端守卫只是 UX 优化，真正的访问控制由后端 AdminOnly 中间件保证。
 *   - 未登录 → /login
 *   - 已登录但非 admin → /mypanel（用户主面板）
 */
export function AdminRoute({ children }) {
	const loginState = useSelector((state) => state.login);
	const location = useLocation();

	if (loginState.isLogin !== true) {
		return <Navigate to="/login" state={{ from: location }} replace />;
	}
	const role = loginState?.jwt?.role;
	if (role !== 'admin') {
		return <Navigate to="/mypanel" replace />;
	}
	return children;
}

export default PrivateRoute;
