import { useEffect, useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, reset } from "../store/message";
import axios from "axios";
import { logout } from "../store/login";
import UserComp from "./userComp";
import Alert from "./alert";
import { IStore as RootState, User } from "../types";

const Home = () => {

	const [users, setUsers] = useState<User[]>([]);
	const [activeTab, setActiveTab] = useState<number>(-1);

	const dispatch = useDispatch();
	const loginState = useSelector((state: RootState) => state.login);
	const message = useSelector((state: RootState) => state.message);
	const rerenderSignal = useSelector((state: RootState) => state.rerender);

	const activateTab = (index: number) => {
		activeTab === index ? setActiveTab(-1) : setActiveTab(index);
	  };

	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(reset());
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
					(ele: User) => ele.email === loginState.jwt.Email
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
		<div className="my-3 px-3">
			<Alert message={message.content} type={message.type} shown={message.show} close={() => { dispatch(reset()); }} />

			<div className="flex justify-start">
				{/* four react buttons to enable filer */}
				<button
					className="w-30 focus:outline-none text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-2.5 py-2.5 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
					onClick={() => {
						const updatedUsers: User[] = users
							// put admin at the top of the list
							.reduce((acc: User[], ele: User) => {
								if (ele.role === "admin") {
									return [ele, ...acc];
								}
								return [...acc, ele];
							}, [])
							// sort the normal users by used_by_current_month.amount
							.sort((a: User, b: User) => {
								if ((a.role === "admin") || (b.role === "admin")) {
									return 0;
								}
								return (
									b.used_by_current_month.amount - a.used_by_current_month.amount
								);
							});
						setUsers(updatedUsers);
					}}
				>
					By Role
				</button>
				<button
					className="w-30 focus:outline-none text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-2.5 py-2.5 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
					onClick={() => {
						const updatedUsers = users
							// put status:plain at the top of the list
							.reduce((acc: User[], ele: User) => {
								if (ele.status === "plain") {
									return [...acc, ele];
								}
								return [ele, ...acc];
							}, [])
							// sort the plain users by used_by_current_month.amount
							.sort((a: User, b: User) => {
								if ((a.status !== "plain") || (b.status !== "plain")) {
									return 0;
								}
								return (
									b.used_by_current_month.amount - a.used_by_current_month.amount
								);
							});
						setUsers(updatedUsers);
					}}
				>
					Online
				</button>
				<button
					className="w-30 focus:outline-none text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-2.5 py-2.5 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
					onClick={() => { 
						const updatedUsers = users
							// put status:plain at the top of the list
							.reduce((acc: User[], ele: User) => {
								if (ele.status === "plain") {
									return [ele, ...acc];
								}
								return [...acc, ele];
							}, [])
							// sort the plain users by used_by_current_month.amount
							.sort((a, b) => {
								if ((a.status !== "plain") || (b.status !== "plain")) {
									return 0;
								}
								return (
									b.used_by_current_day.amount - a.used_by_current_day.amount
								);
							});
						setUsers(updatedUsers);
					 }}
				>
					Today
				</button>
				<button
					className="w-30 focus:outline-none text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-2.5 py-2.5 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
					onClick={() => { 
						const updatedUsers = users
							// put status:plain at the top of the list
							.reduce((acc: User[], ele: User) => {
								if (ele.status === "plain") {
									return [ele, ...acc];
								}
								return [...acc, ele];
							}, [])
							// sort the plain users by used_by_current_month.amount
							.sort((a, b) => {
								if ((a.status !== "plain") || (b.status !== "plain")) {
									return 0;
								}
								return (
									b.used_by_current_month.amount - a.used_by_current_month.amount
								);
							});
						setUsers(updatedUsers);
					 }}
				>
					By Month
				</button>
				<button
					className="w-30 focus:outline-none text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-2.5 py-2.5 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
					onClick={() => { 
						const updatedUsers = users
							// put status:plain at the top of the list
							.reduce((acc: User[], ele: User) => {
								if (ele.status === "plain") {
									return [ele, ...acc];
								}
								return [...acc, ele];
							}, [])
							// sort the plain users by used
							.sort((a, b) => {
								if ((a.status !== "plain") || (b.status !== "plain")) {
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
