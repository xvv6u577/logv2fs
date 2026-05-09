import { useDispatch } from "react-redux";
import { alert, success } from "../store/message";
import { useAddUser } from "../hooks/useQueries";

function AddUser({ btnName }) {
	const dispatch = useDispatch();
	const addUserMutation = useAddUser();

	// 基于一个新的随机 uuid 派生用户字段：
	// - email_as_id：uuid 去掉所有横杠（32 位 hex）
	// - name：       "newuser-" + uuid 的第一段（横杠分割后的第 0 段，8 位）
	// 节点协议层的 uuid 由后端 SignUp 自行生成，这里无需关心。
	const buildNewUser = () => {
		const rawUUID = crypto.randomUUID();
		const emailAsId = rawUUID.replace(/-/g, "");
		const name = `newuser-${rawUUID.split("-")[0]}`;

		return {
			email_as_id: emailAsId,
			password: emailAsId,
			role: "normal",
			name,
			path: "ray",
			status: "plain",
			uuid: "",
			remark: "",
		};
	};

	const handleAddUser = () => {
		if (addUserMutation.isPending) return;

		const userData = buildNewUser();

		addUserMutation.mutate(userData, {
			onSuccess: () => {
				dispatch(success({ show: true, content: `用户 ${userData.name} 添加成功！` }));
			},
			onError: (err) => {
				const errMsg = err.response?.data?.error || err.toString();
				dispatch(alert({ show: true, content: errMsg }));
			},
		});
	};

	return (
		<button
			type="button"
			onClick={handleAddUser}
			disabled={addUserMutation.isPending}
			aria-label={btnName}
			className="group relative inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white
				bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700
				rounded-lg shadow-lg hover:shadow-xl transform hover:scale-105 transition-all duration-200
				focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-gray-900
				disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
		>
			<div className="absolute inset-0 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg blur opacity-75
				group-hover:opacity-100 transition duration-200"></div>
			<div className="relative flex items-center">
				{addUserMutation.isPending ? (
					<>
						<svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
							<circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
							<path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
						</svg>
						创建中...
					</>
				) : (
					<>
						<svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
							<path strokeLinecap="round" strokeLinejoin="round" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
						</svg>
						{btnName}
					</>
				)}
			</div>
		</button>
	);
}

export default AddUser;
