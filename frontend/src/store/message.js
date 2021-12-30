import { createSlice } from "@reduxjs/toolkit";

export const messageSlice = createSlice({
	name: "message",
	initialState: {
		show: false,
		content: "",
	},
	reducers: {
		alert: (state, action) => {
			return action.payload;
		},
	},
});

export const { alert } = messageSlice.actions;
export default messageSlice.reducer;
