import { useEffect, useState } from "react";
import {
	Container,
	Navbar,
	Nav,
	Button,
	Alert,
	Badge,
	ListGroup,
	Card,
} from "react-bootstrap";
import { useSelector, useDispatch } from "react-redux";
import { logout } from "../store/login";
import { alert } from "../store/message";
import axios from "axios";
import { formatBytes } from "../service/service";

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

		if (loginState.jwt.Role === "admin") {
			axios
				.get("http://localhost:8079/v1/alluser", {
					headers: { token },
				})
				.then((response) => {
					setUsers(response.data);
				})
				.catch((err) => {
					dispatch(alert({ show: true, content: err.toString() }));
				});
		} else if (loginState.jwt.Role === "normal") {
			axios
				.get("http://localhost:8079/v1/user/" + loginState.jwt.Email, {
					headers: { token },
				})
				.then((response) => {
					setUsers([response.data]);
				})
				.catch((err) => {
					dispatch(alert({ show: true, content: err.toString() }));
				});
		}
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
						<Nav className="me-auto">
							<Nav.Link href="/macos">MacOS</Nav.Link>
						</Nav>
						<Nav className="me-auto">
							<Nav.Link href="/windows">Windows</Nav.Link>
						</Nav>
						<Nav className="me-auto">
							<Nav.Link href="/iphone">IPhone/IPad</Nav.Link>
						</Nav>
						<Nav className="me-auto">
							<Nav.Link href="/android">Android</Nav.Link>
						</Nav>
					</Navbar.Collapse>
					<Navbar.Collapse className="justify-content-end">
						{loginState.jwt.Role === "admin" && (
							<Button variant="success">添加用户</Button>
						)}
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
				{loginState.jwt.Role === "admin" ? (
					<ListGroup as="ol" className="py-2" numbered>
						{users.map((element, index) => (
							<ListGroup.Item
								as="li"
								className="d-flex justify-content-between align-items-start"
								variant={element.email === loginState.jwt.Email ? "dark":"light"}
							>
								<div className="ms-2 me-auto">
									<div className="fw-bold">
										<h5>
											{element.email}
											<Badge bg="success" className="mx-1" pill>
												{element.role === "admin" ? "管理员" : "普通用户"}
											</Badge>
											<Badge variant="info" pill>
												{element.status === "plain" ? "在线" : "已下线"}
											</Badge>
										</h5>
									</div>
									<h6>
										<Badge bg="light" text="dark">
											总流量:{" "}
										</Badge>
										{formatBytes(element.credit)}
										<Badge bg="light" text="dark">
											已用流量:{" "}
										</Badge>
										{formatBytes(element.used)}
									</h6>
								</div>
								<div className="d-flex justify-content-center align-items-center my-auto">
									{element.status === "plain" ? (
										<Button variant="primary mx-1" size="lg">
											下线
										</Button>
									) : (
										<Button variant="primary mx-1" size="lg" disabled>
											下线
										</Button>
									)}
									{element.status === "plain" ? (
										<Button variant="primary mx-1" size="lg" disabled>
											上线
										</Button>
									) : (
										<Button variant="primary mx-1" size="lg">
											上线
										</Button>
									)}
									<Button variant="primary mx-1" size="lg">
										删除
									</Button>
								</div>
							</ListGroup.Item>
						))}
					</ListGroup>
				) : (
					<Card style={{ width: "auto" }}>
						<Card.Header>Basic Information</Card.Header>
						<Card.Body>
							<Card.Title>{users[0] && users[0].email}</Card.Title>
							<Card.Text>
								<h6>
									<Badge bg="light" text="dark">
										总流量:{" "}
									</Badge>
									{formatBytes(users[0] && users[0].credit)}
									<Badge bg="light" text="dark">
										已用流量:{" "}
									</Badge>
									{formatBytes(users[0] && users[0].used)}
								</h6>
							</Card.Text>
							<Button variant="primary">Go somewhere</Button>
						</Card.Body>
					</Card>
				)}
			</Container>
		</Container>
	);
};

export default Home;
