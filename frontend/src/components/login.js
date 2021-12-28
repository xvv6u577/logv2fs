import React, { useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { Container, Form, Button, Card, Alert } from "react-bootstrap";
import { Navigate } from "react-router-dom";
import axios from "axios";
import { login } from "../store/login";
import { alert } from "../store/message";

const Login = () => {
	const [name, setName] = useState("");
	const [password, setPassword] = useState("");

	const dispatch = useDispatch();
	// const user = useSelector(state => state.login.user);
	const isLogin = useSelector((state) => state.login.isLogin);
	const content = useSelector((state) => state.message.content);
	const show = useSelector((state) => state.message.show);

	const handleSubmit = (e) => {
		e.preventDefault();

		axios
			.post("http://localhost:8079/v1/login", {
				email: name,
				password: password,
			})
			.then((response) => {
				dispatch(login({ isLogin: true, user: response.data }));
				if (response.data) {
					localStorage.setItem("token", JSON.stringify(response.data.token));
				}
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});

	};

	if (isLogin) {
		return <Navigate to="/home" />;
	}

	return (
		<Container className="login d-flex justify-content-center align-items-center">
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
				<Alert show={show} variant="danger">{content}</Alert>
			</Card>
		</Container>
	);
};

export default Login;
