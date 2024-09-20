import { useEffect, useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, reset } from "../store/message";
import axios from "axios";
import { logout } from "../store/login";
import UserComp from "./userComp";
import Alert from "./alert";

const Home = () => {

	const [users, setUsers] = useState([]);
	const [activeTab, setActiveTab] = useState(-1);

	const dispatch = useDispatch();
	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);
	const rerenderSignal = useSelector((state) => state.rerender);

	const activateTab = (index) => {
		activeTab === index ? setActiveTab(-1) : setActiveTab(index);
	};

	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(reset({}));
			}, 5000);
		}
	}, [message, dispatch]);

	useEffect(() => {
		axios
			.get(process.env.REACT_APP_API_HOST + "n778cf", {
				headers: { token: loginState.token },
			})
			.then((response) => {
				setUsers(response.data);
				// let user = response.data.filter(
				// 	(ele) => ele.email === loginState.jwt.Email
				// );
				// if (user.length !== 0) {
				// } else {
				// 	dispatch(logout());
				// }
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
		<div className="my-3 px-3">
			<Alert message={message.content} type={message.type} shown={message.show} close={() => { dispatch(reset({})); }} />

			<div className="flex justify-start">
				{/* four react buttons to enable filer */}
				<button
					className="w-20 focus:outline-none text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-1.5 py-1 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
					onClick={() => {
						const updatedUsers = users
							.reduce((acc, user) => {
								return user.role === "admin" ? [user, ...acc] : [...acc, user];
							}, [])
							.sort((a, b) => {
								if (a.role === "admin" || b.role === "admin") return 0;

								const getTraffic = user => user.monthly_logs?.[0]?.traffic ?? 0;
								return getTraffic(b) - getTraffic(a);
							});
						setUsers(updatedUsers);
					}}
				>
					By Role
				</button>
				<button
					className="w-20 focus:outline-none text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-1.5 py-1 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
					onClick={() => {
						const updatedUsers = users
							// put status:plain at the top of the list
							.reduce((acc, ele) => {
								if (ele.status === "plain") {
									return [ele, ...acc];
								}
								return [...acc, ele];
							}, [])
							// sort the plain users by user.used
							.sort((a, b) => {
								if ((a.status !== "plain") | (b.status !== "plain")) {
									return 0;
								}

								const getTraffic = user => user.monthly_logs?.[0]?.traffic ?? 0;
								return getTraffic(b) - getTraffic(a);
							});
						setUsers(updatedUsers);
					}}
				>
					Online
				</button>
				<button
					className="w-20 focus:outline-none text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-1.5 py-1 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
					onClick={() => {
						const updatedUsers = users
							// put status:plain at the top of the list
							.reduce((acc, ele) => {
								if (ele.status === "plain") {
									return [ele, ...acc];
								}
								return [...acc, ele];
							}, [])
							// sort the plain users by used_by_current_month.amount
							.sort((a, b) => {
								if ((a.status !== "plain") | (b.status !== "plain")) {
									return 0;
								}

								const getTraffic = user => user.daily_logs?.[0]?.traffic ?? 0;
								return getTraffic(b) - getTraffic(a);
							});
						setUsers(updatedUsers);
					}}
				>
					Today
				</button>
				<button
					className="w-20 focus:outline-none text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-1.5 py-1 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
					onClick={() => {
						const updatedUsers = users
							// put status:plain at the top of the list
							.reduce((acc, ele) => {
								if (ele.status === "plain") {
									return [ele, ...acc];
								}
								return [...acc, ele];
							}, [])
							// sort the plain users by used_by_current_month.amount
							.sort((a, b) => {
								if ((a.status !== "plain") | (b.status !== "plain")) {
									return 0;
								}
								
								const getTraffic = user => user.monthly_logs?.[0]?.traffic ?? 0;
								return getTraffic(b) - getTraffic(a);
							});
						setUsers(updatedUsers);
					}}
				>
					By Month
				</button>
				<button
					className="w-20 focus:outline-none text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-1.5 py-1 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
					onClick={() => {
						const updatedUsers = users
							// put status:plain at the top of the list
							.reduce((acc, ele) => {
								if (ele.status === "plain") {
									return [ele, ...acc];
								}
								return [...acc, ele];
							}, [])
							// sort the plain users by used
							.sort((a, b) => {
								if ((a.status !== "plain") | (b.status !== "plain")) {
									return 0;
								}
								return (
									b.used - a.used
								);
							});
						setUsers(updatedUsers);
					}}
				>
					By Used
				</button>
			</div>

			<div id="accordion-collapse" data-accordion="collapse">
				{users.map((element, index) => (
					<UserComp
						user={element}
						index={index}
						key={index}
						active={!(activeTab === index)}
						update={() => activateTab(index)}
					/>
				))
				}
			</div>
		</div >
	);
};


export default Home;
