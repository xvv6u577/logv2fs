import React from "react";
import { useSelector } from "react-redux";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Container } from "react-bootstrap";
import "./App.css";
import Login from "./components/login";
import Home from "./components/home";
import Menu from "./components/menu";
import Macos from "./components/macos";
import Windows from "./components/windows";
import Android from "./components/android";
import Iphone from "./components/iphone";
import Footer from "./components/footer";

function RequireAuth({ children }) {
	const loginState = useSelector((state) => state.login);

	return loginState.isLogin === true ? (
		children
	) : (
		<Navigate to="/login" replace />
	);
}

function App() {
	return (
		<BrowserRouter>
			<Routes>
				<Route path="/" element={<Login />}></Route>
				<Route path="/login" element={<Login />}></Route>
				<Route
					path="/home"
					element={
						<RequireAuth>
							<Container className="main px-0" fluid>
								<Menu />
								<Home />
								<Footer />
							</Container>
						</RequireAuth>
					}
				></Route>
				<Route
					path="/macos"
					element={
						<RequireAuth>
							<Container className="main px-0" fluid>
								<Menu />
								<Macos />
								<Footer />
							</Container>
						</RequireAuth>
					}
				></Route>
				<Route
					path="/windows"
					element={
						<RequireAuth>
							<Container className="main px-0" fluid>
								<Menu />
								<Windows />
								<Footer />
							</Container>
						</RequireAuth>
					}
				></Route>
				<Route
					path="/iphone"
					element={
						<RequireAuth>
							<Container className="main px-0" fluid>
								<Menu />
								<Iphone />
								<Footer />
							</Container>
						</RequireAuth>
					}
				></Route>
				<Route
					path="/android"
					element={
						<RequireAuth>
							<Container className="main px-0" fluid>
								<Menu />
								<Android />
								<Footer />
							</Container>
						</RequireAuth>
					}
				></Route>
			</Routes>
		</BrowserRouter>
	);
}

export default App;
