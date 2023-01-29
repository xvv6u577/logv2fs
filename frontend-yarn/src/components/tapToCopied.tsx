import React, { useState } from "react";
import "../style/tapToCopied.css";

interface Props {
	children: React.ReactNode;
  }

  const TapToCopied: React.FC<Props> = (props) => {
	const [tooltipVisible, setTooltipVisible] = useState(false);
	const text = props.children?.toString() as string;

	return (
		<div className="me-3 inline-flex items-center">
			<div
				className="px-2"
				id="tap-to-copy"
				style={{
					textOverflow: "ellipsis",
					whiteSpace: "nowrap",
					overflow: "hidden",
					display: "inline-block",
				}}
			>
				{text}
			</div>
			<span>
				{" | "}
			</span>
			<div className="inline-block mx-1 relative">
				<button 
					type="button" 
					onClick={() => {
						navigator.clipboard.writeText(text);
						setTooltipVisible(true);
						setTimeout(() => {
							setTooltipVisible(false);
						}, 3000);
					}}
					className="mb-2 md:mb-0 text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded text-sm px-3 py-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
				>
					Copy
				</button>
				{tooltipVisible && 
					<div 
						style={{
							bottom: "15%",
							right: "110%",
						}}
						className="absolute z-1 visible inline-block px-3 py-2 text-sm font-medium text-white bg-gray-900 rounded-lg shadow-sm opacity-1 tooltip dark:bg-gray-700">
						Copied!
					</div>
				}
			</div>
		</div>
	);
};

export default TapToCopied;
