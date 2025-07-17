import React, { useState, useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import { Navigate } from "react-router-dom";
import axios from "axios";
import { alert, success } from "../store/message";
import { login } from "../store/login";
import Alert from "./alert";

const Login = () => {
	const [name, setName] = useState("");
	const [password, setPassword] = useState("");
	const [isLoading, setIsLoading] = useState(false);
	const [showPassword, setShowPassword] = useState(false);

	const dispatch = useDispatch();

	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);

	// 通用样式类
	const styles = {
		input: "w-full px-4 py-3 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200",
		button: "w-full px-4 py-3 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-gray-900 disabled:opacity-50 disabled:cursor-not-allowed",
		label: "block text-sm font-medium text-gray-300 mb-2",
		card: "bg-gray-800 rounded-xl shadow-2xl border border-gray-700",
	};

	const year = new Date().getFullYear();

	const handleSubmit = (e) => {
		e.preventDefault();
		setIsLoading(true);

		axios
			.post(process.env.REACT_APP_API_HOST + "login", {
				email_as_id: name,
				password: password,
			})
			.then((response) => {
				if (response.data) {
					localStorage.setItem("token", JSON.stringify(response.data.token));
					dispatch(login({ token: response.data.token }));
					dispatch(success({ show: true, content: "登录成功！" }));
				}
			})
			.catch((err) => {
				if (err.response) {
					dispatch(alert({ show: true, content: err.response.data.error || "登录失败" }));
				} else {
					dispatch(alert({ show: true, content: "用户名或密码错误！" }));
				}
				console.log(err.toString());
			})
			.finally(() => {
				setIsLoading(false);
			});
	};

	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(alert({ show: false }));
			}, 5000);
		}
	}, [dispatch, message]);

	if (loginState.isLogin) {
		return <Navigate to="/mypanel" />;
	}

	return (
		<div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 flex items-center justify-center p-4">
			<Alert 
				message={message.content} 
				type={message.type} 
				shown={message.show} 
				close={() => { dispatch(alert({ show: false })); }} 
			/>
			
			{/* 背景装饰 */}
			<div className="absolute inset-0 overflow-hidden">
				<div className="absolute -top-40 -right-40 w-80 h-80 bg-blue-500 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-pulse"></div>
				<div className="absolute -bottom-40 -left-40 w-80 h-80 bg-purple-500 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-pulse"></div>
			</div>

			{/* 登录卡片 */}
			<div className={`${styles.card} p-8 w-full max-w-md relative z-10`}>
				{/* 头部 */}
				<div className="text-center mb-8">
					{/* Logo/图标 */}
					<div className="w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center mx-auto mb-4">
						<svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
						</svg>
					</div>
					<h1 className="text-2xl font-bold text-white mb-2">欢迎回来</h1>
					<p className="text-gray-400">请登录您的账户</p>
				</div>

				{/* 登录表单 */}
				<form onSubmit={handleSubmit} className="space-y-6">
					{/* 用户名输入 */}
					<div>
						<label htmlFor="username" className={styles.label}>
							用户名
						</label>
						<div className="relative">
							<div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
								<svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
								</svg>
							</div>
							<input
								id="username"
								type="text"
								placeholder="请输入用户名"
								value={name}
								onChange={(e) => setName(e.target.value)}
								className={`${styles.input} pl-10`}
								required
								disabled={isLoading}
							/>
						</div>
					</div>

					{/* 密码输入 */}
					<div>
						<label htmlFor="password" className={styles.label}>
							密码
						</label>
						<div className="relative">
							<div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
								<svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
								</svg>
							</div>
							<input
								id="password"
								type={showPassword ? "text" : "password"}
								placeholder="请输入密码"
								value={password}
								onChange={(e) => setPassword(e.target.value)}
								className={`${styles.input} pl-10 pr-10`}
								required
								disabled={isLoading}
							/>
							<button
								type="button"
								className="absolute inset-y-0 right-0 pr-3 flex items-center"
								onClick={() => setShowPassword(!showPassword)}
								disabled={isLoading}
							>
								{showPassword ? (
									<svg className="h-5 w-5 text-gray-400 hover:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L3 3m6.878 6.878L21 21" />
									</svg>
								) : (
									<svg className="h-5 w-5 text-gray-400 hover:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
										<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
									</svg>
								)}
							</button>
						</div>
					</div>

					{/* 登录按钮 */}
					<button
						type="submit"
						className={styles.button}
						disabled={isLoading || !name.trim() || !password.trim()}
					>
						{isLoading ? (
							<div className="flex items-center justify-center">
								<svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
									<circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
									<path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
								</svg>
								登录中...
							</div>
						) : (
							"登录"
						)}
					</button>
				</form>

				{/* 底部信息 */}
				<div className="mt-8 text-center">
					<p className="text-gray-400 text-sm">
						© {year} Logv2 App. 保留所有权利.
					</p>
				</div>
			</div>
		</div>
	);
};

export default Login;
