import { useSelector } from "react-redux";
import TapToCopied from "./tapToCopied";

function WindowsV2rayN() {
	const loginState = useSelector((state) => state.login);

	return (
		<div className="xl:container xl:mx-auto px-5 xl:px-20">
			<h1 className="text-3xl font-bold mb-6 text-gray-800 dark:text-white">
				Windows 安装 v2rayN 指南
			</h1>

			<ol className="list-decimal list-inside space-y-6 text-gray-700 dark:text-gray-300">
				<li className="pb-4 border-b border-gray-200 dark:border-gray-700">
					<h2 className="inline-block text-xl font-semibold mb-2">下载客户端 v2rayN</h2>
					<p className="ml-6 mt-2">
						下载链接：
						<a
							href="https://lllinter.oss-cn-hangzhou.aliyuncs.com/geo/v2rayN-windows-64-desktop.zip"
							download
							target="_blank"
							rel="noopener noreferrer"
							className="inline-block px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors duration-200"
						>
							下载客户端 v2rayN
						</a>
	
					</p>
				</li>

				<li className="pb-4 border-b border-gray-200 dark:border-gray-700">
					<h2 className="inline-block text-xl font-semibold mb-2">解压并启动</h2>
					<p className="ml-6 mt-2">
						1. 将下载的 ZIP 文件解压到任意文件夹<br />
						2. 双击运行 <code className="bg-gray-100 dark:bg-gray-800 px-1 rounded">v2rayN.exe</code><br />
						3. 程序启动后会在系统托盘显示图标
					</p>
				</li>

				<li className="pb-4 border-b border-gray-200 dark:border-gray-700">
					<h2 className="inline-block text-xl font-semibold mb-2">添加订阅</h2>
					<p className="ml-6 mt-2">
						1. 点击顶部菜单"订阅分组"→"订阅分组设置"<br />
						2. 点击"添加"按钮<br />
						3. 在 URL 输入框中粘贴以下地址：<br />
						<TapToCopied>{process.env.REACT_APP_FILE_AND_SUB_URL + "/static/" + loginState.jwt.email}</TapToCopied>
					</p>
				</li>

				<li className="pb-4 border-b border-gray-200 dark:border-gray-700">
					<h2 className="inline-block text-xl font-semibold mb-2">更新订阅节点</h2>
					<p className="ml-6 mt-2">
						1. 点击顶部菜单"订阅分组"→"更新全部订阅（不通过代理）"<br />
						2. 等待节点列表刷新完成<br />
						3. 在节点列表中选择一个节点，右键点击→"设为活动服务器"
					</p>
				</li>

				<li className="pb-4 border-b border-gray-200 dark:border-gray-700">
					<h2 className="inline-block text-xl font-semibold mb-2">启用系统代理</h2>
					<p className="ml-6 mt-2">
						1. 点击底部状态栏"系统代理"→选择"自动配置系统代理"<br />
						2. 确认左下角代理状态显示为已启用
					</p>
				</li>

				<li>
					<h2 className="inline-block text-xl font-semibold mb-2">验证安装</h2>
					<p className="ml-6 mt-2">
						1. 打开浏览器，输入 <TapToCopied>https://www.google.com</TapToCopied><br />
						2. 如果能够成功打开，则表示安装和配置已完成
					</p>
				</li>
			</ol>

			<div className="mt-8 p-4 bg-blue-100 dark:bg-blue-900 rounded-md">
				<p className="text-blue-800 dark:text-blue-200 font-semibold">
					提示：如遇杀毒软件拦截，请将 v2rayN 所在文件夹加入白名单，或以管理员身份运行程序。
				</p>
			</div>
		</div>
	);
}

export default WindowsV2rayN;