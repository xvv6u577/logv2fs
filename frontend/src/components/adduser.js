import { useState } from "react";
import {
	Button,
	Alert,
	Modal,
	Form,
	Row,
	Col,
} from "react-bootstrap";
import { useSelector, useDispatch } from "react-redux";
import { alert, success } from "../store/message";
import { doRerender } from "../store/rerender";
import axios from "axios";

function AddUser({ btnName }) {

    const loginState = useSelector((state) => state.login);
    const dispatch = useDispatch();
    const message = useSelector((state) => state.message);
	const rerenderSignal = useSelector((state) => state.rerender);

	const [show, setShow] = useState(false);
	const handleClose = () => setShow(false);
	const handleShow = () => setShow(true);

	const initialState = {
		email: "",
		password: "",
		name: "",
		path: "ray",
		role: "normal",
		uuid: ""
	};
	const [{ email, password, name, path, role, uuid }, setState] =
		useState(initialState);
	const clearState = () => {
		setState({ ...initialState });
	};

	const handleAddUser = (e) => {
		e.preventDefault();
		axios({
			method: "post",
			url: process.env.REACT_APP_API_HOST + "signup",
			headers: { token: loginState.token },
			data: {
				role,
				email,
				password,
				name,
				path,
				status: "plain",
				uuid
			},
		})
			.then((response) => {
				dispatch(success({ show: true, content: "user " + name +  " added in success!" }));
				dispatch(doRerender({rerender: !rerenderSignal.rerender}))
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
			<Button variant="success" size="" onClick={handleShow}>
				{btnName}
			</Button>

			<Modal show={show} onHide={handleClose} size="lg" aria-labelledby="contained-modal-title-vcenter" centered >
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
								<Form.Label>User Type(optinal)</Form.Label>
								<Form.Select name="role" onChange={onChange} value={role}>
									<option value="admin">Admin</option>
									<option value="normal">Normal</option>
								</Form.Select>
							</Form.Group>

							<Form.Group as={Col} controlId="formGridTag">
								<Form.Label>Name(optinal)</Form.Label>
								<Form.Control
									type="input"
									name="name"
									onChange={onChange}
									placeholder="name"
									value={name}
								/>
							</Form.Group>

							<Form.Group as={Col} controlId="formGridPath">
								<Form.Label>Path(optinal)</Form.Label>
								<Form.Control
									type="input"
									name="path"
									onChange={onChange}
									placeholder="optional"
									value={path}
								/>
							</Form.Group>
						</Row>
						<Row className="mb-3">
							<Form.Group as={Col} controlId="formGridUserUuid">
								<Form.Label>UUID(optinal)</Form.Label>
								<Form.Control
									type="input"
									name="uuid"
									onChange={onChange}
									placeholder="UUID"
									value={uuid}
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

export default AddUser