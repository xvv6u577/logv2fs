import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";

export const loginSlice = createSlice({
	name: "login",
	initialState: {
		isLogin: false,
		user: {},
	},
	reducers: {
		login: (state, action) => { 
			return action.payload 
		},
		redirect: (state, action) => {},
	},
});

export const { login } = loginSlice.actions;
// export const user = (state) => state.login.user;
// export const isLogin = (state) => state.login.isLogin;

export default loginSlice.reducer;
