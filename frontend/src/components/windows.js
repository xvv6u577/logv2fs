import { useSelector } from "react-redux";
import TapToCopied from "./tapToCopied";
import MyLightbox from "./MyLightbox";
import WindowsImages1 from "../images/windows-1.png";
import WindowsImages2 from "../images/windows-2.png";
import WindowsImages3 from "../images/windows-3.png";
import WindowsImages4 from "../images/windows-4.png";
import WindowsImages5 from "../images/windows-5.png";


function Windows() {
	const loginState = useSelector((state) => state.login);
	const img1 = [WindowsImages1];
	const img2 = [WindowsImages2];
	const img3 = [WindowsImages3];
	const img4 = [WindowsImages4];
	const img5 = [WindowsImages5];

	return (
		<div className="xl:container xl:mx-auto px-5 xl:px-20">
			<h1 className="my-4 px-auto text-4xl font-extrabold tracking-tight leading-none text-gray-900 md:text-5xl lg:text-6xl dark:text-white">
				Windows 系统中安装 clash
			</h1>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 1: 安装 clash 客户端</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				客户端下载:{" "} <TapToCopied>{process.env.REACT_APP_FILE_AND_SUB_URL + "/dl/Clash.for.Windows.Setup.0.19.26.exe"}</TapToCopied><br />
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				下载之后，双击安装包，运行安装程序，并授权 clash 接入网络。若标题栏出现小黑猫图标, 说明 clash 已运行。
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 2: 添加配置</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				双击下方标题栏 clash 图标, 选择 General 页，<br /></p>
			<MyLightbox device={'desktop'} images={img1} />
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 2-1: 点击箭头1指向处</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				点按 Service Mode 右边的 Manage, 弹出的对话框！如果 Current Status 是 Inactive, 点按 Install 按钮。
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				服务安装成功后，程序会自动重启！
			</p>
			<MyLightbox device={'desktop'} images={img2} />
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 2-2: 点击箭头2指向处</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				再次回到 General , Service Mode 右边的图标会变成绿色, Current Status 是 Active, 说明服务启动成功。
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				在 General 页面, 点击TUN Mode 右边的小齿轮，弹出的对话框，不用更改任何设置，点“保存”。
			</p>

			<MyLightbox device={'desktop'} images={img5} />
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 2-3: 添加URL</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				双击下方标题栏 clash 图标, 选择 Profiles 页，复制下面的 Url, 
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				<TapToCopied>{process.env.REACT_APP_FILE_AND_SUB_URL + "/clash/" + loginState.jwt.Email + ".yaml"}</TapToCopied>
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				粘贴到提示信息"Download from URL" 的输入框中，点按 Download 按钮，添加配置。
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				添加成功后, Profiles 页面有新的条目出现。
			</p>
			<div className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
			</div>
			<MyLightbox device={'desktop'} images={img3} />
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 3: 日常运行设置</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				右击下方标题栏 clash 图标, Proxy Mode, 选择 Rule!
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				选中 TUN Mode!
			</p>
			<MyLightbox device={'desktop'} images={img4} />
			<div className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				打开浏览器，访问 <TapToCopied>https://www.google.com</TapToCopied>，如果能正常访问，说明配置成功。
			</div>
		</div>
	);
}

export default Windows;
