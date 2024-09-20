import React, { useState, useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import { useNavigate, Navigate } from "react-router-dom";
import axios from "axios";
import { alert } from "../store/message";
import { login } from "../store/login";

const Login = () => {
	const [name, setName] = useState("");
	const [password, setPassword] = useState("");

	const dispatch = useDispatch();
	const navigate = useNavigate();

	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);

	const handleSubmit = (e) => {
		e.preventDefault();

		axios
			.post(process.env.REACT_APP_API_HOST + "login", {
				email_as_id: name,
				password: password,
			})
			.then((response) => {
				if (response.data) {
					console.log(response.data);
					localStorage.setItem("token", JSON.stringify(response.data.token));
					dispatch(login({ token: response.data.token }));
					// navigate("/mypanel");
				}
			})
			.catch((err) => {
				if (err.response) {
					dispatch(alert({ show: true, content: err.response.data.error }));
				} else {
					dispatch(
						alert({ show: true, content: "name or password 输入错误!" })
					);
				}
				console.log(err.toString());
			});
	};

	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(alert({ show: false }));
			}, 10000);
		}
	}, [dispatch, message]);

	if (loginState.isLogin) {
		return <Navigate to="/mypanel" />;
	}

	return (
		<div className="container px-5 py-24 mx-auto md:w-96">
			<form onSubmit={handleSubmit}>
				<h3 className="mb-4 text-xl font-medium text-white-900 dark:text-white">Welcome to sign in!</h3>
				<div className="mb-6">
					<label htmlFor="email" className="block mb-2 text-sm font-medium text-white-900 dark:text-gray-300">Username:</label>
					<input type="text" placeholder="username" onChange={(e) => setName(e.target.value)} className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" required />
				</div>
				<div className="mb-6">
					<label htmlFor="password" className="block mb-2 text-sm font-medium text-white-900 dark:text-gray-300">Password</label>
					<input type="password" placeholder="********" autoComplete="" onChange={(e) => setPassword(e.target.value)} className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" required="" />
				</div>
				
				<button type="submit" className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800">Submit</button>
				<div className="">
					{message.content}
				</div>
			</form>
		</div>
	);
};

export default Login;
