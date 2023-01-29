import React from "react";
import { useSelector } from "react-redux";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { LoginState } from "./types";
import "./App.css";
import Login from "./components/login";
import Home from "./components/home";
import Mypanel from "./components/mypanel";
import Menu from "./components/menu";
import Footer from "./components/footer";
import Macos from "./components/macos";
import Windows from "./components/windows";
import Iphone from "./components/iphone";
import Android from "./components/android";

interface Props {
	children: React.ReactNode;
}

function RequireAuth({ children }: Props): JSX.Element {
	const loginState: LoginState = useSelector((state: any) => state.login);

	if (loginState.isLogin === true) {
		return (children as JSX.Element);
	}
	return <Navigate to="/login" replace />
}

const App: React.FC = () => {
	return (
		<BrowserRouter>
			<Routes>
				<Route path="/" element={<Login />}></Route>
				<Route path="/login" element={<Login />}></Route>
				<Route
					path="/mypanel"
					element={
						<RequireAuth>
							<div
								className="flex-1 flex flex-col md:container md:mx-auto"
							>
								<Menu />
								<Mypanel />
								<Footer />
							</div>
						</RequireAuth>
					}
				></Route>
				<Route
					path="/home"
					element={
						<RequireAuth>
							<div
								className="flex-1 flex flex-col md:container md:mx-auto"
							>
								<Menu />
								<Home />
								<Footer />
							</div>
						</RequireAuth>
					}
				></Route>
				<Route
					path="/macos"
					element={
						<RequireAuth>
							<div
								className="flex-1 flex flex-col md:container md:mx-auto"
								// fluid="true"
							>
								<Menu />
								<Macos />
								<Footer />
							</div>
						</RequireAuth>
					}
				></Route>
				<Route
					path="/windows"
					element={
						<RequireAuth>
							<div
								className="flex-1 flex flex-col md:container md:mx-auto"
								// fluid="true"
							>
								<Menu />
								<Windows />
								<Footer />
							</div>
						</RequireAuth>
					}
				></Route>
				<Route
					path="/iphone"
					element={
						<RequireAuth>
							<div
								className="flex-1 flex flex-col md:container md:mx-auto"
								// fluid="true"
							>
								<Menu />
								<Iphone />
								<Footer />
							</div>
						</RequireAuth>
					}
				></Route>
				<Route
					path="/android"
					element={
						<RequireAuth>
							<div
								className="flex-1 flex flex-col md:container md:mx-auto"
								// fluid="true"
							>
								<Menu />
								<Android />
								<Footer />
							</div>
						</RequireAuth>
					}
				></Route>
			</Routes>
		</BrowserRouter>
	);
};

export default App;

// import './App.css';
// import HelloWorld from './helloworld';

// function App() {
//   return (
//     <div className="App">
//       <HelloWorld />
//     </div>
//   );
// }

// export default App;
