import { v4 as uuidv4 } from 'uuid';

const TapToCopied = (props) => {
	const id = uuidv4();

	return (
		<div className="me-3 inline-flex items-center">
			<div
				className="px-2 max-w-sm"
				style={{
					maxWidth: "230px",
					textOverflow: "ellipsis",
					whiteSpace: "nowrap",
					overflow: "hidden",
					display: "inline-block",
				}}
			>
				{props.children}
			</div>
			<span>
				{" | "}
			</span>
			<div className="inline-block mx-1">
				<button 
					data-tooltip-trigger="click"
					data-tooltip-target={"tooltip-left-"+id} 
					data-tooltip-placement="left" 
					type="button" 
					onClick={() => {
						navigator.clipboard.writeText(props.children);
						console.log("Copied to clipboard: " + props.children, "tooltip-left-"+id);
					}}
					className="mb-2 md:mb-0 text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded text-sm px-3 py-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
				>
					Copy
				</button>
				<div 
					id={"tooltip-left-"+id} 
					role="tooltip" 
					className="absolute z-10 invisible inline-block px-3 py-2 text-sm font-medium text-white bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700">
					Copied!
					<div className="tooltip-arrow" data-popper-arrow></div>
				</div>
			</div>
		</div>
	);
};

export default TapToCopied;
