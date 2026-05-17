import { useSelector } from "react-redux";
import TapToCopied from "./tapToCopied";

function Macos() {
	const loginState = useSelector((state) => state.login);


	return (
		<div className="xl:container xl:mx-auto px-5 xl:px-20">
			<h1 className="text-3xl font-bold mb-6 text-gray-800 dark:text-white">
				Mac OS 安装 Sing-box
			</h1>

			<p className="mb-6 text-red-600 dark:text-red-400 font-semibold">
				注意： 只有 MAC OSX 13 以上的系统，才能使用 Sing-box 客户端; MAC OSX 13 以下的系统，只能使用 Clash Verge 客户端。
			</p>

			<h2 className="text-2xl font-semibold mb-4 text-gray-700 dark:text-gray-300">
				安装 Sing-box 客户端
			</h2>

			<div className="bg-gray-100 dark:bg-gray-700 p-4 rounded-md mb-6">
				<p className="text-sm text-gray-600 dark:text-gray-400 mb-2">Apple ID 登录信息：</p>
				<p className="mb-1">
					Apple ID: <TapToCopied>warley8013@gmail.com</TapToCopied>
				</p>
				<p className="mb-2">
					Password: <TapToCopied>Google#2020</TapToCopied>
				</p>
				<p className="text-xs text-red-500 dark:text-red-400">
					注意: Apple ID 登陆时需要双重认证! 给我留言, 我收到验证码会发你。
				</p>
			</div>

			<ol className="list-decimal list-inside space-y-4 mb-6">
				<li>打开 App Store，点击左上角“商店”，点击“退出”，退出你当前账号！</li>
				<li>再次点击左上角“商店”，点击“登录”，输入 Apple ID 邮箱和密码，点击“下一步”，输入两步验证码。</li>
				<li>打开App Store，搜索 "sing-box VT" 并安装。</li>
				<li>安装后，打开软件，点击 "Install Network Extensions"，然后点击 "Allow"。</li>
				<li>
					添加配置：
					<ul className="list-disc list-inside ml-4 mt-2 space-y-2">
						<li>点击 "Profiles"</li>
						<li>
							填入以下参数：
							<div className="bg-gray-100 dark:bg-gray-700 p-3 rounded-md mt-2">
								<p>Name: w8</p>
								<p>
									URL: <TapToCopied>{process.env.REACT_APP_FILE_AND_SUB_URL + "/singbox/" + loginState.jwt.email}</TapToCopied>
								</p>
								<p>Auto Update: ON</p>
								<p>Auto Update Interval: 360</p>
							</div>
						</li>
						<li>点击 "Save" 保存配置</li>
					</ul>
				</li>
				<li>点击 "Dashboard" 标签页，打开 "Enable" 开关以启用 VPN。</li>
			</ol>

			<p className="text-gray-600 dark:text-gray-400">
				打开浏览器，访问 <TapToCopied>https://www.google.com</TapToCopied> 以验证配置是否成功。
			</p>
		</div>



	);
}

export default Macos;
