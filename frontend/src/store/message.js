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
		}
	},
});

export const { info, success, alert } = messageSlice.actions;
export default messageSlice.reducer;
