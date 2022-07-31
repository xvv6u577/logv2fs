import { createSlice } from "@reduxjs/toolkit";

export const messageSlice = createSlice({
	name: "message",
	initialState: {
		show: false,
		type: "",
		content: "",
	},
	reducers: {
		info: (state, action) => {
			return {...action.payload, type: "info"};
		},
		alert: (state, action) => {
			return {...action.payload, type: "warning"};
		},
		success: (state, action) => {
			return {...action.payload, type: "success"};
		},
		reset: (state) => {
			return {show: false, type: "", content: ""};
		}
	},
});

export const { info, success, alert, reset } = messageSlice.actions;
export default messageSlice.reducer;
