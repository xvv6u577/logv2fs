import { createSlice } from "@reduxjs/toolkit";

export const rerenderSlice = createSlice({
	name: "rerender page",
	initialState: {
        rerender: false
	},
	reducers: {
		doRerender: (state, action) => {
			return {...action.payload};
		},
		
	},
});

export const { doRerender } = rerenderSlice.actions;
export default rerenderSlice.reducer;
