import React, { useState, useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import { Container, Form, Button, Card, Alert } from "react-bootstrap";
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
					navigate("/home");
				}
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.data.toString() }));
			});
	};

	useEffect(()=>{
		if (message.show === true) {
			setTimeout(()=>{
				dispatch(alert({show: false}))
			}, 5000)
		}
	},[dispatch, message])

	if (loginState.isLogin) {
		return <Navigate to="/home" />;
	}

	return (
		<Container className="main d-flex justify-content-center align-items-center">
			<Card>
				<Card.Body>
					<Card.Title>Start from login</Card.Title>
					<Card.Subtitle className="mb-2 text-muted"></Card.Subtitle>
					<Form onSubmit={handleSubmit}>
						<Form.Group className="mb-3" controlId="formBasicEmail">
							<Form.Label>Name</Form.Label>
							<Form.Control
								type="input"
								placeholder="Name"
								onChange={(e) => setName(e.target.value)}
								autoComplete=""
							/>
						</Form.Group>

						<Form.Group className="mb-3" controlId="formBasicPassword">
							<Form.Label>Password</Form.Label>
							<Form.Control
								type="password"
								placeholder="Password"
								onChange={(e) => setPassword(e.target.value)}
								autoComplete=""
							/>
						</Form.Group>
						<Button variant="primary" type="submit">
							Submit
						</Button>
					</Form>
				</Card.Body>
				<Alert show={message.show} variant={message.type}>
					{message.content}
				</Alert>
			</Card>
		</Container>
	);
};

export default Login;
