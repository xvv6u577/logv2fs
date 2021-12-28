import React from "react";
import { BrowserRouter, Routes, Route, Link } from "react-router-dom";
import "./App.css";
import Login from "./components/login";
import Home from "./components/home";

function App() {
	return (
		<BrowserRouter>
			<Routes>
				<Route path="/" element={<Login />}></Route>
				<Route path="/home" element={<Home />}></Route>
				{/* <Route path="/login" element={<Login />}></Route> */}
			</Routes>
		</BrowserRouter>
	);
}

export default App;
