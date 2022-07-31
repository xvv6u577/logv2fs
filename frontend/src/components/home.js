import { useEffect, useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, reset } from "../store/message";
import axios from "axios";
import { logout } from "../store/login";
import UserComp from "./userComp";
import Alert from "./alert";

const Home = () => {
	const [users, setUsers] = useState([]);
	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);
	const rerenderSignal = useSelector((state) => state.rerender);

	const dispatch = useDispatch();

	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(reset({}));
			}, 5000);
		}
	}, [message, dispatch]);

	useEffect(() => {
		axios
			.get(process.env.REACT_APP_API_HOST + "alluser", {
				headers: { token: loginState.token },
			})
			.then((response) => {
				let user = response.data.filter(
					(ele) => ele.email === loginState.jwt.Email
				);
				if (user.length !== 0) {
					setUsers(response.data);
				} else {
					dispatch(logout());
				}
			})
			.catch((err) => {
				if (err.response) {
					dispatch(alert({ show: true, content: err.response.data.error }));
				} else {
					dispatch(alert({ show: true, content: err.toString() }));
				}
			});
	}, [rerenderSignal, loginState.jwt.Email, loginState.token, dispatch]);

	return (
		<div className="my-3">
			<Alert message={message.content} type={message.type} shown={message.show} close={() => { dispatch(reset({})); }} />

			<div id="accordion-collapse" data-accordion="collapse">
				{users
					// put admin at the top of the list
					.reduce((acc, ele) => {
						if (ele.role === "admin") {
							return [ele, ...acc];
						}
						return [...acc, ele];
					}, [])
					// sort the normal users by used_by_current_month.amount
					.sort((a, b) => {
						if ((a.role === "admin") | (b.role === "admin")) {
							return 0;
						}
						return (
							b.used_by_current_month.amount - a.used_by_current_month.amount
						);
					})
					.map((element, index) => (
						<UserComp user={element} index={index} key={index} />
					))
				}
			</div>
		</div >
	);
};


export default Home;
