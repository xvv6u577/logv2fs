import { useSelector } from "react-redux";
import TapToCopied from "./tapToCopied";
import ClashxMac1 from "../images/clashx-mac-1.png";
import ClashxMac2 from "../images/clashx-mac-2.png";
import ClashxMac3 from "../images/clashx-mac-3.png";
import MyLightbox from "./MyLightbox";
import { IStore as RootState } from "../types";

function Macos() {
	const loginState = useSelector((state: RootState) => state.login);
	const img1 = [ClashxMac1];
	const img2 = [ClashxMac2];
	const img3 = [ClashxMac3];


	return (
		<div className="xl:container xl:mx-auto px-5 xl:px-20">
			<h1 className="my-4 px-auto text-4xl font-extrabold tracking-tight leading-none text-gray-900 md:text-5xl lg:text-6xl dark:text-white">
				Mac os 系统中安装 clashX
			</h1>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 1: 安装 clashX 客户端</p>
			<div className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				客户端下载:{" "}<br /> <TapToCopied>{process.env.REACT_APP_FILE_AND_SUB_URL + "/dl/ClashX-v1.95.0.1.dmg"}</TapToCopied>
			</div>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				把 clashX 移到/Application 文件夹。四指捏合, 调出 launchpad, 点按 clashX 图标。若标题栏出现小黑猫图标, 说明 clashX 已运行。
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 2: 添加配置</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				点按标题栏 clashX 图标, 依次选择 Config -{'>'} Remote config -{'>'} Manage -{'>'} Add
			</p>
			<MyLightbox  device={'desktop'}  images={img1} />
			<div className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-12 dark:text-sky-200">
				选择 URL, 在 URL 中输入:<br />
				<TapToCopied>{process.env.REACT_APP_FILE_AND_SUB_URL + "/clash/" + loginState.jwt.Email+".yaml"}</TapToCopied><br />
				Config Name 可以随意填写, 例如: clashX
			</div>
			<MyLightbox  device={'desktop'}  images={img2} />
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 3: 日常运行设置</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				Proxy Mode, 选择 Rule<br />
				选中 Enhanced Mode<br />
			</p>
			<MyLightbox  device={'desktop'}  images={img3} />
			<div className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				打开浏览器，访问 <TapToCopied>https://www.google.com</TapToCopied>，如果能正常访问，说明配置成功。
			</div>
		</div>
	);
}

export default Macos;