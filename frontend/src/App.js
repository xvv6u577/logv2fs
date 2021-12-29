import React from "react";
import { useSelector } from "react-redux";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import "./App.css";
import Login from "./components/login";
import Home from "./components/home";

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
				<Route
					path="/home"
					element={
						<RequireAuth>
							<Home />
						</RequireAuth>
					}
				></Route>
				<Route path="/login" element={<Login />}></Route>
			</Routes>
		</BrowserRouter>
	);
}

export default App;
