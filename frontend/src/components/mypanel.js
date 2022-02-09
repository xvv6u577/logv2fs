import { useEffect, useState } from "react";
import { Container, Badge } from "react-bootstrap";
import { alert } from "../store/message";
import { useSelector, useDispatch } from "react-redux";
import { formatBytes } from "../service/service";
import axios from "axios";

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
		<Container className="py-5">
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

			<div className="d-flex flex-column py-3">
				<div className="py-3">
					<div className="h4 py-3">
						每月流量<span className="h6">(月份/流量)</span>
					</div>
					{user.traffic_by_month &&
						user.traffic_by_month
							.sort((a, b) => b.period - a.period)
							.map((element) => {
								return (
									<Badge pill bg="dark" text="white">
										{element.period} / {formatBytes(element.amount)}
									</Badge>
								);
							})}
				</div>
				<div className="py-3">
					<div className="h4 py-3">
						过去3个月每日流量<span className="h6">(日期/流量)</span>
					</div>
					{user.traffic_by_day &&
						user.traffic_by_day
							.sort((a, b) => b.period - a.period)
							.slice(0, 90)
							.map((element) => {
								return (
									<Badge pill bg="dark" text="white">
										{element.period} / {formatBytes(element.amount)}
									</Badge>
								);
							})}
				</div>
			</div>
		</Container>
	);
}

export default Mypanel;
