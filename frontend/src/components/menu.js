import { useSelector, useDispatch } from "react-redux";
import { Container, Navbar, Nav, Button } from "react-bootstrap";
import { logout } from "../store/login";
import AddUser from "./adduser";
import AddNode from "./addNode";

const Menu = () => {
	const loginState = useSelector((state) => state.login);
	const dispatch = useDispatch();

	const handleLogout = (e) => {
		dispatch(logout());
	};

	return (
		<Navbar bg="dark" variant="dark" expand="lg">
			<Container>
				<Navbar.Brand href="/">Logv2ray Frontend</Navbar.Brand>
				<Navbar.Toggle aria-controls="basic-navbar-nav" />
				<Navbar.Collapse id="basic-navbar-nav">
					{loginState.jwt.Role === "admin" && (
						<Nav>
							<Nav.Link href="/home">User Management Panel</Nav.Link>
						</Nav>
					)}
					<Nav>
						<Nav.Link href="/mypanel">My Panel</Nav.Link>
					</Nav>
					<Nav>
						<Nav.Link href="/macos">MacOS</Nav.Link>
					</Nav>
					<Nav>
						<Nav.Link href="/windows">Windows</Nav.Link>
					</Nav>
					<Nav>
						<Nav.Link href="/iphone">IPhone/IPad</Nav.Link>
					</Nav>
					<Nav>
						<Nav.Link href="/android">Android</Nav.Link>
					</Nav>
				</Navbar.Collapse>
				<Navbar.Collapse className="justify-content-end">
					{loginState.jwt.Role === "admin" && (
						<>
							<Nav className="mx-2">
									<AddNode btnName="添加节点" />
							</Nav>
							<Nav className="mx-2">
									<AddUser btnName="添加用户" />
							</Nav>
						</>
					)}
					<Nav>
						<Navbar.Text className="mx-2">
							Signed in as: <b>{loginState.jwt.Email}</b>,
						</Navbar.Text>
					</Nav>
					<Nav>
							<Button
								variant="outline-secondary"
								size="sm"
								onClick={handleLogout}
							>
								logout
							</Button>
					</Nav>
				</Navbar.Collapse>
			</Container>
		</Navbar>
	);
};

export default Menu;
