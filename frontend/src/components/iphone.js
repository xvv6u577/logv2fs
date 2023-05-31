import { useSelector } from "react-redux";
import TapToCopied from "./tapToCopied";
import IphoneImages1 from "../images/iphone-1.jpeg";
import IphoneImages2 from "../images/iphone-2.jpeg";
import IphoneImages3 from "../images/iphone-3.jpeg";
import IphoneImages4 from "../images/iphone-4.jpeg";
import MyLightbox from "./MyLightbox";

function Ihpone() {
	const loginState = useSelector((state) => state.login);
	const img1 = [IphoneImages1];
	const img2 = [IphoneImages2];
	const img3 = [IphoneImages3];
	const img4 = [IphoneImages4];

	return (
		<div className="xl:container xl:mx-auto px-5 xl:px-20">
			<h1 className="my-4 px-auto text-4xl font-extrabold tracking-tight leading-none text-gray-900 md:text-5xl lg:text-6xl dark:text-white">
				iphone/ipad 中安装 Shadowrocket
			</h1>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 1: 安装 Shadowrocket 客户端</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				依次点按 settings -{'>'} Apple ID、iCloud、媒体与购买项目 -{'>'} 媒体与购买项目
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				退出你的Apple ID, 用下面的 Apple ID(美区)登入
			</p>
			<MyLightbox device={'mobile'} images={img1} /><br />
			<MyLightbox device={'mobile'} images={img2} />
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				<code>
					Apple ID: <TapToCopied>warley8013@gmail.com</TapToCopied><br />
					Password: <TapToCopied>Google#2020</TapToCopied><br />
				</code>
				<code className="text-sm py-5">
					Notice: <br />
					1. 上面apple id已经购买shadowrocket, 登陆后可直接安装。<br />
					2. apple ID 登陆时, 需要双重认证, 给我发信息, 我会发你认证数字。<br />
				</code>
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				打开app store, 搜索 "shadowrocket" , 安装！
			</p>

			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 2: 添加配置</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				打开 Shadowrocket, 点按shadowrocket右上角“+”号
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				类型选择“Subscribe”,URL 填入下面地址:
				<TapToCopied>{process.env.REACT_APP_FILE_AND_SUB_URL + "/static/" + loginState.jwt.Email}</TapToCopied>,
				备注"clash"!
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				点按右上角“完成”按钮。若配置添加成功, 回到 Shadowrocket 首页, 有新项目生成！
			</p>

			<MyLightbox device={'mobile'} images={img3} />
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 3: 日常运行设置</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				点按 Shadowrocket 左下角“首页” -{'>'} 点按“全局路由” -{'>'} 选择“配置”
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				点按 Shadowrocket 右下角“设置” -{'>'} 找到“订阅” -{'>'} 分别选中“打开时更新”，“自动后台更新”！
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				点按新生成项目左边箭头，展开列表，选中任何一个节点！被选中节点会成为当前在用节点，左端出现橙色小圆点
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				点按"未连接"右边单选框，打开 vpn 接入网络
			</p>
			<MyLightbox device={'mobile'} images={img4} />
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				打开浏览器，地址栏输入 <TapToCopied>https://www.google.com</TapToCopied>！如果能正常访问，说明配置成功。
			</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-xl sm:px-16 xl:px-10 dark:text-gray-400">Step 4: 退出公用 Apple ID</p>
			<p className="my-6 text-baseg font-normal text-gray-500 lg:text-base sm:px-16 xl:px-10 dark:text-sky-200">
				安装结束, 请从“媒体与购买项目”退出Apple ID! <br />
			依次点按 settings -> Apple ID、iCloud、媒体与购买项目 -> 媒体与购买项目，点“退出”
			</p>
		</div>
	);
}

export default Ihpone;
