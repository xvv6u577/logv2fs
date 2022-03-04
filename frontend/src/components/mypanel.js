import { useEffect, useState } from "react";
import { Container, Table } from "react-bootstrap";
import { alert } from "../store/message";
import { useSelector, useDispatch } from "react-redux";
import { formatBytes } from "../service/service";
import axios from "axios";
import TapToCopied from "./tapToCopied";

function Mypanel() {
	const [user, setUser] = useState({});
	const dispatch = useDispatch();
	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);
	const rerenderSignal = useSelector((state) => state.rerender);

	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(alert({ show: false }));
			}, 5000);
		}
	}, [message, dispatch]);

	useEffect(() => {
		axios
			.get(process.env.REACT_APP_API_HOST + "user/" + loginState.jwt.Email, {
				headers: { token: loginState.token },
			})
			.then((response) => {
				setUser(response.data);
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	}, [rerenderSignal, loginState.jwt.Email, loginState.token, dispatch]);

	return (
		<Container className="py-3">
			<div className="row mypanel-row justify-content-evenly">
				<div className="mypanel-card col">
					<div className="h3">
						{user.used_by_current_day &&
							formatBytes(user.used_by_current_day.amount)}
					</div>
					<p>
						今日已用流量 (
						{user.used_by_current_day && user.used_by_current_day.period})
					</p>
				</div>
				<div className="mypanel-card col">
					<div className="h3">
						{user.used_by_current_month &&
							formatBytes(user.used_by_current_month.amount)}
					</div>
					<p>
						本月已用流量 (
						{user.used_by_current_month && user.used_by_current_month.period})
					</p>
				</div>
				<div className="mypanel-card col">
					<div className="h3">{user && formatBytes(user.used)}</div>
					<p>已用总流量</p>
				</div>
			</div>

			<div className="my-5 px-5 h6 small">
				<div className="my-1">
					用户名: <TapToCopied>{user.email}</TapToCopied>
				</div>
				<div className="my-1">
					密码:{" "}
					<TapToCopied>
						{" "}
						{user.email && user.email.length < 6 ? "mypassword" : user.email}{" "}
					</TapToCopied>
				</div>
				<div className="my-1">
					path: <TapToCopied>{user.path}</TapToCopied>
				</div>
				<div className="my-1">
					uuid: <TapToCopied>{user.uuid}</TapToCopied>
				</div>
				<div className="text-break my-1">
					SubUrl:{" "}
					<TapToCopied>
						{process.env.REACT_APP_FILE_AND_SUB_URL + "/static/" + user.email}
					</TapToCopied>
				</div>
			</div>

			<div className="d-flex flex-column">
				<div className="pb-3">
					<div className="h4 py-3">
						每月流量<span className="h6">(月份/流量)</span>
					</div>
					<Table hover bordered size="sm" variant="dark">
						<tbody>
							{user.traffic_by_month &&
								user.traffic_by_month
									.sort((a, b) => b.period - a.period)
									.map((item, index) => {
										return (
											<tr key={item.id}>
												<td>{index + 1}</td>
												{Object.values(item).map((val, i) => (
													<td>{i > 0 ? formatBytes(val) : val}</td>
												))}
											</tr>
										);
									})}
						</tbody>
					</Table>
				</div>
				<div className="">
					<div className="h4 pb-3">
						过去3个月每日流量<span className="h6">(日期/流量)</span>
					</div>
					<Table hover bordered size="sm" variant="dark">
						<tbody>
							{user.traffic_by_day &&
								user.traffic_by_day
									.sort((a, b) => b.period - a.period)
									.slice(0, 90)
									.map((item, index) => {
										return (
											<tr key={item.id}>
												<td>{index + 1}</td>
												{Object.values(item).map((val, i) => (
													<td>{i > 0 ? formatBytes(val) : val}</td>
												))}
											</tr>
										);
									})}
						</tbody>
					</Table>
				</div>
			</div>
		</Container>
	);
}

export default Mypanel;
