import { Container, Badge } from "react-bootstrap";
import { useSelector } from "react-redux";

function Ihpone() {
	const loginState = useSelector((state) => state.login);

	return (
		<Container className="content">
			<h1 className="py-3">iphone/ipad 客户端:</h1>
			<h3 className="py-2">step 1、安装 "shadowrocket"。</h3>
			<p>
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						设置
					</Badge>
				</div>
				&#x2192;
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						Apple ID、iCloud、媒体与购买项目
					</Badge>
				</div>
				&#x2192;
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						媒体与购买项目
					</Badge>
				</div>
				，退出你的Apple ID，用下面的 Apple ID（美区）登入。
			</p>
			<p>
				apple ID:
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						warley8013@gmail.com
					</Badge>
				</div>
			</p>
			<p>
				pwd:
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						Google@521
					</Badge>
				</div>
			</p>
			<p>打开app store，查找 "shadowrocket" ，安装。</p>
			提示:
			<ul>
				<li>上面提供的apple id已经购买shadowrocket，登陆之后可以直接安装。</li>
				<li>
					apple ID
					登陆的时候，需要双重认证，给我发信息提示，我会发给你认证数字。
				</li>
			</ul>
			<h3 className="py-2">step 2、导入配置参数</h3>
			<p>
				点按shadowrocket右上角“+” &#x2192; 类型选择“Subscribe” &#x2192;
				填入以下配置
			</p>
			<p>
				备注:
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						uvp
					</Badge>
				</div>
			</p>
			<p>
				URL:{" "}
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						{process.env.REACT_APP_SUBURL + "/static/" + loginState.jwt.Email}
					</Badge>
				</div>
			</p>
			<p>点按右上角“完成”。</p>
			<h3 className="py-2">Step 3: 设置运行模式</h3>
			<p>
				回到主界面，向右滑动新添加的选项uvp，点按“更新”。项目下方会出现2个列表，选中其中一个后，左边会现黄色圆点。
			</p>
			<p>回到主界面，"全局路由"选择"配置"。</p>
			<p>
				打开"未连接"右边单选框，若标题栏出现“VPN”字样，说明 App 已经在运行。
			</p>
			<p>此时，shadowrocket 接管手机的网络连接，手机中App已处于可番茄状态。</p>
		</Container>
	);
}

export default Ihpone;
