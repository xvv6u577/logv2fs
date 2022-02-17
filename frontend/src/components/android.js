import { Container, Badge } from "react-bootstrap";
import { useSelector } from "react-redux";

function Android() {
	const loginState = useSelector((state) => state.login);
	return (
		<Container className="content">
			<h1 className="py-3">Android 客户端: </h1>
			<h3 className="py-2">step 1: 安装v2rayNG</h3>
			<p>
				客户端下载:{" "}
				<div className="inline h4">
					<a href={process.env.REACT_APP_FILE_AND_SUB_URL +"/dl/"+ "v2rayNG_1.2.6.apk"}>
						{process.env.REACT_APP_FILE_AND_SUB_URL +"/dl/"+ "v2rayNG_1.2.6.apk"}
					</a>
				</div>
			</p>
			<h3 className="py-2">step 2: 添加配置参数</h3>
			<p>
				打开软件，进入主界面。点按左上角“☰” &#x2192; 订单设置 &#x2192;
				右上角“+”，填入下面参数
			</p>
			<p>
				备注:
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						w8
					</Badge>
				</div>
			</p>
			<p>
				地址(url):
				<div className="inline h4">
					<Badge bg="secondary" pill className="mx-1">
						{process.env.REACT_APP_FILE_AND_SUB_URL + "/static/" + loginState.jwt.Email}
					</Badge>
				</div>
			</p>
			<p>点界面右上角的对勾，保存配置。</p>
			<h3 className="py-2">step 3: 运行App</h3>
			<p>回到主界面，点按右上角“⋮” &#x2192; 点按“更新配置”。</p>
			<p>
				再次回到主界面，点右下角的v2rayNG图标，启动程序。如果出现网络连接请求，点击确定。安装完成!
			</p>
		</Container>
	);
}

export default Android;
