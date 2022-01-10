import { useCallback, useEffect, useState } from "react";
import {
	Container,
	Navbar,
	Nav,
	Button,
	Alert,
	Badge,
	ListGroup,
	Card,
	Modal,
	Form,
	Row,
	Col,
} from "react-bootstrap";
import { useSelector, useDispatch } from "react-redux";
import { logout } from "../store/login";
import { alert, success } from "../store/message";
import axios from "axios";
import { formatBytes } from "../service/service";

const token = JSON.parse(localStorage.getItem("token"));

function AddUser({ btnName, addUserFunc }) {
	const [show, setShow] = useState(false);
	const handleClose = () => setShow(false);
	const handleShow = () => setShow(true);

	const initialState = {
		email: "",
		password: "",
		name: "",
		path: "ray",
		role: "normal",
	};
	const [{ email, password, name, path, role }, setState] =
		useState(initialState);
	const clearState = () => {
		setState({ ...initialState });
	};

	const dispatch = useDispatch();
	const message = useSelector((state) => state.message);

	const handleAddUser = (e) => {
		e.preventDefault();
		axios({
			method: "post",
			url: process.env.REACT_APP_API_HOST + "signup",
			headers: { token },
			data: {
				role,
				email,
				password,
				name,
				path,
				status: "plain",
			},
		})
			.then((response) => {
				dispatch(success({ show: true, content: "user added in success!" }));
				addUserFunc();
				clearState();
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	};

	const onChange = (e) => {
		const { name, value } = e.target;
		setState((prevState) => ({ ...prevState, [name]: value }));
	};

	return (
		<>
			<Button variant="success" size="lg" onClick={handleShow}>
				{btnName}
			</Button>

			<Modal show={show} onHide={handleClose}>
				<Modal.Header closeButton>
					<Modal.Title>Add User</Modal.Title>
				</Modal.Header>
				<Modal.Body>
					<Form id="editForm" onSubmit={handleAddUser}>
						<Row className="mb-3">
							<Form.Group as={Col} controlId="formGridEmail">
								<Form.Label>Email(4-100 characters)*</Form.Label>
								<Form.Control
									type="input"
									name="email"
									placeholder="email"
									value={email}
									onChange={onChange}
									required
								/>
							</Form.Group>

							<Form.Group as={Col} controlId="formGridPassword">
								<Form.Label>Password(6+ characters)*</Form.Label>
								<Form.Control
									type="test"
									name="password"
									onChange={onChange}
									placeholder="password"
									value={password}
									required
								/>
							</Form.Group>
						</Row>

						<Row className="mb-3">
							<Form.Group as={Col} controlId="formGridUserType">
								<Form.Label>User Type</Form.Label>
								<Form.Select name="role" onChange={onChange} value={role}>
									<option value="admin">Admin</option>
									<option value="normal">Normal</option>
								</Form.Select>
							</Form.Group>

							<Form.Group as={Col} controlId="formGridTag">
								<Form.Label>Name</Form.Label>
								<Form.Control
									type="input"
									name="name"
									onChange={onChange}
									placeholder="name"
									value={name}
								/>
							</Form.Group>

							<Form.Group as={Col} controlId="formGridPath">
								<Form.Label>Path</Form.Label>
								<Form.Control
									type="input"
									name="path"
									onChange={onChange}
									placeholder="optional"
									value={path}
								/>
							</Form.Group>
						</Row>
					</Form>
					<Alert show={message.show} variant={message.type}>
						{message.content}
					</Alert>
				</Modal.Body>
				<Modal.Footer>
					<Button variant="secondary" onClick={handleClose}>
						关闭
					</Button>
					<Button type="submit" variant="primary" form="editForm">
						提交
					</Button>
				</Modal.Footer>
			</Modal>
		</>
	);
}

const Home = () => {
	const [users, setUsers] = useState([]);
	const [rerender, updateState] = useState(0);
	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);

	const dispatch = useDispatch();

	const handleLogout = (e) => {
		dispatch(logout());
	};

	const handleAddUser = (e) => {
		console.log(e);
	};

	const handleOnline = useCallback((name) => {
		axios
			.get(process.env.REACT_APP_API_HOST + "takeuseronline/" + name, {
				headers: { token },
			})
			.then((response) => {
				updateState(rerender + 5);
				dispatch(success({ show: true, content: response.data.message }));
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	}, []);

	const handleOffline = useCallback((name) => {
		axios
			.get(process.env.REACT_APP_API_HOST + "takeuseroffline/" + name, {
				headers: { token },
			})
			.then((response) => {
				updateState(rerender + 3);
				dispatch(success({ show: true, content: response.data.message }));
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	}, []);

	const handleDeleteUser = useCallback((name) => {
		axios
			.get(process.env.REACT_APP_API_HOST + "deluser/" + name, {
				headers: { token },
			})
			.then((response) => {
				updateState(rerender + 1);
				dispatch(success({ show: true, content: response.data.message }));
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	}, []);

	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(alert({ show: false }));
			}, 5000);
		}
	}, [message]);

	useEffect(() => {
		if (loginState.jwt.Role === "admin") {
			axios
				.get(process.env.REACT_APP_API_HOST + "alluser", {
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
				.get(process.env.REACT_APP_API_HOST + "user/" + loginState.jwt.Email, {
					headers: { token },
				})
				.then((response) => {
					setUsers([response.data]);
				})
				.catch((err) => {
					dispatch(alert({ show: true, content: err.toString() }));
				});
		}
	}, [dispatch, rerender]);

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
							<AddUser
								btnName="添加用户"
								addUserFunc={() => updateState(rerender + 9)}
							/>
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
				<Alert show={message.show} variant={message.type}>
					{" "}
					{message.content}{" "}
				</Alert>
				{loginState.jwt.Role === "admin" ? (
					<ListGroup as="ol" className="py-2" numbered>
						{users.map((element, index) => (
							<ListGroup.Item
								as="li"
								className="d-flex justify-content-between align-items-start"
							>
								<div className="ms-2 me-auto">
									<div className="fw-bold">
										<h5>
											{element.email}
											<Badge bg="success" className="mx-1" pill>
												{element.role === "admin" ? "管理员" : "普通用户"}
											</Badge>
											<Badge bg="primary" className="mx-1" pill>
												{element.status === "plain" ? "在线" : "已下线"}
											</Badge>
											{element.email === loginState.jwt.Email && (
												<Badge bg="info" className="mx-1" pill>
													It's Me
												</Badge>
											)}
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
									<EditUser
										btnName="Edit"
										editUserFunc={() => updateState(rerender + 7)}
										user={element}
									/>
									{element.status === "plain" ? (
										<Button
											onClick={() => handleOffline(element.email)}
											variant="success mx-1"
											size="lg"
										>
											Disable
										</Button>
									) : (
										<Button
											onClick={() => handleOffline(element.email)}
											variant="success mx-1"
											size="lg"
											disabled
										>
											Disable
										</Button>
									)}
									{element.status === "plain" ? (
										<Button
											onClick={() => handleOnline(element.email)}
											variant="success mx-1"
											size="lg"
											disabled
										>
											Enable
										</Button>
									) : (
										<Button
											onClick={() => handleOnline(element.email)}
											variant="success mx-1"
											size="lg"
										>
											Enable
										</Button>
									)}
									<ConfirmDelUser
										btnName="Delete User"
										deleteUserFunc={() => handleDeleteUser(element.email)}
									/>
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

function ConfirmDelUser({ btnName, deleteUserFunc }) {
	const [show, setShow] = useState(false);

	const handleClose = () => setShow(false);
	const handleShow = () => setShow(true);

	return (
		<>
			<Button variant="dark mx-1" size="lg" onClick={handleShow}>
				{btnName}
			</Button>

			<Modal show={show} onHide={handleClose}>
				<Modal.Header closeButton>
					<Modal.Title>Notice!</Modal.Title>
				</Modal.Header>
				<Modal.Body>确认删除用户？</Modal.Body>
				<Modal.Footer>
					<Button variant="secondary" onClick={handleClose}>
						关闭
					</Button>
					<Button
						variant="primary"
						onClick={() => {
							deleteUserFunc();
							handleClose();
						}}
					>
						确认
					</Button>
				</Modal.Footer>
			</Modal>
		</>
	);
}

function EditUser({ btnName, user, editUserFunc }) {

	const [show, setShow] = useState(false);
	const handleClose = () => setShow(false);
	const handleShow = () => setShow(true);

	const [status, setStatus] = useState(user.status);
	const [{ used, password, name, role, credit }, setState] = useState({
		used: user.used,
		password: user.password,
		name: user.name,
		role: user.role,
		credit:user.credit
	});

	const onChange = (e) => {
		const { name, value } = e.target;
		setState((prevState) => ({ ...prevState, [name]: value }));
	};

	const dispatch = useDispatch();
	const message = useSelector((state) => state.message);

	useEffect(() => {
		setStatus(user.status);
	}, [user.status]);

	const handleEditUser = (e) => {
		e.preventDefault();
		axios({
			method: "post",
			url: process.env.REACT_APP_API_HOST + "edit/" + user.email,
			headers: { token },
			data: {
				role,
				email: user.email,
				password,
				name,
				used: parseInt(used),
				credit: parseInt(credit),
			},
		})
			.then((response) => {
				dispatch(success({ show: true, content: "user info updated!" }));
				editUserFunc();
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	};

	return (
		<>
			<Button variant="outline-success mx-1" size="lg" onClick={handleShow}>
				{btnName}
			</Button>

			<Modal show={show} onHide={handleClose}>
				<Modal.Header closeButton>
					<Modal.Title>Edit User</Modal.Title>
				</Modal.Header>
				<Modal.Body>
					<Form id="editForm" onSubmit={handleEditUser}>
						<Row className="mb-3">
							<Form.Group controlId="formGridDomains">
								<Form.Label>User Status:</Form.Label>
								<Badge pill bg="light" text="dark" className="mx-1">
									<b className="h4">{status}</b>
								</Badge>
							</Form.Group>
						</Row>
						<Row className="mb-3">
							<Form.Group as={Col} controlId="formGridName">
								<Form.Label>Email</Form.Label>
								<Form.Control
									type="input"
									name="email"
									placeholder={user.email}
									value={user.email}
									disabled
								/>
							</Form.Group>

							<Form.Group as={Col} controlId="formGridPassword">
								<Form.Label>Password</Form.Label>
								<Form.Control
									type="password"
									name="password"
									onChange={onChange}
									placeholder="Password"
									value={password}
									autoComplete=""
								/>
							</Form.Group>
						</Row>
						<hr />
						<Row className="mb-3">
							<Form.Group as={Col} controlId="formGridUserType">
								<Form.Label>User Type</Form.Label>
								<Form.Select
									name="role"
									onChange={onChange}
									value={role}
								>
									<option value="admin">Admin</option>
									<option value="normal">Normal</option>
								</Form.Select>
							</Form.Group>

							<Form.Group as={Col} controlId="formGridTag">
								<Form.Label>Name</Form.Label>
								<Form.Control
									type="input"
									name="name"
									onChange={onChange}
									placeholder={user.name}
									value={name}
								/>
							</Form.Group>
						</Row>

						<Row className="mb-3"></Row>

						<Row className="mb-3">
							<Form.Group as={Col} controlId="formGridTrafficeUsed">
								<Form.Label>已用流量</Form.Label>
								<Form.Control
									type="number"
									name="used"
									onChange={onChange}
									placeholder={user.used}
									value={used}
								/>
							</Form.Group>

							<Form.Group as={Col} controlId="formGridTrafficCredit">
								<Form.Label>每月限额</Form.Label>
								<Form.Control
									type="number"
									name="credit"
									onChange={onChange}
									placeholder={user.credit}
									value={credit}
								/>
							</Form.Group>
						</Row>

						<hr />
						<Row className="mb-3">
							<Form.Group controlId="formGridDomains">
								<Form.Label>Domains: </Form.Label>
								<Badge pill bg="light" text="dark">
									<b className="h5">
										{user.domain ? user.domain : "w8.undervineyard.com"}
									</b>
								</Badge>
							</Form.Group>

							<Form.Group controlId="formGridPath">
								<Form.Label>Path: </Form.Label>
								<Badge pill bg="light" text="dark" className="mx-1">
									<b className="h5">{user.path}</b>
								</Badge>
							</Form.Group>

							<Form.Group controlId="formGridUuid">
								<Form.Label>UUID: </Form.Label>
								<Badge pill bg="light" text="dark" className="mx-1">
									<b className="h5">{user.uuid}</b>
								</Badge>
							</Form.Group>
						</Row>
					</Form>
					<Alert show={message.show} variant={message.type}>
						{message.content}
					</Alert>
				</Modal.Body>
				<Modal.Footer>
					<Button variant="secondary" onClick={handleClose}>
						关闭
					</Button>
					<Button type="submit" variant="primary" form="editForm">
						提交
					</Button>
				</Modal.Footer>
			</Modal>
		</>
	);
}

export default Home;
