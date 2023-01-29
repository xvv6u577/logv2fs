import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { MessageState, MessagePayload } from "../types";

const initialState: MessageState = {
	show: false,
	type: "",
	content: "",
};

export const messageSlice = createSlice({
	name: "message",
	initialState,
	reducers: {
		info: (state, action: PayloadAction<MessagePayload>) => {
			return { ...action.payload, type: "info" };
		},
		alert: (state, action: PayloadAction<MessagePayload>) => {
			return { ...action.payload, type: "warning" };
		},
		success: (state, action: PayloadAction<MessagePayload>) => {
			return { ...action.payload, type: "success" };
		},
		reset: (state) => {
			return { show: false, type: "", content: "" };
		},
	},
});

export const { info, success, alert, reset } = messageSlice.actions;
export default messageSlice.reducer;
