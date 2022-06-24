import React, { useState, useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import { Container, Alert } from "react-bootstrap";
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
				email: name,
				password: password,
			})
			.then((response) => {
				if (response.data) {
					localStorage.setItem("token", JSON.stringify(response.data.token));
					dispatch(login({ token: response.data.token }));
					navigate("/mypanel");
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
		<Container className="auth-wrapper">
			<form className="auth-inner" onSubmit={handleSubmit}>
				<h3>Sign In</h3>
				<div className="mb-3">
					<label>Name</label>
					<input
						type="input"
						className="form-control"
						placeholder="Enter name"
						onChange={(e) => setName(e.target.value)}
					/>
				</div>
				<div className="mb-3">
					<label>Password</label>
					<input
						type="password"
						className="form-control"
						placeholder="Enter password"
						onChange={(e) => setPassword(e.target.value)}
					/>
				</div>
				{/* <div className="mb-3">
					<div className="custom-control custom-checkbox">
						<input
							type="checkbox"
							className="custom-control-input"
							id="customCheck1"
						/>
						<label className="custom-control-label" htmlFor="customCheck1">
							Remember me
						</label>
					</div>
				</div> */}
				<div className="d-grid">
					<button type="submit" className="btn btn-primary">
						Submit
					</button>
				</div>
				<Alert show={message.show} variant={message.type}>
					{message.content}
				</Alert>
				{/* <p className="forgot-password text-right">
					Forgot <a href="#">password?</a>
				</p> */}
			</form>
		</Container>
	);
};

export default Login;
