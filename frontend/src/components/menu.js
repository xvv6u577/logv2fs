import { useSelector, useDispatch } from "react-redux";
import { useState } from "react";
import { logout } from "../store/login";
import {logoBase64} from "./logoImage"

const Menu = () => {
	const loginState = useSelector((state) => state.login);
	const dispatch = useDispatch();
	
	// 移动端折叠菜单状态管理
	const [isNodesOpen, setIsNodesOpen] = useState(false);
	const [isPaymentOpen, setIsPaymentOpen] = useState(false);
	const [isClientsOpen, setIsClientsOpen] = useState(false);

	const handleLogout = (e) => {
		dispatch(logout());
	};

	// 切换Nodes菜单展开状态
	const toggleNodesMenu = () => {
		setIsNodesOpen(!isNodesOpen);
	};

	// 切换Payment菜单展开状态
	const togglePaymentMenu = () => {
		setIsPaymentOpen(!isPaymentOpen);
	};

	// 切换Clients菜单展开状态
	const toggleClientsMenu = () => {
		setIsClientsOpen(!isClientsOpen);
	};

	// 下拉箭头组件
	const ChevronIcon = ({ isOpen }) => (
		<svg 
			className={`ml-1 h-4 w-4 transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`}
			fill="none" 
			stroke="currentColor" 
			viewBox="0 0 24 24"
		>
			<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
		</svg>
	);

	return (
		<header className="text-gray-400 bg-gray-900 body-font">
			<div className="mx-auto flex flex-wrap p-5 flex-col md:flex-row items-center">
				<a href="/user" className="flex title-font font-medium items-center text-white mb-4 md:mb-0">
					<img alt="logo" src={logoBase64}></img>
					<span className="ml-3 text-xl">Logv2fs Frontend</span>
				</a>
				<nav className="md:mr-auto md:ml-4 md:py-1 md:pl-4 md:border-l md:border-gray-700 flex flex-wrap items-center text-base justify-center">
					{loginState.jwt.role === "admin" && (
						<>
							<a className="mr-5 hover:text-white" href="/user">User</a>
							
							{/* Nodes 二级菜单 - 桌面端悬停，移动端点击 */}
							<div className="relative mr-5 group">
								{/* 桌面端悬停触发 */}
								<button 
									className="hidden md:flex items-center hover:text-white"
									onClick={toggleNodesMenu}
								>
									Nodes
									<ChevronIcon isOpen={false} />
								</button>
								
								{/* 移动端点击触发 */}
								<button 
									className="flex md:hidden items-center hover:text-white"
									onClick={toggleNodesMenu}
								>
									Nodes
									<ChevronIcon isOpen={isNodesOpen} />
								</button>
								
								{/* 桌面端下拉菜单 - 悬停显示 */}
								<div className="hidden md:block absolute left-0 mt-2 w-48 bg-gray-800 rounded-md shadow-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 z-50">
									<div className="py-1">
										<a href="/nodes" className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 hover:text-white">
											节点监控
										</a>
									<a href="/addnode" className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 hover:text-white">
										添加节点
									</a>
									</div>
								</div>
							</div>
							
							{/* 移动端 Nodes 折叠菜单 */}
							<div className="block md:hidden w-full">
								{isNodesOpen && (
									<div className="ml-4 mt-2 space-y-1">
										<a href="/nodes" className="block text-sm text-gray-300 hover:text-white py-1">
											节点监控
										</a>
									<a href="/addnode" className="block text-sm text-gray-300 hover:text-white py-1">
										添加节点
									</a>
									</div>
								)}
							</div>
							
							{/* Payment 二级菜单 - 桌面端悬停，移动端点击 */}
							<div className="relative mr-5 group">
								{/* 桌面端悬停触发 */}
								<button 
									className="hidden md:flex items-center hover:text-white"
									onClick={togglePaymentMenu}
								>
									Payment
									<ChevronIcon isOpen={false} />
								</button>
								
								{/* 移动端点击触发 */}
								<button 
									className="flex md:hidden items-center hover:text-white"
									onClick={togglePaymentMenu}
								>
									Payment
									<ChevronIcon isOpen={isPaymentOpen} />
								</button>
								
								{/* 桌面端下拉菜单 - 悬停显示 */}
								<div className="hidden md:block absolute left-0 mt-2 w-48 bg-gray-800 rounded-md shadow-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 z-50">
									<div className="py-1">
										<a href="/paymentrecords" className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 hover:text-white">
											Payment Records
										</a>
										<a href="/paymentstatistics" className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 hover:text-white">
											Payment Statistics
										</a>
									</div>
								</div>
							</div>
							
							{/* 移动端 Payment 折叠菜单 */}
							<div className="block md:hidden w-full">
								{isPaymentOpen && (
									<div className="ml-4 mt-2 space-y-1">
										<a href="/paymentrecords" className="block text-sm text-gray-300 hover:text-white py-1">
											Payment Records
										</a>
										<a href="/paymentstatistics" className="block text-sm text-gray-300 hover:text-white py-1">
											Payment Statistics
										</a>
									</div>
								)}
							</div>
						</>
					)}
					
					<a className="mr-5 hover:text-white" href="/mypanel">My Panel</a>
					
					{/* Clients 二级菜单 - 桌面端悬停，移动端点击 */}
					<div className="relative mr-5 group">
						{/* 桌面端悬停触发 */}
						<button 
							className="hidden md:flex items-center hover:text-white"
							onClick={toggleClientsMenu}
						>
							Clients
							<ChevronIcon isOpen={false} />
						</button>
						
						{/* 移动端点击触发 */}
						<button 
							className="flex md:hidden items-center hover:text-white"
							onClick={toggleClientsMenu}
						>
							Clients
							<ChevronIcon isOpen={isClientsOpen} />
						</button>
						
						{/* 桌面端下拉菜单 - 悬停显示 */}
						<div className="hidden md:block absolute left-0 mt-2 w-48 bg-gray-800 rounded-md shadow-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 z-50">
							<div className="py-1">
								<a href="/macos" className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 hover:text-white">
									MacOS
								</a>
								<a href="/windows" className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 hover:text-white">
									Windows
								</a>
								<a href="/iphone" className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 hover:text-white">
									iPhone/iPad
								</a>
								<a href="/android" className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 hover:text-white">
									Android
								</a>
							</div>
						</div>
					</div>
					
					{/* 移动端 Clients 折叠菜单 */}
					<div className="block md:hidden w-full">
						{isClientsOpen && (
							<div className="ml-4 mt-2 space-y-1">
								<a href="/macos" className="block text-sm text-gray-300 hover:text-white py-1">
									MacOS
								</a>
								<a href="/windows" className="block text-sm text-gray-300 hover:text-white py-1">
									Windows
								</a>
								<a href="/iphone" className="block text-sm text-gray-300 hover:text-white py-1">
									iPhone/iPad
								</a>
								<a href="/android" className="block text-sm text-gray-300 hover:text-white py-1">
									Android
								</a>
							</div>
						)}
					</div>
				</nav>
				<span className="hover:text-white" href="#">Signed in as: <b>{loginState.jwt.email}</b></span>
				<button
					className="w-full sm:w-auto block text-white-900 bg-white hover:bg-gray-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-1.5 py-1 m-1 text-center dark:bg-gray-600 dark:hover:bg-gray-800 dark:focus:ring-gray-800"
					onClick={handleLogout}
				>
					<svg fill="none" className="inline-block h-4 w-4 mr-1" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24">
						<path d="M5 12h14M12 5l7 7-7 7"></path>
					</svg>
					logout
				</button>
			</div>
		</header>
	);
};

export default Menu;