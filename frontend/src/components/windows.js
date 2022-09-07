import { useSelector } from "react-redux";
import TapToCopied from "./tapToCopied";

function Windows() {
	const loginState = useSelector((state) => state.login);

	return (
		<div className="md:container md:mx-auto px-20">
			<h1 className="my-4 px-auto text-4xl font-extrabold tracking-tight leading-none text-gray-900 md:text-5xl lg:text-6xl dark:text-white">
				Windows 系统中安装 clash
			</h1>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 1: 安装 clash 客户端</p>
			<div className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				客户端下载:{" "}<br /> <TapToCopied>{process.env.REACT_APP_FILE_AND_SUB_URL + "/dl/Clash.for.Windows.Setup.0.19.26.exe"}</TapToCopied>
			</div>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				下载之后，双击安装包，运行安装程序，并授权 clash 接入网络。若标题栏出现小黑猫图标, 说明 clash 已运行。
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 2: 添加配置</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				双击下方标题栏 clash 图标, 选择 General 页，<br />
				点按 Service Mode 右边的 Manage, 点按 Install 按钮，安装服务。<br />
				安装成功后，程序会自动重启！<br />
				再次回到 General 页, Service Mode 右边的图标会变成绿色，说明服务已启动。<br />
			</p>
			<div className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-12 dark:text-sky-200">
				双击下方标题栏 clash 图标, 选择 Profiles 页，复制下面的 Url, <br />
				<p className="py-5"><TapToCopied>{process.env.REACT_APP_FILE_AND_SUB_URL + "/clash/" + loginState.jwt.Email+".yaml"}</TapToCopied></p>
				粘贴到提示信息"Download from URL" 的输入框中，点按 Download 按钮，添加配置。<br />
				添加成功后, Profiles 页面有新的条目出现。<br />
			</div>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 3: 日常运行设置</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				右击下方标题栏 clash 图标, Proxy Mode, 选择 Rule<br />
				选中 TUN Mode<br />
				选中 System Proxy<br />
			</p>
			<div className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				打开浏览器，访问 <TapToCopied>https://www.google.com</TapToCopied>，如果能正常访问，说明配置成功。
			</div>
		</div>
	);
}

export default Windows;
