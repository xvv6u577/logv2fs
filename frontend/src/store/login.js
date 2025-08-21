import { createSlice } from "@reduxjs/toolkit";

const token = localStorage.getItem("token");

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

		// Ensure base64 string has correct padding
		let payload = parts[1];
		// Add necessary padding characters
		while (payload.length % 4 !== 0) {
			payload += '=';
		}

		// Replace URL-safe base64 characters with standard base64 characters
		const base64Payload = payload.replace(/-/g, '+').replace(/_/g, '/');
		let decodedJWT;
		try {
			decodedJWT = JSON.parse(atob(base64Payload));
		} catch (decodeError) {
			console.error('Base64 decode error:', decodeError);
			localStorage.removeItem("token");
			return {isLogin: false, jwt: {}, token: ""};
		}

		if (
			typeof decodedJWT.exp !== "number" ||
			isNaN(decodedJWT.exp) ||
			decodedJWT.exp * 1000 < Date.now()
		) {
			return {isLogin: false, jwt: decodedJWT, token}
		}

		return {isLogin: true, jwt: decodedJWT, token}
	} catch (error) {
		console.error('JWT decode error:', error);
		localStorage.removeItem("token");
		return {isLogin: false, jwt: {}, token: ""}
	}
}

const initialState = jwtVerify(token);

const loginSlice = createSlice({
	name: 'login',
	initialState,
	reducers: {
		login: (state, action) => { 
			const token = action.payload && action.payload.token ? action.payload.token : "";
			localStorage.setItem("token", token);
			return jwtVerify(token);
		},
		logout: (state, action) => {
			localStorage.removeItem("token");
			return {
				isLogin: false,
				jwt: {},
				token: ""
			};
		}
	}
});

export const { login, logout } = loginSlice.actions;
export default loginSlice.reducer;
