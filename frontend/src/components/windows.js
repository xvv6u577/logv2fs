import { useSelector } from "react-redux";
import TapToCopied from  "./tapToCopied";
// import MyLightbox from "./MyLightbox";
// import WindowsImages1 from "../images/windows-1.png";
// import WindowsImages2 from "../images/windows-2.png";
// import WindowsImages3 from "../images/windows-3.png";
// import WindowsImages4 from "../images/windows-4.png";
// import WindowsImages5 from "../images/windows-5.png";


function Windows() {
	const loginState = useSelector((state) => state.login);
	// const img1 = [WindowsImages1];
	// const img2 = [WindowsImages2];
	// const img3 = [WindowsImages3];
	// const img4 = [WindowsImages4];
	// const img5 = [WindowsImages5];

	return (
		<div className="xl:container xl:mx-auto px-5 xl:px-20">
			<h1 className="text-3xl font-bold mb-6 text-gray-800 dark:text-white">
				Windows 安装 Clash Verge 指南
			</h1>

			<ol className="list-decimal list-inside space-y-6 text-gray-700 dark:text-gray-300">
				<li className="pb-4 border-b border-gray-200 dark:border-gray-700">
					<h2 className="inline-block text-xl font-semibold mb-2">下载客户端 Clash Verge</h2>
					<p className="ml-6 mt-2">
						下载链接：
						<TapToCopied>{process.env.REACT_APP_FILE_AND_SUB_URL + "/dl/Clash.Verge_1.7.3_x64-setup.exe"}</TapToCopied>
					</p>
				</li>

				<li className="pb-4 border-b border-gray-200 dark:border-gray-700">
					<h2 className="inline-block text-xl font-semibold mb-2">安装"服务模式"</h2>
					<p className="ml-6 mt-2">
						1. 点击左边栏"配置"<br />
						2. 点按"服务模式"右边的盾牌图标<br />
						3. 按照提示完成"服务模式"的安装
					</p>
				</li>

				<li className="pb-4 border-b border-gray-200 dark:border-gray-700">
					<h2 className="inline-block text-xl font-semibold mb-2">添加配置文件</h2>
					<p className="ml-6 mt-2">
						1. 点击左边栏"订阅"<br />
						2. 点击"添加"按钮<br />
						3. 在URL输入框中粘贴以下地址：<br />
						<TapToCopied>{process.env.REACT_APP_FILE_AND_SUB_URL + "/verge/" + loginState.jwt.Email}</TapToCopied>
					</p>
				</li>

				<li className="pb-4 border-b border-gray-200 dark:border-gray-700">
					<h2 className="inline-block text-xl font-semibold mb-2">启用服务</h2>
					<p className="ml-6 mt-2">
						1. 点击左边栏"配置"<br />
						2. 打开"服务模式"开关<br />
						3. 打开"系统代理"开关
					</p>
				</li>

				<li>
					<h2 className="inline-block text-xl font-semibold mb-2">验证安装</h2>
					<p className="ml-6 mt-2">
						1. 打开您的浏览器<br />
						2. 尝试访问 <TapToCopied>www.google.com</TapToCopied><br />
						3. 如果能够成功打开，则表示安装和配置已完成
					</p>
				</li>
			</ol>

			<div className="mt-8 p-4 bg-blue-100 dark:bg-blue-900 rounded-md">
				<p className="text-blue-800 dark:text-blue-200 font-semibold">
					提示：如果在安装过程中遇到任何问题，请确保您有管理员权限，并尝试以管理员身份运行 Clash Verge。
				</p>
			</div>
		</div>
	);
}

export default Windows;
