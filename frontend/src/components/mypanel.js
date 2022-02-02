import { useEffect,useState } from "react";
import { Container, Card,Badge } from "react-bootstrap";
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
		<Container className="py-3">
			<Card className="user-page-card" bg="light">
				<Card.Header>{user && user.name}</Card.Header>
				<Card.Body>
					<Card.Title></Card.Title>
					<Card.Text>
						<div className="">
							<Badge bg="info" text="dark">
								今日: {user.used_by_current_day && user.used_by_current_day.period}
								已用流量:{" "}
								{user.used_by_current_day && formatBytes(user.used_by_current_day.amount)}
							</Badge>
						</div>

						<div className="">
							<Badge bg="info" text="dark">
								本月: {user.used_by_current_month && user.used_by_current_month.period}
								已用流量:{" "}
								{user.used_by_current_month && formatBytes(user.used_by_current_month.amount)}
							</Badge>
						</div>

						<div className="">
							<Badge bg="info" text="dark">
								已用总流量: {user && formatBytes(user.used)}
							</Badge>
						</div>
					</Card.Text>

					<h6>每月流量（月份/流量）:</h6>
					<div>
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
					<h6 className="pt-2">每日流量(日期/流量):</h6>
					<div>
						{user.traffic_by_day &&
							user.traffic_by_day
								.sort((a, b) => b.period - a.period)
								.map((element) => {
									return (
										<Badge pill bg="dark" text="white">
											{element.period} / {formatBytes(element.amount)}
										</Badge>
									);
								})}
					</div>
				</Card.Body>
			</Card>
		</Container>
	);
}

export default Mypanel;
