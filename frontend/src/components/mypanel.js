import { useEffect, useState } from "react";
import { Container } from "react-bootstrap";
import { alert } from "../store/message";
import { useSelector, useDispatch } from "react-redux";
import { formatBytes } from "../service/service";
import axios from "axios";
import TapToCopied from "./tapToCopied";
import TrafficTable from "./trafficTable";

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
	}, [rerenderSignal, loginState, dispatch]);

	return (
		<Container className="py-3">
			<h3>#6 Github Actions Test</h3>
			<div className="row mypanel-row justify-content-evenly">
				<div className="mypanel-card col">
					<div className="h3">
						{user.used_by_current_day &&
							formatBytes(user.used_by_current_day.amount)}
					</div>
					<p>
						Traffic Used Today (
						{user.used_by_current_day && user.used_by_current_day.period})
					</p>
				</div>
				<div className="mypanel-card col">
					<div className="h3">
						{user.used_by_current_month &&
							formatBytes(user.used_by_current_month.amount)}
					</div>
					<p>
						Traffic Used This Month (
						{user.used_by_current_month && user.used_by_current_month.period})
					</p>
				</div>
				<div className="mypanel-card col">
					<div className="h3">{user && formatBytes(user.used)}</div>
					<p>Traffic Used In Total</p>
				</div>
			</div>

			<div
				className="d-md-flex flex-column justify-content-between border border-info border-3 rounded-3  mx-auto my-3 p-3"
				style={{ width: "50%" }}
			>
				<div className="d-flex justify-content-between">
					<span className="">用户名:</span>{" "}
					<TapToCopied>{user.email}</TapToCopied>
				</div>
				<div className="d-flex justify-content-between">
					<span className="">path: </span>
					<TapToCopied>{user.path}</TapToCopied>
				</div>
				<div className="d-md-flex justify-content-between">
					<span className="">uuid: </span>
					<TapToCopied>{user.uuid}</TapToCopied>
				</div>
				<div className="d-md-flex justify-content-between">
					<span className="">SubUrl:</span>
					<TapToCopied>
						{process.env.REACT_APP_FILE_AND_SUB_URL + "/static/" + user.email}
					</TapToCopied>
				</div>
			</div>

			<row className="">
				<div className="pb-3 d-flex flex-column">
					<div className="h5 py-3 text-center">
						Monthly Traffic in the Past 1 Year
					</div>
					<TrafficTable data={user.traffic_by_month} limit={12} by="月份" />
				</div>
				<div className="d-flex flex-column">
					<div className="h5 pb-3 text-center">
						Daily Traffic in the Past 3 Month
					</div>
					<TrafficTable data={user.traffic_by_day} limit={90} by="日期" />
				</div>
			</row>
		</Container>
	);
}

export default Mypanel;
