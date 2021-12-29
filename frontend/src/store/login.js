import { createSlice } from "@reduxjs/toolkit";

const token = JSON.parse(localStorage.getItem("token"));

export const loginSlice = createSlice({
	name: "login",
	initialState: {
		isLogin: token ? true : false,
		token: token ? token : "",
	},
	reducers: {
		login: (state, action) => { 
			localStorage.setItem("token", JSON.stringify(action.payload.token));
			return {
				isLogin: true,
				...action.payload
			} 
		},
		logout: (state, action) => {
			localStorage.removeItem("token");
			return {
				isLogin: false,
				...action.payload
			}
		}
	}
});

export const { login, logout } = loginSlice.actions;
export default loginSlice.reducer;
