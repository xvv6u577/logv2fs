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
// export const show = (state) => state.message.show;
// export const content = (state) => state.message.content;

export default messageSlice.reducer;
