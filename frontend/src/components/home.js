import { useCallback, useEffect, useState } from "react";
import {
	Container,
	Button,
	Alert,
	Badge,
	ListGroup,
	Card,
	Modal,
	Form,
	Row,
	Col,
	OverlayTrigger,
	Tooltip,
} from "react-bootstrap";
import { useSelector, useDispatch } from "react-redux";
import { alert, messageSlice, success } from "../store/message";
import { formatBytes } from "../service/service";
import axios from "axios";

const Home = () => {
	const [users, setUsers] = useState([]);
	const [rerender, updateState] = useState(0);
	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);
	const rerenderSignal = useSelector((state) => state.rerender);

	const dispatch = useDispatch();

	const handleOnline = (name) => {
		axios
			.get(process.env.REACT_APP_API_HOST + "takeuseronline/" + name, {
				headers: { token: loginState.token },
			})
			.then((response) => {
				updateState(rerender + 5);
				dispatch(success({ show: true, content: response.data.message }));
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	};

	const handleOffline = (name) => {
		axios
			.get(process.env.REACT_APP_API_HOST + "takeuseroffline/" + name, {
				headers: { token: loginState.token },
			})
			.then((response) => {
				updateState(rerender + 3);
				dispatch(success({ show: true, content: response.data.message }));
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	};

	const handleDeleteUser = (name) => {
		axios
			.get(process.env.REACT_APP_API_HOST + "deluser/" + name, {
				headers: { token: loginState.token },
			})
			.then((response) => {
				updateState(rerender + 1);
				dispatch(success({ show: true, content: response.data.message }));
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	};

	useEffect(() => {
		updateState(rerender + 9);
	}, [rerenderSignal]);

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
					headers: { token: loginState.token },
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
					headers: { token: loginState.token },
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
								<OverlayTrigger
									key={index}
									placement="right"
									overlay={
										<Tooltip id={`tooltip-${index}`}>
											Tooltip on <strong>{index}</strong>.
										</Tooltip>
									}
								>
									<div className="fw-bold">
										<h5>
											<b>{element.name}</b>
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
								</OverlayTrigger>

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
										size=""
									>
										Disable
									</Button>
								) : (
									<Button
										onClick={() => handleOffline(element.email)}
										variant="success mx-1"
										size=""
										disabled
									>
										Disable
									</Button>
								)}
								{element.status === "plain" ? (
									<Button
										onClick={() => handleOnline(element.email)}
										variant="success mx-1"
										size=""
										disabled
									>
										Enable
									</Button>
								) : (
									<Button
										onClick={() => handleOnline(element.email)}
										variant="success mx-1"
										size=""
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
	);
};

function ConfirmDelUser({ btnName, deleteUserFunc }) {
	const [show, setShow] = useState(false);

	const handleClose = () => setShow(false);
	const handleShow = () => setShow(true);

	return (
		<>
			<Button variant="dark mx-1" size="" onClick={handleShow}>
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
		credit: user.credit,
	});

	const onChange = (e) => {
		const { name, value } = e.target;
		setState((prevState) => ({ ...prevState, [name]: value }));
	};

	const dispatch = useDispatch();
	const message = useSelector((state) => state.message);
	const loginState = useSelector((state) => state.login);

	useEffect(() => {
		setStatus(user.status);
	}, [user.status]);

	const handleEditUser = (e) => {
		e.preventDefault();
		axios({
			method: "post",
			url: process.env.REACT_APP_API_HOST + "edit/" + user.email,
			headers: { token: loginState.token },
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
			<Button variant="outline-success mx-1" size="" onClick={handleShow}>
				{btnName}
			</Button>

			<Modal
				show={show}
				onHide={handleClose}
				size="lg"
				aria-labelledby="contained-modal-title-vcenter"
				centered
			>
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
