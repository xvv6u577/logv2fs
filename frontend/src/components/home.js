import { useEffect, useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, reset} from "../store/message";
import axios from "axios";
import { logout } from "../store/login";
import UserComp from "./userComp";
import Alert from "./alert";

const Home = () => {
	const [users, setUsers] = useState([]);
	const loginState = useSelector((state) => state.login);
	const message = useSelector((state) => state.message);
	const rerenderSignal = useSelector((state) => state.rerender);

	const dispatch = useDispatch();

	useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(reset({}));
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
				if (err.response) {
					dispatch(alert({ show: true, content: err.response.data.error }));
				} else {
					dispatch(alert({ show: true, content: err.toString() }));
				}
			});
	}, [rerenderSignal, loginState.jwt.Email, loginState.token, dispatch]);

	return (
		<div className="my-3">
			<Alert message={message.content} type={message.type} shown={message.show} close={() => { dispatch(reset({})); }}/>

			<div id="accordion-collapse" data-accordion="collapse">
				{users
					// put admin at the top of the list
					.reduce((acc, ele) => {
						if (ele.role === "admin") {
							return [ele, ...acc];
						}
						return [...acc, ele];
					}, [])
					// sort the normal users by used_by_current_month.amount
					.sort((a, b) => {
						if ((a.role === "admin") | (b.role === "admin")) {
							return 0;
						}
						return (
							b.used_by_current_month.amount - a.used_by_current_month.amount
						);
					})
					.map((element, index) => ( 
						<UserComp user={element} index={index} key={index} />
					))
				}
			</div>

			{/* <ListGroup
				as="ol"
				class="list-group list-group-striped list-group-hover"
				numbered
			>
				<Accordion>
					{users
						// put admin at the top of the list
						.reduce((acc, ele) => {
							if (ele.role === "admin") {
								return [ele, ...acc];
							}
							return [...acc, ele];
						}, [])
						// sort the normal users by used_by_current_month.amount
						.sort((a, b) => {
							if ((a.role === "admin") | (b.role === "admin")) {
								return 0;
							}
							return (
								b.used_by_current_month.amount - a.used_by_current_month.amount
							);
						})
						.map((element, index) => (
							<Accordion.Item eventKey={index} key={index}>
								<ListGroup.Item as="li" class="d-flex align-items-center" key={index}>
									<div class="me-auto " key={index}>
										<div class="home-traffic-fs">
											<span class="badge rounded-pill bg-secondary">
												{index + 1}
											</span>{" "}
											<span class="">{element.name}</span>
											{element.status === "plain" ? (
												<span class="badge rounded-pill bg-success mx-1">
													online
												</span>
											) : (
												<span class="badge rounded-pill bg-danger mx-1">
													offline
												</span>
											)}
											{element.role === "admin" ? (
												<span class="badge rounded-pill bg-dark mx-1">
													admin
												</span>
											) : (
												<span class="badge rounded-pill bg-primary mx-1">
													user
												</span>
											)}
											{element.email === loginState.jwt.Email && (
												<span class="badge rounded-pill bg-info text-dark">
													Me
												</span>
											)}
										</div>

										<div class="home-traffic-fs">
											Today: {formatBytes(element.used_by_current_day.amount)}
											{", "}
											This month:{" "}
											{formatBytes(element.used_by_current_month.amount)}
											{", "}
											Used: {formatBytes(element.used)}
										</div>
									</div>
									<div class="d-flex justify-content-center align-items-center" key={index}>
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
												variant="primary mx-1 btn-custom"
												size="sm"
											>
												Disable
											</Button>
										) : (
											<Button
												onClick={() => handleOnline(element.email)}
												variant="success mx-1 btn-custom"
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
									<div class="d-md-flex flex-row justify-content-between px-md-5">
										<div class="p-2 flex-fill px-md-5 border border-info border-3 rounded-3  m-1">
											<div class="d-flex justify-content-between py-1">
												<span class="">用户名:</span>{" "}
												<TapToCopied>{element.email}</TapToCopied>
											</div>
											<div class="d-flex justify-content-between py-1">
												<span class="">path: </span>
												<TapToCopied>{element.path}</TapToCopied>
											</div>
											<div class="d-md-flex justify-content-between py-1">
												<span class="">uuid: </span>
												<TapToCopied>{element.uuid}</TapToCopied>
											</div>
											<div class="d-md-flex justify-content-between py-1">
												<span class="">SubUrl:</span>
												<TapToCopied>
													{process.env.REACT_APP_FILE_AND_SUB_URL +
														"/static/" +
														element.email}
												</TapToCopied>
											</div>
										</div>

										<div class="p-2 flex-fill px-md-5 border border-info border-3 rounded-3 m-1">
											{element &&
												Object.entries(element.node_in_use_status).map(
													([key, value]) => (
														<div class="d-flex flex-row justify-content-between py-1">
															<span class="me-auto my-1">Node: {key}</span>
															{value ? (
																<Button
																	variant="primary btn-custom"
																	size="sm"
																	onClick={() =>
																		handleDisableNode({
																			email: element.email,
																			node: key,
																		})
																	}
																>
																	Disable
																</Button>
															) : (
																<Button
																	variant="success btn-custom"
																	size="sm"
																	onClick={() =>
																		handleEnableNode({
																			email: element.email,
																			node: key,
																		})
																	}
																>
																	Enable
																</Button>
															)}
														</div>
													)
												)}
										</div>
									</div>
									<div class="home-traffic-fs pt-2">
										<h4 class=" text-center">Monthly Traffic </h4>
										<TrafficTable data={element.traffic_by_month} by="月份" />
									</div>
									<div class="home-traffic-fs">
										<h4 class="pt-2 text-center">Daily Traffic</h4>
										<TrafficTable data={element.traffic_by_day} by="日期" />
									</div>
								</Accordion.Body>
							</Accordion.Item>
						))}
				</Accordion>
			</ListGroup> */}
		</div >
	);
};


export default Home;
