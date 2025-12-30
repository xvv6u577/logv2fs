import { useEffect, useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, reset, success } from "../store/message";
import Alert from "./alert";
import { useMonitoredDomainsList, useUpdateMonitoredDomains } from "../hooks/useQueries";

/**
 * 域名监控组件
 * 用于监控域名的到期时间
 */
function DomainMonitor() {
	// 使用 React Query 获取数据
	const { 
		data: monitoredDomains = [], 
		isLoading: loading, 
		error: domainsError, 
		refetch: refetchDomains 
	} = useMonitoredDomainsList();
	
	// 使用 mutation hook
	const updateDomainsMutation = useUpdateMonitoredDomains();
	
	// 表单状态
	const [newDomain, setNewDomain] = useState("");
	const [newRemark, setNewRemark] = useState("");

	const dispatch = useDispatch();
	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);

	// 通用样式类
	const styles = {
		button: "px-4 py-2 rounded-lg font-medium text-sm transition-colors focus:outline-none focus:ring-2",
		buttonPrimary: "bg-blue-600 hover:bg-blue-700 text-white focus:ring-blue-500",
		buttonSecondary: "bg-gray-600 hover:bg-gray-700 text-white focus:ring-gray-500",
		input: "w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-transparent",
		card: "bg-gray-800 rounded-lg shadow-lg hover:shadow-xl transition-all duration-200",
	};

	// 消息自动隐藏
	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(reset({}));
			}, 5000);
		}
	}, [message, dispatch]);
	
	// 错误处理
	useEffect(() => {
		if (domainsError) {
			dispatch(alert({ show: true, content: domainsError.toString() }));
		}
	}, [domainsError, dispatch]);

	// 提交表单 - 更新域名监控列表
	const handleAddDomain = (e) => {
		e.preventDefault();
		updateDomainsMutation.mutate(monitoredDomains, {
			onSuccess: (data) => {
				dispatch(success({ show: true, content: data.message }));
				refetchDomains();
			},
			onError: (err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			}
		});
	};

	// 添加新域名到列表
	const addNewDomain = () => {
		if (newDomain.length > 0 && newRemark.length > 0) {
			const tempDomains = monitoredDomains?.filter(item => item.domain === newDomain) || [];
			if (tempDomains.length === 0) {
				const updatedDomains = [...(monitoredDomains || []), { 
					domain: newDomain, 
					remark: newRemark, 
					days_to_expire: -1, 
					expired_date: "" 
				}];
				updateDomainsMutation.mutate(updatedDomains, {
					onSuccess: (data) => {
						dispatch(success({ show: true, content: "域名添加成功" }));
						refetchDomains();
					},
					onError: (err) => {
						dispatch(alert({ show: true, content: err.toString() }));
					}
				});
			}
			setNewDomain("");
			setNewRemark("");
		} else {
			dispatch(alert({ show: true, content: "域名和备注不能为空" }));
		}
	};

	// 从列表中移除域名
	const removeDomain = (domainToRemove) => {
		const updatedDomains = monitoredDomains?.filter(item => item.domain !== domainToRemove) || [];
		updateDomainsMutation.mutate(updatedDomains, {
			onSuccess: (data) => {
				dispatch(success({ show: true, content: "域名删除成功" }));
				refetchDomains();
			},
			onError: (err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			}
		});
	};

	// 域名卡片组件
	const DomainCard = ({ domain, index }) => (
		<div className={`${styles.card} p-6 relative`}>
			{/* 删除按钮 */}
			<button 
				className="absolute top-4 right-4 text-gray-400 hover:text-red-400 transition-colors"
				onClick={() => removeDomain(domain.domain)}
				aria-label="删除域名"
			>
				<svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>

			{/* 域名信息 */}
			<div className="mb-4">
				<h2 className="text-xl font-bold text-blue-300 mb-2">{domain.remark}</h2>
				<h3 className="text-lg font-semibold text-white mb-1">{domain.domain}</h3>
			</div>

			{/* 到期时间信息 */}
			<div className="text-center">
				<div className="text-3xl font-bold text-white mb-2">
					{domain.days_to_expire}天
				</div>
				<p className="text-gray-400 text-sm">
					到期时间: {domain.expired_date}
				</p>
			</div>
		</div>
	);

	return (
		<div className="min-h-screen bg-gray-900 text-white p-6">
			{/* 消息提示 */}
			<Alert 
				message={message.content} 
				type={message.type} 
				shown={message.show} 
				close={() => { dispatch(reset({})); }} 
			/>

			{/* 页面标题 */}
			<div className="mb-8">
				<h1 className="text-3xl font-bold mb-2">域名监控</h1>
				<p className="text-gray-400">监控域名到期时间，及时续费</p>
			</div>

			{/* 添加域名表单 */}
			<div className={`${styles.card} p-6 mb-8`}>
				<h3 className="text-lg font-semibold text-white mb-4">添加域名监控</h3>
				<form onSubmit={handleAddDomain}>
					<div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
						<input
							type="text"
							placeholder="域名"
							value={newDomain}
							onChange={(e) => setNewDomain(e.target.value.replace(/\s/g, ""))}
							className={styles.input}
							aria-label="域名输入框"
						/>
						<input
							type="text"
							placeholder="备注"
							value={newRemark}
							onChange={(e) => setNewRemark(e.target.value.replace(/\s/g, ""))}
							className={styles.input}
							aria-label="备注输入框"
						/>
						<button
							type="button"
							onClick={addNewDomain}
							className={`${styles.button} ${styles.buttonPrimary}`}
						>
							添加域名
						</button>
					</div>
					<button
						type="submit"
						className={`${styles.button} ${styles.buttonSecondary}`}
					>
						更新域名监控
					</button>
				</form>
			</div>

			{/* 域名列表 */}
			{loading ? (
				// 加载中状态
				<div className={`${styles.card} p-8 text-center`}>
					<div className="flex flex-col items-center">
						<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mb-4"></div>
						<h3 className="text-lg font-medium text-gray-300 mb-2">加载中...</h3>
						<p className="text-gray-400">正在获取域名监控数据</p>
					</div>
				</div>
			) : (monitoredDomains?.length || 0) > 0 ? (
				<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
					{monitoredDomains?.map((domain, index) => (
						<DomainCard key={index} domain={domain} index={index} />
					)) || []}
				</div>
			) : (
				// 空状态
				<div className={`${styles.card} p-8 text-center`}>
					<svg className="mx-auto h-12 w-12 text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9v-9m0-9v9" />
					</svg>
					<h3 className="text-lg font-medium text-gray-300 mb-2">暂无域名监控</h3>
					<p className="text-gray-400">添加域名开始监控到期时间</p>
				</div>
			)}
		</div>
	);
}

export default DomainMonitor;
