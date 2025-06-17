import React from "react";
import { useSelector } from "react-redux";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import "./App.css";
import Login from "./components/login";
import Home from "./components/home";
import Menu from "./components/menu";
import Macos from "./components/macos";
import Windows from "./components/windows";
import Iphone from "./components/iphone";
import Android from "./components/android";
import Footer from "./components/footer";
import Mypanel from "./components/mypanel";
import Nodes from "./components/nodes";
import AddNode from "./components/addNode";

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
				<Route path="/home" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<Home />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/login" element={<Login />} />
				<Route path="/mypanel" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<Mypanel />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/addnode" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<AddNode />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/nodes" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<Nodes />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/macos" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<Macos />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/windows" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<Windows />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/iphone" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<Iphone />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/android" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<Android />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/" element={<Login />} />
			</Routes>
		</BrowserRouter>
	);
}

export default App;
