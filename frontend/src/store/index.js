import { configureStore } from '@reduxjs/toolkit';
import loginSlice from '../store/login';
import messageSlice from '../store/message';

/**
 * Redux Store 配置
 * 
 * 保留的状态：
 * - login: 用户认证状态（JWT、token）
 * - message: 全局提示消息
 * 
 * 移除的状态：
 * - rerender: 已由 React Query 的 invalidateQueries 替代
 */
export const store = configureStore({
  reducer: {
    login: loginSlice,
    message: messageSlice,
  },
});
