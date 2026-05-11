import React from 'react';

/**
 * ErrorBoundary 捕获子树中抛出的渲染期错误，避免整页白屏。
 *
 * 限制：React Error Boundary 只能捕获渲染、生命周期、构造函数中的错误，
 * 不能捕获事件回调、setTimeout、Promise 中的异常——那些应在调用点 try/catch。
 */
class ErrorBoundary extends React.Component {
	constructor(props) {
		super(props);
		this.state = { hasError: false, error: null };
	}

	static getDerivedStateFromError(error) {
		return { hasError: true, error };
	}

	componentDidCatch(error, info) {
		console.error('UI ErrorBoundary caught:', error, info);
	}

	handleReload = () => {
		window.location.reload();
	};

	render() {
		if (!this.state.hasError) {
			return this.props.children;
		}

		return (
			<div className="min-h-screen bg-gray-900 text-gray-100 flex items-center justify-center p-6">
				<div className="max-w-md w-full bg-gray-800 border border-gray-700 rounded-xl shadow-xl p-6 text-center">
					<div className="w-14 h-14 mx-auto mb-4 rounded-full bg-red-500/20 flex items-center justify-center">
						<svg className="w-7 h-7 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
								d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<h2 className="text-xl font-semibold mb-2">页面出错了</h2>
					<p className="text-gray-400 text-sm mb-6">
						应用遇到了一个意外错误。您可以尝试刷新页面，问题持续请联系管理员。
					</p>
					<button
						onClick={this.handleReload}
						className="px-5 py-2.5 bg-blue-600 hover:bg-blue-700 rounded-lg text-white font-medium transition"
					>
						刷新页面
					</button>
				</div>
			</div>
		);
	}
}

export default ErrorBoundary;
