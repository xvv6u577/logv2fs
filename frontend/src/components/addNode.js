import { useState,useEffect } from "react";
import {
	Modal,
	Button,
	Form,
	ListGroup,
	InputGroup,
	FormControl,
} from "react-bootstrap";
import { useSelector, useDispatch } from "react-redux";
import { alert, success } from "../store/message";
import { doRerender } from "../store/rerender";
import axios from "axios";

const AddNode = ({ btnName }) => {
	const [show, setShow] = useState(false);
	const [domains, setDomains] = useState({ });
	const [newdomain, updateNewdomain] = useState("");

	const dispatch = useDispatch();
	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);
	const rerenderSignal = useSelector((state) => state.rerender);

	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(alert({ show: false }));
			}, 5000);
		}
	}, [message, dispatch]);

	useEffect(() => {
		axios
			.get(process.env.REACT_APP_API_HOST + "user/" + loginState.jwt.Email, {
				headers: { token: loginState.token },
			})
			.then((response) => {
				setDomains(response.data.nodeinuse);
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	}, [rerenderSignal, loginState.jwt.Email, loginState.token, dispatch]);

	const handleAddNode = (e) => {
		e.preventDefault();

		axios({
			method: "put",
			url: process.env.REACT_APP_API_HOST + "addnode",
			headers: { token: loginState.token },
			data: domains,
		})
			.then((response) => {
				console.log(response);
				dispatch(success({ show: true, content: "Domain added in success!" }));
				dispatch(doRerender({rerender: !rerenderSignal.rerender}))
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	};

	return (
		<>
			<Button
				variant="outline-secondary"
				size="sm"
				onClick={() => setShow(!show)}
			>
				{btnName}
			</Button>

			<Modal
				show={show}
				onHide={() => setShow(false)}
				size="sm"
				aria-labelledby="contained-modal-title-vcenter"
				centered
			>
				<Modal.Header closeButton>
					<Modal.Title>Add Node</Modal.Title>
				</Modal.Header>
				<Modal.Body>
					<ListGroup>
						{Object.keys(domains).map((key) => (
							<ListGroup.Item
								className="d-inline-flex flex-row justify-content-between"
								eventKey={domains[key]}
							>
								<span className="">
									{" "}
									{key} {":"} {domains[key]}{" "}
								</span>
								<i
									className="bi bi-x-lg"
									onClick={() => {
										delete domains[key];
										setDomains({ ...domains });
									}}
								></i>
							</ListGroup.Item>
						))}
					</ListGroup>
					<Form id="addNodeForm" onSubmit={handleAddNode}>
						<InputGroup className="mb-3">
							<FormControl
								placeholder="New Domain"
								onChange={(e) => updateNewdomain(e.target.value)}
								value={newdomain}
							/>
							<Button
								variant="outline-secondary"
								onClick={() => {
									if (newdomain.length > 0) {
										setDomains((prevState) => ({
											...prevState,
											[newdomain.split(".")[0]]: newdomain,
										}));
									}
									updateNewdomain("");
								}}
							>
								Add Domain
							</Button>
						</InputGroup>
					</Form>
					<Modal.Footer>
						<Button variant="primary" type="submit" form="addNodeForm">
							Add Node
						</Button>
					</Modal.Footer>
				</Modal.Body>
			</Modal>
		</>
	);
};

export default AddNode;
