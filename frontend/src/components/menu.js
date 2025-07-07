import { useSelector, useDispatch } from "react-redux";
import { useState } from "react";
import { logout } from "../store/login";

const Menu = () => {
	const loginState = useSelector((state) => state.login);
	const dispatch = useDispatch();
	
	// 移动端折叠菜单状态管理
	const [isPaymentOpen, setIsPaymentOpen] = useState(false);
	const [isClientsOpen, setIsClientsOpen] = useState(false);

	const handleLogout = (e) => {
		dispatch(logout());
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
				<a href="/home" className="flex title-font font-medium items-center text-white mb-4 md:mb-0">
					<img alt="logo" src="data:image/png;base64,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"></img>
					<span className="ml-3 text-xl">Logv2fs Frontend</span>
				</a>
				<nav className="md:mr-auto md:ml-4 md:py-1 md:pl-4 md:border-l md:border-gray-700 flex flex-wrap items-center text-base justify-center">
					{loginState.jwt.Role === "admin" && (
						<>
							<a className="mr-5 hover:text-white" href="/home">User</a>
							<a className="mr-5 hover:text-white" href="/nodes">Nodes</a>
							<a className="mr-5 hover:text-white" href="/addnode">Add Node</a>
							
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
				<span className="hover:text-white" href="#">Signed in as: <b>{loginState.jwt.Email}</b></span>
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