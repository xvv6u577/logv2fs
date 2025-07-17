import { useState, useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, success, reset } from "../store/message";
import { doRerender } from "../store/rerender";
import Alert from "./alert";
import axios from "axios";

const AddNode = () => {
	const [nodes, setNodes] = useState([]);
	const [enableOpenai, setEnableOpenai] = useState(false);
	
	const initialState = {
		type: "reality",
		remark: "",
		domain: "",
		ip: "",
		uuid: "",
		path: "",
		sni: "",
		server_port: "",
	};
	
	const [formData, setFormData] = useState(initialState);
	const { type, remark, domain, uuid, path, sni, ip, server_port } = formData;

	const dispatch = useDispatch();
	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);
	const rerenderSignal = useSelector((state) => state.rerender);

	// 通用样式类
	const styles = {
		button: "px-4 py-2 rounded-lg font-medium text-sm transition-colors focus:outline-none focus:ring-2",
		buttonPrimary: "bg-blue-600 hover:bg-blue-700 text-white focus:ring-blue-500",
		buttonSecondary: "bg-gray-600 hover:bg-gray-700 text-white focus:ring-gray-500",
		buttonDanger: "bg-red-600 hover:bg-red-700 text-white focus:ring-red-500",
		buttonSuccess: "bg-green-600 hover:bg-green-700 text-white focus:ring-green-500",
		input: "w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-transparent",
		select: "w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:ring-2 focus:ring-blue-500",
		card: "bg-gray-800 rounded-lg shadow-lg hover:shadow-xl transition-all duration-200",
		label: "block text-sm font-medium text-gray-300 mb-2",
		badge: "px-2 py-1 rounded-full text-xs font-medium",
		badgeReality: "bg-blue-900 text-blue-300",
		badgeHysteria: "bg-purple-900 text-purple-300",
		badgeVless: "bg-green-900 text-green-300",
	};

	const clearState = () => {
		setFormData({ ...initialState });
		setEnableOpenai(false);
	};

	const onChange = (e) => {
		const name = e.target.name;
		const value = e.target.value.replace(/\s/g, "");
		setFormData((prevState) => ({ ...prevState, [name]: value }));
	};

	const returnType = (type) => {
		switch (type) {
			case "reality": return "Reality";
			case "hysteria2": return "Hysteria2";
			case "vlessCDN": return "VLessCDN";
			default: return type;
		}
	};

	const getBadgeStyle = (type) => {
		switch (type) {
			case "reality": return styles.badgeReality;
			case "hysteria2": return styles.badgeHysteria;
			case "vlessCDN": return styles.badgeVless;
			default: return styles.badgeReality;
		}
	};

	useEffect(() => {
		axios
			.get(process.env.REACT_APP_API_HOST + "t7k033", {
				headers: { token: loginState.token },
			})
			.then((response) => {
				setNodes(response.data);
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	}, [rerenderSignal, loginState.token, dispatch]);

	const handleAddNode = (e) => {
		e.preventDefault();
		axios({
			method: "put",
			url: process.env.REACT_APP_API_HOST + "759b0v",
			headers: { token: loginState.token },
			data: nodes,
		})
			.then((response) => {
				dispatch(success({ show: true, content: response.data.message }));
				dispatch(doRerender({ rerender: !rerenderSignal.rerender }));
				clearState();
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	};

	const addNodeToList = () => {
		if (domain.length > 0 && remark.length > 0) {
			setNodes((prevState) => ([
				...prevState,
				{
					type,
					remark,
					domain,
					ip,
					server_port,
					enable_openai: enableOpenai,
					uuid,
					path,
					sni
				}
			]));
			clearState();
		} else {
			dispatch(alert({ show: true, content: "域名和备注字段不能为空" }));
		}
	};

	const removeNode = (remarkToRemove) => {
		setNodes((prevState) => (prevState.filter((n) => n.remark !== remarkToRemove)));
	};

	const handleCopyNode = (nodeToCopy) => {
		const { enable_openai, ...nodeData } = nodeToCopy;

		setFormData({ ...initialState, ...nodeData });
		setEnableOpenai(enable_openai || false);

		window.scrollTo({ top: 0, behavior: "smooth" });
		
		dispatch(success({ show: true, content: `已复制节点 ${nodeToCopy.remark} 的信息到表单` }));
	};

	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(reset({}));
			}, 5000);
		}
	}, [message, dispatch]);

	// 节点卡片组件
	const NodeCard = ({ node, index }) => (
		<div className={`${styles.card} p-6 relative`}>
			<div className="absolute top-4 right-4 flex items-center space-x-2">
				<button
					onClick={() => handleCopyNode(node)}
					className="p-1 rounded-full text-gray-400 hover:bg-gray-700 hover:text-blue-400 transition-all duration-200"
					title="复制到表单"
				>
					<svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
						<path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
					</svg>
				</button>
				<button
					onClick={() => removeNode(node.remark)}
					className="p-1 rounded-full text-gray-400 hover:bg-gray-700 hover:text-red-400 transition-all duration-200"
					title="删除节点"
				>
					<svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
					</svg>
				</button>
			</div>

			<div className="mb-4">
				<div className="flex items-center space-x-3 mb-2">
					<span className={`${styles.badge} ${getBadgeStyle(node.type)}`}>
						{returnType(node.type)}
					</span>
					<span className="bg-gray-700 text-gray-300 px-2 py-1 rounded text-sm font-mono">
						#{index + 1}
					</span>
				</div>
				<h3 className="text-lg font-semibold text-white">{node.remark}</h3>
				<p className="text-gray-400">{node.domain}</p>
			</div>

			<div className="grid grid-cols-2 gap-4 text-sm">
				<div>
					<span className="text-gray-400">IP: </span>
					<span className="text-white font-mono">{node.ip || "None"}</span>
				</div>
				<div>
					<span className="text-gray-400">Port: </span>
					<span className="text-white font-mono">{node.server_port || "None"}</span>
				</div>
				<div>
					<span className="text-gray-400">OpenAI: </span>
					<span className={`font-medium ${node.enable_openai ? "text-green-400" : "text-red-400"}`}>
						{node.enable_openai ? "Yes" : "No"}
					</span>
				</div>
				<div>
					<span className="text-gray-400">UUID: </span>
					<span className="text-white font-mono text-xs">
						{node.uuid ? `${node.uuid.substring(0, 8)}...` : "None"}
					</span>
				</div>
				{node.path && (
					<div className="col-span-2">
						<span className="text-gray-400">Path: </span>
						<span className="text-white font-mono">{node.path}</span>
					</div>
				)}
				{node.sni && (
					<div className="col-span-2">
						<span className="text-gray-400">SNI: </span>
						<span className="text-white font-mono">{node.sni}</span>
					</div>
				)}
			</div>
		</div>
	);

	return (
		<div className="min-h-screen bg-gray-900 text-white p-6">
			<Alert 
				message={message.content} 
				type={message.type} 
				shown={message.show} 
				close={() => { dispatch(reset({})); }} 
			/>

			{/* 页面标题 */}
			<div className="mb-8">
				<h1 className="text-3xl font-bold mb-2">添加节点</h1>
				<p className="text-gray-400">管理和配置网络节点</p>
			</div>

			{/* 添加节点表单 */}
			<div className={`${styles.card} p-6 mb-8`}>
				<h2 className="text-xl font-semibold text-white mb-6">添加新节点</h2>
				
				<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
					{/* 基本信息 */}
					<div className="space-y-4">
						<h3 className="text-lg font-medium text-gray-300 border-b border-gray-600 pb-2">基本信息</h3>
						
						<div>
							<label className={styles.label}>节点类型</label>
							<select
								name="type"
								onChange={onChange}
								value={type}
								className={styles.select}
							>
								<option value="reality">Reality</option>
								<option value="hysteria2">Hysteria2</option>
								<option value="vlessCDN">VLessCDN</option>
							</select>
						</div>

						<div>
							<label className={styles.label}>备注名称 *</label>
							<input
								type="text"
								name="remark"
								onChange={onChange}
								value={remark}
								className={styles.input}
								placeholder="节点备注名称"
							/>
						</div>

						<div>
							<label className={styles.label}>域名 *</label>
							<input
								type="text"
								name="domain"
								onChange={onChange}
								value={domain}
								className={styles.input}
								placeholder="example.com"
							/>
						</div>
					</div>

					{/* 连接信息 */}
					<div className="space-y-4">
						<h3 className="text-lg font-medium text-gray-300 border-b border-gray-600 pb-2">连接信息</h3>
						
						<div>
							<label className={styles.label}>IP 地址/域名</label>
							<input
								type="text"
								name="ip"
								onChange={onChange}
								value={ip}
								className={styles.input}
								placeholder="192.168.1.1 或 example.com"
							/>
						</div>

						<div>
							<label className={styles.label}>端口</label>
							<input
								type="text"
								name="server_port"
								onChange={onChange}
								value={server_port}
								className={styles.input}
								placeholder="443"
							/>
						</div>

						<div>
							<label className={styles.label}>UUID</label>
							<input
								type="text"
								name="uuid"
								onChange={onChange}
								value={uuid}
								className={styles.input}
								placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
							/>
						</div>
					</div>

					{/* 高级选项 */}
					<div className="space-y-4">
						<h3 className="text-lg font-medium text-gray-300 border-b border-gray-600 pb-2">高级选项</h3>
						
						<div>
							<label className={styles.label}>路径</label>
							<input
								type="text"
								name="path"
								onChange={onChange}
								value={path}
								className={styles.input}
								placeholder="/path"
							/>
						</div>

						<div>
							<label className={styles.label}>SNI</label>
							<input
								type="text"
								name="sni"
								onChange={onChange}
								value={sni}
								className={styles.input}
								placeholder="example.com"
							/>
						</div>

						<div>
							<label className="flex items-center space-x-3 cursor-pointer">
								<input
									type="checkbox"
									onChange={(e) => setEnableOpenai(e.target.checked)}
									checked={enableOpenai}
									className="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500 focus:ring-2"
								/>
								<span className={styles.label + " mb-0"}>启用 OpenAI</span>
							</label>
						</div>
					</div>
				</div>

				{/* 操作按钮 */}
				<div className="flex space-x-4 mt-8 pt-6 border-t border-gray-700">
					<button
						type="button"
						onClick={addNodeToList}
						className={`${styles.button} ${styles.buttonPrimary}`}
					>
						添加到列表
					</button>
					<button
						type="button"
						onClick={clearState}
						className={`${styles.button} ${styles.buttonSecondary}`}
					>
						清空表单
					</button>
				</div>
			</div>

			{/* 节点列表 */}
			<div className="mb-8">
				<div className="flex items-center justify-between mb-6">
					<h2 className="text-xl font-semibold text-white">
						节点列表 ({nodes.length})
					</h2>
				</div>

				{/* 节点网格 */}
				<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
					{nodes.length === 0 ? (
						<div className={`${styles.card} p-8 text-center col-span-full`}>
							<svg className="mx-auto h-12 w-12 text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
							</svg>
							<h3 className="text-lg font-medium text-gray-300 mb-2">暂无节点</h3>
							<p className="text-gray-400">添加节点开始管理</p>
						</div>
					) : (
						nodes.map((node, index) => (
							<NodeCard key={index} node={node} index={index} />
						))
					)}
				</div>
			</div>

			{/* 提交按钮 */}
			{nodes.length > 0 && (
				<div className="flex justify-center">
					<form onSubmit={handleAddNode}>
						<button
							type="submit"
							className={`${styles.button} ${styles.buttonSuccess} px-8 py-3 text-lg`}
						>
							更新节点配置 ({nodes.length} 个节点)
						</button>
					</form>
				</div>
			)}
		</div>
	);
};

export default AddNode;