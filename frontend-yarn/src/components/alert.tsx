import React from "react";

type AlertProps = {
	message: string | null;
	type: string;
	shown: boolean;
	close: () => void;
};

const Alert: React.FC<AlertProps> = ({ message, type, shown, close }) => {
	if (message === null && shown === false) {
		return null;
	}
	return (
		<>
			<div
				id="popup-modal"
				className={`${
					shown ? "" : "hidden"
				} overflow-y-auto overflow-x-hidden fixed top-0 right-0 left-0 z-50 md:inset-0 h-modal md:h-full justify-center items-center flex`}
				onClick={() => {
					close();
				}}
			>
				<div
					className="relative p-4 w-full max-w-md h-full md:h-auto"
					onClick={(e) => {
						e.stopPropagation();
					}}
				>
					<div className="relative bg-white rounded-lg shadow dark:bg-gray-700">
						<button
							type="button"
							onClick={() => {
								close();
							}}
							className="absolute top-3 right-2.5 text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-auto inline-flex items-center dark:hover:bg-gray-800 dark:hover:text-white"
						>
							<svg
								aria-hidden="true"
								className="w-5 h-5"
								fill="currentColor"
								viewBox="0 0 20 20"
								xmlns="http://www.w3.org/2000/svg"
							>
								<path
									fillRule="evenodd"
									d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
									clipRule="evenodd"
								></path>
							</svg>
							<span className="sr-only">Close modal</span>
						</button>
						<div className="p-6 text-center">
							{type === "info" && (
								<svg
									aria-hidden="true"
									className="mx-auto mb-4 w-14 h-14 text-gray-400 dark:text-gray-400"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
									xmlns="http://www.w3.org/2000/svg"
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										strokeWidth="2"
										d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
									></path>
								</svg>
							)}
							{type === "warning" && (
								<svg
									aria-hidden="true"
									className="mx-auto mb-4 w-14 h-14 text-orange-500 dark:text-orange-500"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
									xmlns="http://www.w3.org/2000/svg"
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										strokeWidth="2"
										d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
									></path>
								</svg>
							)}
							{type === "success" && (
								<svg
									aria-hidden="true"
									className="mx-auto mb-4 w-14 h-14 text-green-500 dark:text-green-500"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
									xmlns="http://www.w3.org/2000/svg"
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										strokeWidth="2"
										d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
									></path>
								</svg>
							)}
							<h3 className="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">
								{message}
							</h3>
						</div>
					</div>
				</div>
			</div>
		</>
	);
};

export default Alert;
