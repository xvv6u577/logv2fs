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
							<div className="local_container" fluid="true">
								<Menu />
								<Home />
								<Footer />
							</div>
						</RequireAuth> } >
				</Route>
				<Route path="/login" element={ <Login /> }></Route>
				<Route path="/mypanel" element={ 
						<RequireAuth> <div className="local_container" fluid="true">
								<Menu />
								<Mypanel />
								<Footer />
							</div>
						</RequireAuth> } >
				</Route>
				<Route path="/addnode" element={
						<RequireAuth>
							<div className="local_container" fluid="true">
								<Menu />
								<AddNode />
								<Footer />
							</div>
						</RequireAuth>}>
				</Route>
				<Route path="/nodes" element={
						<RequireAuth> <div className="local_container" fluid="true">
								<Menu />
								<Nodes />
								<Footer />
							</div>
						</RequireAuth> } >
				</Route>
				<Route path="/macos" element={
						<RequireAuth>
							<div className="local_container" fluid="true">
								<Menu />
								<Macos />
								<Footer />
							</div>
						</RequireAuth> } >
				</Route>
				<Route path="/windows" element={
						 <RequireAuth>
							<div className="local_container" fluid="true">
								<Menu />
								<Windows />
								<Footer />
							</div>
						</RequireAuth> } >
						</Route>
				<Route path="/iphone" element={
						<RequireAuth>
							<div className="local_container" fluid="true">
								<Menu />
								<Iphone />
								<Footer />
							</div>
						</RequireAuth> } >
				</Route>
				<Route path="/android" element={
						<RequireAuth>
							<div className="local_container" fluid="true">
								<Menu />
								<Android />
								<Footer />
							</div>
						</RequireAuth> } >
				</Route>
				<Route path="/" element={ <Login /> }></Route>
			</Routes>
		</BrowserRouter>
	);
}

export default App;
