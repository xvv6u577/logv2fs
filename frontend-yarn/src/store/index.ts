import { configureStore } from "@reduxjs/toolkit";
import loginSlice from "./login";
import messageSlice from "./message";
import rerenderSlice from "./rerender";
import { IStore } from "../types";

const store = configureStore<IStore>({
	reducer: {
		login: loginSlice,
		message: messageSlice,
		rerender: rerenderSlice,
	},
});

export default store;
