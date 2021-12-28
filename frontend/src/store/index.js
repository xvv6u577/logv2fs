import { configureStore } from '@reduxjs/toolkit';
import loginSlice from '../store/login';
import messageSlice from '../store/message';

export const store = configureStore({
  reducer: {
    login: loginSlice,
    message: messageSlice,
  },
});
