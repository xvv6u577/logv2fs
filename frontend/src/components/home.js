import { useEffect, useState } from "react";
import {
	Container,
	Navbar,
	Nav,
	Button,
	Alert,
	Badge,
	ListGroup,
	Card
} from "react-bootstrap";
import { useSelector, useDispatch } from "react-redux";
import { logout } from "../store/login";
import { alert } from "../store/message";
import axios from "axios";

const Home = () => {
	const [users, setUsers] = useState([]);
	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);
	const dispatch = useDispatch();

	const handleLogout = (e) => {
		dispatch(logout());
	};

	useEffect(() => {
		const token = JSON.parse(localStorage.getItem("token"));
		axios
			.get("http://localhost:8079/v1/alluser", {
				headers: { token },
			})
			.then((response) => {
				setUsers(response.data);
				console.log(response.data);
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	}, [dispatch]);

	return (
		<Container className="main" fluid>
			<Navbar bg="light" expand="lg">
				<Container>
					<Navbar.Brand href="/">Logv2ray Frontend</Navbar.Brand>
					<Navbar.Toggle aria-controls="basic-navbar-nav" />
					<Navbar.Collapse id="basic-navbar-nav">
						<Nav className="me-auto">
							<Nav.Link href="/home">Home</Nav.Link>
						</Nav>
					</Navbar.Collapse>
					<Navbar.Collapse className="justify-content-end">
					<Button variant="success">添加用户</Button>
						<Navbar.Text className="mx-2">
							Signed in as: <b>{loginState.jwt.Email}</b>,
						</Navbar.Text>
						<Navbar.Text>
							<Button variant="link" onClick={handleLogout}>
								logout
							</Button>
						</Navbar.Text>
					</Navbar.Collapse>
				</Container>
			</Navbar>
			<Container className="py-3">
				<Alert show={message.show} variant="danger">
					{" "}
					{message.content}{" "}
				</Alert>
				<ListGroup as="ol" className="py-2" numbered>
					{users.map((element, index) => (
						<ListGroup.Item
							as="li"
							className="d-flex justify-content-between align-items-start"
						>
							<div className="ms-2 me-auto">
								<div className="fw-bold">
									<h5>{element.email}<Badge bg="info" className="mx-1">{element.role === "admin" ? "管理员":"用户"}</Badge>
										<Badge variant="primary" pill>
											{element.status === "plain" ? "在线" : "已下线"}
										</Badge>
									</h5>
								</div>
								<h6>已用流量: <Badge bg="light" text="dark">{element.used}</Badge>总流量: <Badge bg="light" text="dark">{element.credit}</Badge></h6>
							</div>
							<div className="d-flex justify-content-center align-items-center my-auto">
								<Button variant="success mx-1">上线</Button>
								<Button variant="success mx-1">下线</Button>
								<Button variant="success mx-1">删除</Button>
							</div>
						</ListGroup.Item>
					))}
				</ListGroup>
			</Container>
		</Container>
	);
};

export default Home;
