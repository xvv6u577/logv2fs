import { useSelector } from "react-redux";
import TapToCopied from "./tapToCopied";
// import AndroidImages1 from "../images/android-1.png";
// import AndroidImages2 from "../images/android-2.png";
// import AndroidImages3 from "../images/android-3.png";
// import AndroidImages4 from "../images/android-4.png";
// import AndroidImages5 from "../images/android-5.jpeg";
// import MyLightbox from "./MyLightbox";

function Android() {

	const loginState = useSelector((state) => state.login);
	// const img1 = [AndroidImages1];
	// const img2 = [AndroidImages2];
	// const img3 = [AndroidImages3];
	// const img4 = [AndroidImages4];
	// const img5 = [AndroidImages5];


	return (
		<div className="xl:container xl:mx-auto mx-auto px-4 py-8 max-w-3xl">
			<h1 class="text-3xl font-bold text-center text-gray-800 dark:text-white mb-8">Android 安装 Sing-box 指南</h1>

			<ol class="space-y-8">
				<li>
					<h2 class="text-xl font-semibold text-gray-800 dark:text-white mb-2">下载客户端 Sing-box</h2>
					<p class="text-gray-700 dark:text-gray-300">下载链接：
						<a href="https://w8.undervineyard.com/dl/Singbox-for-Android-1.8.11-universal.apk" class="text-blue-500 hover:text-blue-600 underline">
							https://w8.undervineyard.com/dl/Singbox-for-Android-1.8.11-universal.apk
						</a>
					</p>
				</li>

				<li>
					<h2 class="text-xl font-semibold text-gray-800 dark:text-white mb-2">添加配置文件</h2>
					<ol class="list-decimal pl-6 space-y-2 text-gray-700 dark:text-gray-300">
						<li>在客户端主界面，点按"profiles"</li>
						<li>选择 "New Profile"</li>
						<li>填写以下信息：
							<ul class="list-disc pl-6 mt-2 space-y-1">
								<li>Name：w8</li>
								<li>Type: remote</li>
								<li>URL：<TapToCopied>{process.env.REACT_APP_FILE_AND_SUB_URL + "/singbox/" + loginState.jwt.Email}</TapToCopied></li>
								<li>Update Interval: 360</li>
							</ul>
						</li>
					</ol>
				</li>

				<li>
					<h2 class="text-xl font-semibold text-gray-800 dark:text-white mb-2">启用服务</h2>
					<ol class="list-decimal pl-6 space-y-2 text-gray-700 dark:text-gray-300">
						<li>在客户端主界面，点按"Dashboard"</li>
						<li>打开 Enable 开关按钮</li>
					</ol>
				</li>

				<li>
					<h2 class="text-xl font-semibold text-gray-800 dark:text-white mb-2">验证安装</h2>
					<ol class="list-decimal pl-6 space-y-2 text-gray-700 dark:text-gray-300">
						<li>打开您的浏览器</li>
						<li>尝试访问 <TapToCopied>www.google.com</TapToCopied></li>
						<li>如果能够成功打开，则表示安装和配置已完成</li>
					</ol>
				</li>
			</ol >

		<div class="bg-yellow-100 dark:bg-yellow-900 border-l-4 border-yellow-500 text-yellow-700 dark:text-yellow-200 p-4 mt-8">
			<p class="font-semibold">提示：</p>
			<p>如果在安装过程中遇到任何问题，请确保您的Android设备允许安装来自未知来源的应用，并检查网络连接是否正常。</p>
		</div>
		</div >
	);
}

export default Android;
