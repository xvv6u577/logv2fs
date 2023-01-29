import { createSlice } from "@reduxjs/toolkit";
import { LoginState } from "../types";

const token: any = JSON.parse(localStorage.getItem("token") as string);

function jwtVerify(token: string | null): LoginState {
	if (token === null) {
		return { isLogin: false, jwt: {}, token: "" };
	}

	const decodedJWT = JSON.parse(atob(token.split(".")[1]));
	if (decodedJWT.exp * 1000 < Date.now()) {
		return { isLogin: false, jwt: decodedJWT, token };
	}

	return { isLogin: true, jwt: decodedJWT, token };
}

const initialState = jwtVerify(token);

export const loginSlice = createSlice({
	name: "login",
	initialState,
	reducers: {
		login: (state, action: { payload: { token: string } }) => {
			return jwtVerify(action.payload.token);
		},
		logout: (state) => {
			localStorage.removeItem("token");
			return {
				isLogin: false,
				jwt: {},
				token: "",
			};
		},
	},
});

export const { login, logout } = loginSlice.actions;
export default loginSlice.reducer;
