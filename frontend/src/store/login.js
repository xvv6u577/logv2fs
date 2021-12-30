import { createSlice } from "@reduxjs/toolkit";

const token = JSON.parse(localStorage.getItem("token"));

function jwtVerify(token) {

	if (token === null) {
		return {isLogin: false, jwt: {}}
	}

	const decodedJWT = JSON.parse(atob(token.split(".")[1]))
	if (decodedJWT.exp * 1000 < Date.now()){
		return {isLogin: false, jwt: decodedJWT}
	}

	return {isLogin: true, jwt: decodedJWT}
}

export const loginSlice = createSlice({
	name: "login",
	initialState: jwtVerify(token),
	reducers: {
		login: (state, action) => { 

			localStorage.setItem("token", JSON.stringify(action.payload.token));
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
