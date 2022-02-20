import { useSelector } from "react-redux";
import { Container, } from "react-bootstrap";
import TapToCopied from "./tapToCopied";

function Macos() {
	const loginState = useSelector((state) => state.login);

	return (
		<Container className="py-3 content">
			<h1 className="py-3">Mac系统客户端</h1>

			<h3 className="py-2">
				<p>Step 1: 安装 V2ray 客户端</p>
			</h3>
			<p>
				客户端下载:{" "}
				<TapToCopied>{process.env.REACT_APP_FILE_AND_SUB_URL + "/dl/v2rayx.zip"}</TapToCopied>
			</p>
			<p>
				解压之后, 把 v2rayx 移到/Application 文件夹。四指捏合, 调出
				launchpad, 点按 v2rayx图标。若标题栏出现钻石图标, 说明 v2ray 已运行。
			</p>
			<h3 className="py-2">step 2:添加配置</h3>
			<p>
				点按标题栏 App 图标, 依次选择 Configure... &#x2192; Import... &#x2192;
				Import from other links... &#x2192; 输入
			</p>
			<p>
				<TapToCopied>{process.env.REACT_APP_FILE_AND_SUB_URL + "/static/" + loginState.jwt.Email}</TapToCopied>
			</p>
			<p>点按 OK。若添加成功, 配置对话框左侧vmess servers有新项目产生。</p>
			<h3 className="py-2">step 3: 运行 v2ray 客户端</h3>
			<p>
				点击App图标, 选择 Server... &#x2192; w8-hk-gcp 或
				rm-la-twitter, 选择其中一个!
			</p>
			<p>点击App图标, 选择 Routing Rule &#x2192; bypasscn_private_apple。</p>
			<p>点击App图标, 选择 “Global Mode”</p>
			<p>点击App图标, 选中 load core, 运行客户端。</p>

			<p>此时 safari/chrome/firefox 已经打开 Google 了。</p>
		</Container>
	);
}

export default Macos;
