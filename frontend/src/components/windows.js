import { useEffect, useState } from "react";
import { Container, Alert, Badge } from "react-bootstrap";
import { useSelector, useDispatch } from "react-redux";
import axios from "axios";
import { alert, info, success } from "../store/message";

function Windows() {
	const [user, updateUser] = useState({});
	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);
	const dispatch = useDispatch();

	useEffect(() => {
		axios
			.get(process.env.REACT_APP_API_HOST + "user/" + loginState.jwt.Email, {
				headers: { token: loginState.token },
			})
			.then((response) => {
				updateUser(response.data);
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	}, [dispatch, loginState.jwt.Email, loginState.token]);

	return (
		<Container className="content py-3">
			<Alert show={message.show} variant={message.type}>
				{" "}
				{message.content}{" "}
			</Alert>
			<h1 className="py-3">Windows 客户端 (v2rayw)</h1>
			<h3 className="py-2">step 1: window 系统时间校准</h3>
			<p>
				请确保你的PC已联网! windows 的本地时间和标准时间相差必须在 90s
				以内（时区无关），否则客户端运行会有问题。
			</p>
			<p>
				右键点击右下角的“时间” &#x2192; 点击“调整日期时间” &#x2192;
				把自动设置时间和自动设置时区勾选上即可.
			</p>
			<h3 className="py-2">step 2: 下载客户端</h3>
			<p>
				下载客户端:{" "}
				<div className="inline h4">
					<a href={process.env.REACT_APP_SUBURL + "/dl/v2rayw.zip"}>
						{process.env.REACT_APP_SUBURL + "/dl/v2rayw.zip"}
					</a>
				</div>
			</p>
			<p>
				解压v2rayw.zip。打开解压后文件夹，双击运行v2rayw.exe文件。windows右下角状态栏出现绿色W形状图标。
			</p>
			<h3 className="py-2">step 3: 添加配置</h3>
			<p>
				右击绿色W形状图标 &#x2192; 配置...,
				在v2rayw配置面板中，依次填入下面参数:
			</p>
			<p>
				本地socks5端口:{" "}
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						1080
					</Badge>{" "}
				</div>
				本地http端口:{" "}
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						8008
					</Badge>{" "}
				</div>
			</p>
			<p> 点按“添加”，填入服务器信息&#x2192; </p>
			<p>
				{" "}
				地址:
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						{user.nodeinuse && user.nodeinuse.w8}:443
					</Badge>{" "}
				</div>
				用户ID:{" "}
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						{user.uuid}
					</Badge>{" "}
				</div>
				额外ID:{" "}
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						64
					</Badge>{" "}
				</div>
				等级:{" "}
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						0
					</Badge>{" "}
				</div>
				加密方式:{" "}
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						auto
					</Badge>
				</div>
				标签:{" "}
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						w8
					</Badge>{" "}
				</div>
				网络类型:{" "}
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						ws
					</Badge>
				</div>
			</p>
			<p>点按“传输设置”&#x2192;</p>
			<ul>
				<li>
					WebSocket标签:
					<p>
						路径:{" "}
						<div className="inline h4">
							<Badge bg="secondary" pill className="mx-1">
								/{user.path}
							</Badge>{" "}
						</div>
						http头部: （留空）
					</p>
				</li>
				<li>
					TLS标签:
					<p>选中"启用传输层加密TLS"; </p>
					<p> “域名服务器”，删除“server.cc”，留空白; </p>
				</li>
				<li>其它标签保持原配置，不用更改。</li>
			</ul>
			<p>点按“保存”。v2rayw配置面板中，点按保存。</p>
			<h3 className="py-2">step 4: 设置系统代理</h3>
			<p>windows系统设置中，搜索“proxy”。选中“使用代理服务器”</p>
			<p>
				地址:{" "}
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						127.0.0.1
					</Badge>
				</div>
				端口:{" "}
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						8008
					</Badge>
				</div>
				选中“请勿将代理服务器用于本地地址”。
			</p>
			<h3 className="py-2">step 5: 运行v2rayw</h3>
			<p>
				右击绿色W形状图标 &#x2192; v2ray内部路由规则 &#x2192;
				选择“绕过本地和CN地址”
			</p>
			<p>右击绿色W形状图标 &#x2192; 选择“加载v2ray”</p>
			<p>
				配置完成!
				打开edge/chrome浏览器，输入baidu.com，测试是否能正常联网；输入www.google.com，测试是否能正常联外网。
			</p>
			<p>
				（以上内容，遇到问题，请给我发信息，我们可以约时间通过zoom远程控制帮助你安装。）
			</p>
		</Container>
	);
}

export default Windows;
