import { createSlice } from "@reduxjs/toolkit";

const token = JSON.parse(localStorage.getItem("token"));

function jwtVerify(token) {

	if (token === null || token === undefined || typeof token !== 'string') {
		return {isLogin: false, jwt: {}, token: ""}
	}

	try {
		const parts = token.split(".");
		if (parts.length !== 3) {
			console.warn('Invalid JWT format');
			return {isLogin: false, jwt: {}, token: ""}
		}

		// 确保 base64 字符串有正确的填充
		let payload = parts[1];
		// 添加必要的填充字符
		while (payload.length % 4 !== 0) {
			payload += '=';
		}

		const decodedJWT = JSON.parse(atob(payload))
		if (decodedJWT.exp * 1000 < Date.now()){
			return {isLogin: false, jwt: decodedJWT, token}
		}

		return {isLogin: true, jwt: decodedJWT, token}
	} catch (error) {
		console.error('JWT decode error:', error);
		localStorage.removeItem("token");
		return {isLogin: false, jwt: {}, token: ""}
	}
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
