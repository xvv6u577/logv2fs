import { createSlice } from "@reduxjs/toolkit";

const token = JSON.parse(localStorage.getItem("token"));

interface LoginState {
	isLogin: boolean;
jwt: {};
token: string;
}

function jwtVerify(token): LoginState {

	if (token === null) {
		return {isLogin: false, jwt: {}, token: ""}
	}

	const decodedJWT = JSON.parse(atob(token.split(".")[1]))
	if (decodedJWT.exp * 1000 < Date.now()){
		return {isLogin: false, jwt: decodedJWT, token}
	}

	return {isLogin: true, jwt: decodedJWT, token}
}

const initialState = jwtVerify(token)

export const loginSlice = createSlice({
	name: "login",
	initialState,
	reducers: {
		login: (state, action) => { 
			return jwtVerify(action.payload.token)

		},
		logout: (state, action) => {
			
			localStorage.removeItem("token");
			return {
				isLogin: false,
				jwt: {}
			}
		}
	}
});

export const { login, logout } = loginSlice.actions;
export default loginSlice.reducer;
