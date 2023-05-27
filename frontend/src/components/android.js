import { useSelector } from "react-redux";
import TapToCopied from "./tapToCopied";
import AndroidImages1 from "../images/android-1.png";
import AndroidImages2 from "../images/android-2.png";
import AndroidImages3 from "../images/android-3.png";
import AndroidImages4 from "../images/android-4.png";
import AndroidImages5 from "../images/android-5.jpeg";
import MyLightbox from "./MyLightbox";

function Android() {

	const loginState = useSelector((state) => state.login);
	const img1 = [AndroidImages1];
	const img2 = [AndroidImages2];
	const img3 = [AndroidImages3];
	const img4 = [AndroidImages4];
	const img5 = [AndroidImages5];


	return (
		<div className="xl:container xl:mx-auto px-5 xl:px-20">
			<h1 className="my-4 px-auto text-4xl font-extrabold tracking-tight leading-none text-gray-900 md:text-5xl lg:text-6xl dark:text-white">
				Android 系统中安装 clash
			</h1>
			<p className="my-6 text-base font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 1: 安装 clash 客户端</p>
			<div className="my-6 text-base font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				客户端下载:{" "}<br /> <TapToCopied>{process.env.REACT_APP_FILE_AND_SUB_URL + "/dl/ClashForAndroid-2.5.11-premium.apk"}</TapToCopied>
			</div>
			<p className="my-6 text-base font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				下载之后，运行安装程序，并授权安装。
			</p>
			<p className="my-6 text-base font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 2: 添加配置</p>
			<MyLightbox images={img1} device="mobile" />
			<p className="my-6 text-base font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				打开 clash 客户端，点击“配置”按钮，然后点击右上角“+”按钮，选择“从 URL 导入”
			</p>
			<MyLightbox images={img2} device="mobile" />
			<MyLightbox images={img3} device="mobile" />
			<p className="my-6 text-base font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				复制下面的 URL, 在 URL 下面粘贴:<br />
			</p>
			<p className="my-6 text-base font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				<TapToCopied >{process.env.REACT_APP_FILE_AND_SUB_URL + "/clash/" + loginState.jwt.Email + ".yaml"}</TapToCopied>
			</p>
			<p className="my-6 text-base font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				名称填写: clash, 自动更新填写(半天): 720
			</p>
			<p className="my-6 text-base font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				点按右上角图标，保存配置 <br />
			</p>
			<MyLightbox images={img4} device="mobile" />
			<p className="my-6 text-base font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 3: 日常运行设置</p>
			<div className="my-6 text-base font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				<p>回到 clash 客户端主界面，点按"已停止", 若文字变成"运行中"，则程序已打开</p>
			</div>
			<MyLightbox images={img5} device="mobile" />
			<div className="my-6 text-base font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				打开浏览器，访问 <TapToCopied>https://www.google.com</TapToCopied>，如果能正常访问，说明配置成功。
			</div>
		</div>
	);
}

export default Android;
