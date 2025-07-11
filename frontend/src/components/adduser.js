import { useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, success } from "../store/message";
import { doRerender } from "../store/rerender";
import axios from "axios";

function AddUser({ btnName }) {

	const initialState = {
		email_as_id: "",
		name: "",
		path: "ray",
		role: "normal",
		uuid: "",
		remark: ""
	};
	const [showModal, setShowModal] = useState(false);
	const [{ email_as_id, name, path, role, uuid, remark }, setState] = useState(initialState);
	const [isLoading, setIsLoading] = useState(false);
	const [emailError, setEmailError] = useState("");

	const dispatch = useDispatch();
	const loginState = useSelector((state) => state.login);
	const rerenderSignal = useSelector((state) => state.rerender);

	const clearState = () => {
		setState({ ...initialState });
		setEmailError("");
	};

	// 验证邮箱格式：6位以上字母和数字组合
	const validateEmail = (email) => {
		if (email.length < 6) {
			return "至少需要6个字符";
		}
		if (!/^[a-zA-Z0-9]+$/.test(email)) {
			return "只能包含字母和数字";
		}
		return "";
	};

	const handleAddUser = async (e) => {
		e.preventDefault();
		
		// 验证邮箱格式
		const emailValidationError = validateEmail(email_as_id);
		if (emailValidationError) {
			setEmailError(emailValidationError);
			return;
		}

		setIsLoading(true);
		setEmailError("");

		try {
			await axios({
				method: "post",
				url: process.env.REACT_APP_API_HOST + "signup",
				headers: { token: loginState.token },
				data: {
					email_as_id,
					"password": email_as_id,
					role,
					name,
					path,
					status: "plain",
					uuid,
					remark
				},
			});
			
			dispatch(success({ show: true, content: "用户 " + (name || email_as_id) + " 添加成功！" }));
			dispatch(doRerender({ rerender: !rerenderSignal.rerender }));
			clearState();
			setShowModal(false);
		} catch (err) {
			// 处理数据库中已存在邮箱的错误
			if (err.response?.data?.error?.includes("already exists")) {
				setEmailError("该用户ID已存在，请使用其他ID");
			} else {
				dispatch(alert({ show: true, content: err.response?.data?.error || err.toString() }));
			}
		} finally {
			setIsLoading(false);
		}
	};

	const onChange = (e) => {
		const { name, value } = e.target;
		setState((prevState) => ({ ...prevState, [name]: value }));
		
		// 实时验证邮箱
		if (name === "email_as_id") {
			const error = validateEmail(value);
			setEmailError(error);
		}
	};

	const handleModalClick = (e) => {
		if (e.target === e.currentTarget) {
			setShowModal(false);
		}
	};

	return (
		<>
			<button
				className="group relative inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white 
					bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 
					rounded-lg shadow-lg hover:shadow-xl transform hover:scale-105 transition-all duration-200 
					focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-gray-900"
				type="button"
				onClick={() => setShowModal(true)}
			>
				<div className="absolute inset-0 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg blur opacity-75 
					group-hover:opacity-100 transition duration-200"></div>
				<div className="relative flex items-center">
					<svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
						<path strokeLinecap="round" strokeLinejoin="round" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
					</svg>
					{btnName}
				</div>
			</button>

			{showModal && (
				<div
					className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50 backdrop-blur-sm"
					onClick={handleModalClick}
				>
					<div className="relative w-full max-w-md transform transition-all duration-300 ease-out scale-100">
						<div className="relative bg-gradient-to-br from-gray-800 to-gray-900 rounded-2xl shadow-2xl border border-gray-700">
							{/* 装饰性渐变边框 */}
							<div className="absolute inset-0 bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 rounded-2xl blur opacity-20"></div>
							
							{/* 关闭按钮 */}
							<button
								type="button"
								className="absolute top-4 right-4 z-10 p-2 text-gray-400 hover:text-white hover:bg-gray-700 
									rounded-full transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500"
								onClick={() => setShowModal(false)}
							>
								<svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
									<path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
								</svg>
							</button>

							{/* 模态框内容 */}
							<div className="relative p-8">
								{/* 标题 */}
								<div className="text-center mb-8">
									<div className="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-r from-blue-600 to-purple-600 rounded-full mb-4">
										<svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
											<path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
										</svg>
									</div>
									<h3 className="text-2xl font-bold text-white mb-2">添加新用户</h3>
									<p className="text-gray-400">创建一个新的用户账户</p>
								</div>

								{/* 表单 */}
								<form className="space-y-6" onSubmit={handleAddUser}>
									{/* 用户ID字段 */}
									<div className="space-y-2">
										<label htmlFor="email" className="block text-sm font-medium text-gray-300">
											用户ID <span className="text-red-400">*</span>
										</label>
										<div className="relative">
											<div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
												<svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
												</svg>
											</div>
											<input
												type="text"
												name="email_as_id"
												id="email"
												value={email_as_id}
												onChange={onChange}
												className={`block w-full pl-10 pr-3 py-3 bg-gray-700 border rounded-lg 
													text-white placeholder-gray-400 focus:outline-none focus:ring-2 transition-all duration-200
													${emailError ? 'border-red-500 focus:ring-red-500' : 'border-gray-600 focus:ring-blue-500 focus:border-transparent'}`}
												placeholder="用户唯一标识符"
												required
											/>
										</div>
										{emailError ? (
											<p className="text-xs text-red-400 flex items-center">
												<svg className="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
													<path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
												</svg>
												{emailError}
											</p>
										) : (
											<p className="text-xs text-gray-500">6位以上字母和数字组合，将作为用户唯一标识</p>
										)}
									</div>

									{/* 姓名字段 */}
									<div className="space-y-2">
										<label htmlFor="name" className="block text-sm font-medium text-gray-300">
											显示名称
										</label>
										<div className="relative">
											<div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
												<svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
												</svg>
											</div>
											<input
												type="text"
												id="name"
												name="name"
												value={name}
												onChange={onChange}
												placeholder="用户显示名称（可选）"
												className="block w-full pl-10 pr-3 py-3 bg-gray-700 border border-gray-600 rounded-lg 
													text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 
													focus:border-transparent transition-all duration-200"
											/>
										</div>
									</div>

									{/* 备注字段 */}
									<div className="space-y-2">
										<label htmlFor="remark" className="block text-sm font-medium text-gray-300">
											用户备注
										</label>
										<div className="relative">
											<div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
												<svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-1l-4 4z" />
												</svg>
											</div>
											<textarea
												id="remark"
												name="remark"
												value={remark}
												onChange={onChange}
												placeholder="输入用户备注信息（可选）"
												rows="3"
												className="block w-full pl-10 pr-3 py-3 bg-gray-700 border border-gray-600 rounded-lg 
													text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 
													focus:border-transparent transition-all duration-200 resize-none"
											/>
										</div>
										<p className="text-xs text-gray-500">用于记录用户的特殊说明或备注信息</p>
									</div>

									{/* 角色选择 */}
									<div className="space-y-2">
										<label htmlFor="role" className="block text-sm font-medium text-gray-300">
											用户角色
										</label>
										<select
											id="role"
											name="role"
											value={role}
											onChange={onChange}
											className="block w-full px-3 py-3 bg-gray-700 border border-gray-600 rounded-lg 
												text-white focus:outline-none focus:ring-2 focus:ring-blue-500 
												focus:border-transparent transition-all duration-200"
										>
											<option value="normal">普通用户</option>
											<option value="admin">管理员</option>
										</select>
									</div>

									{/* 提交按钮 */}
									<button
										type="submit"
										disabled={isLoading || emailError}
										className="w-full flex items-center justify-center px-4 py-3 text-sm font-medium text-white 
											bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 
											rounded-lg shadow-lg hover:shadow-xl transform hover:scale-[1.02] transition-all duration-200 
											focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-gray-800
											disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
									>
										{isLoading ? (
											<>
												<svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
													<circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
													<path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
												</svg>
												创建中...
											</>
										) : (
											<>
												<svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
												</svg>
												创建用户
											</>
										)}
									</button>
								</form>
							</div>
						</div>
					</div>
				</div>
			)}
		</>
	);
}

export default AddUser;