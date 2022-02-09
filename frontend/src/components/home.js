import { useEffect, useState } from "react";
import {
	Container,
	Button,
	Alert,
	Badge,
	ListGroup,
	Modal,
	Form,
	Row,
	Col,
	OverlayTrigger,
	Tooltip,
	Accordion,
} from "react-bootstrap";
import { useSelector, useDispatch } from "react-redux";
import { alert, success } from "../store/message";
import { doRerender } from "../store/rerender";
import { formatBytes } from "../service/service";
import axios from "axios";
import { logout } from "../store/login";

const Home = () => {
	const [users, setUsers] = useState([]);
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
				dispatch(doRerender({ rerender: !rerenderSignal.rerender }));
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
				dispatch(doRerender({ rerender: !rerenderSignal.rerender }));
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
				dispatch(doRerender({ rerender: !rerenderSignal.rerender }));
				dispatch(success({ show: true, content: response.data.message }));
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	};

	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(alert({ show: false }));
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
					(ele) => ele.email === loginState.jwt.Email
				);
				if (user.length !== 0) {
					setUsers(response.data);
				} else {
					dispatch(logout());
				}
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	}, [rerenderSignal, loginState.jwt.Email, loginState.token, dispatch]);

	return (
		<Container className="my-3 home-list">
			<Alert show={message.show} variant={message.type}>
				{" "}
				{message.content}{" "}
			</Alert>
			<ListGroup
				as="ol"
				className="list-group list-group-striped list-group-hover"
				numbered
			>
				<Accordion>
					{users.map((element, index) => (
						<Accordion.Item eventKey={index}>
							<ListGroup.Item
								as="li"
								className="d-flex justify-content-between align-items-start"
							>
								<div className="ms-2 me-auto">
									{/* <OverlayTrigger
									key={index}
									placement="right"
									overlay={
										<Tooltip id={`tooltip-${index}`} className="myToolTip">
											domain:{" "}
											<b>{Object.values(element.nodeinuse).toString()}</b>{" "}
											<br />
											uuid: <b>{element.uuid}</b>
											<br />
											path: <b>{element.path}</b>
											<br />
										</Tooltip>
									}
								>
								</OverlayTrigger> */}
									{/* <div
										className="fw-bold info-hover"
										onClick={() => {
											navigator.clipboard.writeText(
												process.env.REACT_APP_API_HOST +
													"suburl/" +
													element.email
											);
										}}
									>
									</div> */}
									<span className="home-traffic-fs">{index+1}</span>
									{"."}
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

									<span className="home-traffic-fs">
										今日流量: {formatBytes(element.used_by_current_day.amount)}
										{"， "}
										本月流量:{" "}
										{formatBytes(element.used_by_current_month.amount)}
										{"， "}
										已用总流量: {formatBytes(element.used)}
									</span>
								</div>
								<div className="d-flex justify-content-center align-items-center">
									<EditUser
										btnName="Edit"
										editUserFunc={() =>
											dispatch(
												doRerender({ rerender: !rerenderSignal.rerender })
											)
										}
										user={element}
									/>
									{element.status === "plain" ? (
										<Button
											onClick={() => handleOffline(element.email)}
											variant="success mx-1"
											size="sm"
										>
											Disable
										</Button>
									) : (
										<Button
											onClick={() => handleOffline(element.email)}
											variant="success mx-1"
											size="sm"
											disabled
										>
											Disable
										</Button>
									)}
									{element.status === "plain" ? (
										<Button
											onClick={() => handleOnline(element.email)}
											variant="success mx-1"
											size="sm"
											disabled
										>
											Enable
										</Button>
									) : (
										<Button
											onClick={() => handleOnline(element.email)}
											variant="success mx-1"
											size="sm"
										>
											Enable
										</Button>
									)}
									<ConfirmDelUser
										btnName="Delete"
										deleteUserFunc={() => handleDeleteUser(element.email)}
									/>
									<Accordion.Header></Accordion.Header>
								</div>
							</ListGroup.Item>
							<Accordion.Body>
								<div className="py-1">
									<p className="lh-sm">用户名: {element.email}</p>
									<p className="lh-sm">密码: {element.email.length < 6 ? "mypassword":element.email}</p>
									<p className="text-break lh-sm">Subscription: {element.suburl}</p>
								</div>
								<div className="home-traffic-fs">
									<h6 className="">每月流量</h6>
									{element.traffic_by_month &&
										element.traffic_by_month
											.sort((a, b) => b.period - a.period)
											.map((element) => {
												return (
													<Badge pill bg="dark" text="white">
														{element.period} / {formatBytes(element.amount)}
													</Badge>
												);
											})}
								</div>
								<div className="home-traffic-fs">
									<h6 className="pt-2">每日流量</h6>
									{element.traffic_by_day &&
										element.traffic_by_day
											.sort((a, b) => b.period - a.period)
											// .slice(0, 90)
											.map((element) => {
												return (
													<Badge pill bg="dark" text="white">
														{element.period} / {formatBytes(element.amount)}
													</Badge>
												);
											})}
								</div>
							</Accordion.Body>
						</Accordion.Item>
					))}
				</Accordion>
			</ListGroup>
		</Container>
	);
};

function ConfirmDelUser({ btnName, deleteUserFunc }) {
	const [show, setShow] = useState(false);

	const handleClose = () => setShow(false);
	const handleShow = () => setShow(true);

	return (
		<>
			<Button variant="dark mx-1" size="sm" onClick={handleShow}>
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
			<Button variant="outline-success mx-1" size="sm" onClick={handleShow}>
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
