import { configureStore } from '@reduxjs/toolkit';
import loginSlice from '../store/login';
import messageSlice from '../store/message';
import rerenderSlice from '../store/rerender';

export const store = configureStore({
  reducer: {
    login: loginSlice,
    message: messageSlice,
    rerender: rerenderSlice,
  },
});
