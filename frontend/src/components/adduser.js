import { useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, success } from "../store/message";
import { doRerender } from "../store/rerender";
import axios from "axios";

function AddUser({ btnName }) {

	const initialState = {
		email: "",
		name: "",
		path: "ray",
		role: "normal",
		uuid: ""
	};
	const [showModal, setShowModal] = useState(false);
	const [{ email, name, path, role, uuid }, setState] = useState(initialState);

	const dispatch = useDispatch();
	const loginState = useSelector((state) => state.login);
	const rerenderSignal = useSelector((state) => state.rerender);


	const clearState = () => {
		setState({ ...initialState });
	};

	const handleAddUser = (e) => {
		e.preventDefault();
		setShowModal(!showModal);

		axios({
			method: "post",
			url: process.env.REACT_APP_API_HOST + "signup",
			headers: { token: loginState.token },
			data: {
				email,
				"password":email,
				role,
				name,
				path,
				status: "plain",
				uuid
			},
		})
			.then((response) => {
				dispatch(success({ show: true, content: "user " + name + " added in success!" }));
				dispatch(doRerender({ rerender: !rerenderSignal.rerender }))
				clearState();
			})
			.catch((err) => {
				dispatch(alert({ show: true, content: err.toString() }));
			});
	};

	const onChange = (e) => {
		const { name, value } = e.target;
		setState((prevState) => ({ ...prevState, [name]: value }));
	};

	return (
		<>
			<button
				className="w-full sm:w-auto  block text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-1.5 py-1 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
				type="button"
				onClick={() => setShowModal(!showModal)}
			>
				<svg xmlns="http://www.w3.org/2000/svg" className="inline-block h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
					<path strokeLinecap="round" strokeLinejoin="round" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
				</svg>
				{btnName}
			</button>

			{showModal ?
				<div
					id="add-user-modal"
					className="overflow-y-auto overflow-x-hidden fixed top-0 right-0 left-0 z-50 w-full md:inset-0 h-modal md:h-full justify-center items-center flex" aria-modal="true"
					role="dialog"
				>
					<div className="relative p-4 w-full max-w-md h-full md:h-auto">
						<div className="relative bg-white rounded-lg shadow dark:bg-gray-700">
							<button
								type="button"
								className="absolute top-3 right-2.5 text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-auto inline-flex items-center dark:hover:bg-gray-800 dark:hover:text-white"
								onClick={() => setShowModal(!showModal)}
							>
								<svg aria-hidden="true" className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd"></path></svg>
								<span className="sr-only">Close modal</span>
							</button>
							<div className="py-6 px-6 lg:px-8">
								<h3 className="mb-4 text-xl font-medium text-gray-900 dark:text-white">Add User</h3>
								<form className="space-y-6" onSubmit={handleAddUser}>
									<div>
										<label htmlFor="email" className="block mb-2 text-sm font-medium text-gray-900 dark:text-gray-300">Email (4-100 characters)*</label>
										<input
											type="input"
											name="email"
											id="email"
											value={email}
											onChange={onChange}
											className="bg-gray-5ˀ0 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-600 dark:border-gray-500 dark:placeholder-gray-400 dark:text-white" 
											placeholder="email"
											required />
									</div>
									<div>
										<label htmlFor="name" className="block mb-2 text-sm font-medium text-gray-900 dark:text-gray-300">Name (optinal)</label>
										<input
											type="text"
											id="name"
											name="name"
											value={name}
											onChange={onChange}
											placeholder="name"
											className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-600 dark:border-gray-500 dark:placeholder-gray-400 dark:text-white"
										/>
									</div>
									<button type="submit" className="w-full text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800">Submit</button>

								</form>
							</div>
						</div>
					</div>
				</div>
				: null}
		</>
	);
}

export default AddUser