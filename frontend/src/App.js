import React from "react";
import { Navbar, Nav, Container, Form, Button } from "react-bootstrap";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import "./App.css";

const Navitation = () => {
	return (
		<Navbar collapseOnSelect expand="lg" bg="" variant="dark">
			<Container>
				<Navbar.Brand href="/">LogV2ray Backend</Navbar.Brand>
				<Navbar.Toggle id="responsive-navbar-nav" />
				<Navbar.Collapse id="responsive-navbar-nav">
					<Nav className="me-auto">
						<Nav.Link href="/">Home</Nav.Link>
					</Nav>
					<Nav>
						<Nav.Link href="/login">Login</Nav.Link>
						<Nav.Link href="/signup">Sign up</Nav.Link>
					</Nav>
				</Navbar.Collapse>
			</Container>
		</Navbar>
	);
};

const Home = () => {
	return (
		<Container className="d-flex justify-content-center align-items-center">
			<h2>Home</h2>
		</Container>
	);
};
const Login = () => {
	return (
		<Container className="d-flex justify-content-center align-items-center">
			<Form className="form">
				<Form.Group className="mb-3" controlId="formBasicEmail">
					<Form.Label>Email address</Form.Label>
					<Form.Control type="email" placeholder="Enter email" />
				</Form.Group>

				<Form.Group className="mb-3" controlId="formBasicPassword">
					<Form.Label>Password</Form.Label>
					<Form.Control type="password" placeholder="Password" />
				</Form.Group>
				<Button variant="success" type="submit">
					Submit
				</Button>
			</Form>
		</Container>
	);
};
const Signup = () => {
	return (
		<Container className="d-flex justify-content-center align-items-center">
			<h2>signup</h2>
		</Container>
	)
};

const App = () => {
	return (
		<Router>
			<Navitation />
			<Routes>
				<Route path="/" element={<Home />} />
				<Route path="/login" element={<Login />} />
				<Route path="/signup" element={<Signup />} />
			</Routes>
		</Router>
	);
};

export default App;
