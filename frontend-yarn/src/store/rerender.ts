import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { RerenderState } from "../types";

const initialState: RerenderState = {
	rerender: false,
};

export const rerenderSlice = createSlice({
	name: "rerender page",
	initialState,
	reducers: {
		doRerender: (state, action: PayloadAction<RerenderState>) => {
			state.rerender = action.payload.rerender;
		},
	},
});

export const { doRerender } = rerenderSlice.actions;
export default rerenderSlice.reducer;
