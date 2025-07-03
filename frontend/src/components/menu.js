import { useSelector, useDispatch } from "react-redux";
import { logout } from "../store/login";
import AddUser from "./adduser";

const Menu = () => {
	const loginState = useSelector((state) => state.login);
	const dispatch = useDispatch();

	const handleLogout = (e) => {
		dispatch(logout());
	};

	return (
		<header className="text-gray-400 bg-gray-900 body-font">
			<div className="mx-auto flex flex-wrap p-5 flex-col md:flex-row items-center">
				<a href="/home" className="flex title-font font-medium items-center text-white mb-4 md:mb-0">
					<img alt="logo" src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADIAAAAyCAYAAAAeP4ixAAAABmJLR0QA/wD/AP+gvaeTAAAHO0lEQVRoge2ZaXBVZxnHf8+5Wxay3NwEcpNQmBQKobbUUkEB0aKUQF06dewIOi1OQcr4yWGEqR1tnDqVqrUzdhxLodUW6Qe0H5zK1oL5gAPTaVh0WMMIZUkCyV2y3IS7nscPJJHk3nPuAtUP5j9zZ+553//7P8//vNtz3gMTmMAE/i8g/+sAsqH3YN2DhuoXUyqt3mWdx614d9yIaovRFfzcPWA2KVKPYZYCYBqDgnaAccbvO9Iu0mLa6fR94N+thvxYVD8EnCAJ05T53kc6TnxiRlTbXJ3BwEoR/TawFPBlaRJUOIjJzrrqmr0iDyUykfoP1G5U5Fe3BLux/Mtdv87EvS0jp3SXuzI4ab2I/BCYWpiKXlY1flnnq9463lD4g7pPG8KHoC6QhGnIZ7xLO/6RSaVgI52hfZ9H2QraVKjGOJxGWF9XteLvtxaG369/wBDzYdNh/M3KBBRgRFWlK7j3R4j8FHAUELAdkig/8fuat4iI5tMwLyOquxxdobKtwNN5hZc/tvurBp4ReSI1UvBmy7qvCLJThNVrnn999/gGRq7KqipdobLX+ORNAKztCpf9XlVHH7Qx3EOqmXsq5x7pDOx5DpGfjS9XVc4cPc/F9msUl7qZt+ReKrxlljp58YVn66pWbMklvpyM3JzY2sq4OZGIJ3hl0w5KvU14p9xD7MYAV9pbeeyphdz/2VlpOuP50aEIV9sP8tiaRRn5QFIxvlDvW344W4zObIRTustNiK3jTQAcOXACb91Cps9ZNFpW1ziXd7f/PGNgmfj1d99vyQecgrlNte0Bq71mBFnnSGVw0nqrJbb7ai/lvvoxZeGeK7jcpRm18uUPY861cM+6bHHaGlFtcw1vdjkjlYjicLqJReNpddHBGA6nJ42fDapsUm1z2XFsjXQFux+lgB3bU1LOQDiSVt4XjlBcWpGvHMC06+HgcjuCfY+IrCrkrm5PGQP9g2nlQ5EoLk9xzjpF7locRhEAppqr7biWRlRbDLmZAOaFCl8DJRUNXG6/llYXi5mIjF0oq+szTnIAJhU34nJWjlw+otpiGa9lRVfwoVlAtX3Y6Yj0B5g8tYmTbRfHlHd3BCgt96fx4zcGLLUCfYeJxkcfiK8zsGCmFddmaDlm20ZsgVQiiqd4EuGeQZLJ0QyDo4fO4Ku7L43fH+rMR96y+6yHFtqQzx3Go7ZxAQfePQLAjcEoH7WepXZ6upF8YIhYLjyWG6JglEFeCegY3DVrMcdb3+Dy+T9zvSPIrPmriEcjeIrLC9ZU1LJx1p0doHLSXDyuakIDR0kke225Fb6bHSkiPLh0LYN9Ae76VDkOp5tg14XbMmIHm6FF/8h/j8uHIS5czuxBRPoDY65LK6pxON0MhK7h8zem8UvLa3IOVpB+qzqboaUdI/9DA0dxOSq4Ebua9WapRBQzlQTAcNyUj/R1k0olSCZiOF0eu+a2MFWvWNXZrFqpsyP/Esk+hmKXUWwPPkZxovVNjvz1ZZKJKNcvnSYRixHp7WD39o1EervHcAf7exDDQSqVk/Y5qwpLI35f2zkgYFVvh0Q8wjfWLuaj/a9iqkmg8yThrsN8bc0y+gLpD7WsqoFzxy9kk+2uq25ut6rMaERb1xQlj539U/XHOytLei3f921x34LZbHj+61RXnWf+4iJ+sOVJFjfP4+PT7xOP/id9cbiKmDF3JXt2neWFDa8xNHjDSvKg3Xt8xjmSKIt9R5DHhRQl/WcYqpxbkJkav4+Vq5eMXpeUFfP0pq/yzm9/RyrpobLmbjwlNcSjg9TPWMSFf+4ndL2Xksb0fExEdtrdK/NkN0TREfNqkte7vX391Bl+Nr/yXaJDMS6euUyoJ8RQJMZQJM6yx+fQ0JiexgCXar2e/Xa6GY24ysJ/TPZVNiMsj3umHAJWZuI5XAaxaGz0emgghNPpRjX7xC0q8dA0zzJ1GgMRfUnk4aQdJ6MRmbk3BnwTQLXV2R+KHgPS8ot5i2az4zf78U6ZhstdhJomqVQcb03uqXoOOFXrnbw9GynHw4e9i1FayWD87LF/8Ze3DxGPmTgcBvXTvTyxYTnFpUUFxJyGpAlLGnwrjmQj5nEctPdZhBdvL678oCKb66uaf5ELN+dJ7Pc1bwGydvGdg2ybcnHH6cTRVX3Jtm89mo2dsxERUX/VwDP8V8zINn9V/wY0dXMNVCNrGl7YIXZ432aUF8gxe84DSRV5LtfhdCsK/qzQEdy/UDBfB+4tVGMs9KSJfC+XiZ0Jt/WhR7XNdS3cs06VTcC0AmUuiehLtd7J27OdJtohLyNvvfh9XyqReNLU1FtrW94IjZSrtrk6w93NhspqVb6EYP+SofQgHDDEeGeK170v22aXC/Ia42Y8+ZTAy4Y4AF4ZKR9+ku8B76mqdAT2zBRxNBmG1ptoGYCBDJgprirm2frqlefz/ZBzR42kSP7BgcM0NfW2FWc4wPbh3wQmMIEJ3Fn8G2f3w23K+XLVAAAAAElFTkSuQmCC"></img>
					<span className="ml-3 text-xl">Logv2fs Frontend</span>
				</a>
				<nav className="md:mr-auto md:ml-4 md:py-1 md:pl-4 md:border-l md:border-gray-700	flex flex-wrap items-center text-base justify-center">
					{loginState.jwt.Role === "admin" && (
						<>
							<a className="mr-5 hover:text-white" href="/home">User</a>
							<a className="mr-5 hover:text-white" href="/nodes">Nodes</a>
							<a className="mr-5 hover:text-white" href="/addnode">Add Node</a>
							<a className="mr-5 hover:text-white" href="/paymentinput">Add Payment</a>
							<a className="mr-5 hover:text-white" href="/paymentstatistics">Payment Stats</a>
						</>
					)}
					<a className="mr-5 hover:text-white" href="/mypanel">My Panel</a>
					<a className="mr-5 hover:text-white" href="/macos">MacOS</a>
					<a className="mr-5 hover:text-white" href="/windows">Windows</a>
					<a className="mr-5 hover:text-white" href="/iphone">IPhone/IPad</a>
					<a className="mr-5 hover:text-white" href="/android">Android</a>
				</nav>
				{loginState.jwt.Role === "admin" && (
					<>
						{/* <button
							className="w-full sm:w-auto block text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 
								font-medium rounded-lg text-sm px-1.5 py-1 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
							onClick={handleWriteToDB}
						>
							<svg xmlns="http://www.w3.org/2000/svg" className="inline-block h-4 w-4 mr-1" viewBox="0 0 20 20" fill="currentColor">
								<path fillRule="evenodd" d="M8 4a3 3 0 00-3 3v4a5 5 0 0010 0V7a1 1 0 112 0v4a7 7 0 11-14 0V7a5 5 0 0110 0v4a3 3 0 11-6 0V7a1 1 0 012 0v4a1 1 0 102 0V7a3 3 0 00-3-3z" clipRule="evenodd" />
							</svg>
							WriteToDB
						</button> */}
						<AddUser btnName="AddUser" />
					</>
				)}
				<span className="hover:text-white" href="#">Signed in as: <b>{loginState.jwt.Email}</b></span>
				<button
					className="w-full sm:w-auto block text-white-900 bg-white hover:bg-gray-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-1.5 py-1 m-1 text-center dark:bg-gray-600 dark:hover:bg-gray-800 dark:focus:ring-gray-800"
					// className="w-full sm:w-auto text-gray-900 bg-white hover:bg-gray-100 border border-gray-200 focus:ring-4 focus:outline-none focus:ring-gray-100 font-medium rounded-md text-sm px-3 py-2.5 mx-2 text-center inline-flex items-center dark:focus:ring-gray-600 dark:bg-gray-800 dark:border-gray-700 dark:text-white dark:hover:bg-gray-700"
					onClick={handleLogout}
				>
					<svg fill="none" className="inline-block h-4 w-4 mr-1" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24">
						<path d="M5 12h14M12 5l7 7-7 7"></path>
					</svg>
					logout
				</button>
			</div>

		</header>
	);
};

export default Menu;
