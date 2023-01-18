import { useState } from "react";

const TapToCopied = (props) => {
	const [showTooltip, setShowTooltip] = useState(false);

	return (
		<span className="me-3 my-1 relative">
			<span className="truncate md:inline">
				{props.children}
			</span>
			{" | "}
			<button 
				type="button" 
				onClick={() => {
					navigator.clipboard.writeText(props.children);
					setShowTooltip(true);
					setTimeout(()=>{
						setShowTooltip(false);
					}, 3000);
				}}
				className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-3 py-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
			>
				Copy
			</button>
			<div 
				id="tooltip-click" 
				role="tooltip" 
				className={`${showTooltip? "visible opacity-1" : "invisible opacity-0"} absolute inline-block z-10 py-2 px-3 mx-2 text-sm font-medium text-white bg-gray-900 rounded-lg shadow-sm tooltip dark:bg-gray-700`}>
				Copied!
				<div 
					className="tooltip-arrow"
					style={{position: "absolute", left: "-61px",bottom:"18px", transform: "translate(59px, 0px)"}}
				></div>
			</div>
		</span>
	);
};

export default TapToCopied;
