import { useRef, useState } from "react";
import { Overlay, Tooltip, Badge } from "react-bootstrap";

const TapToCopied = (props) => {
	const [show, setShow] = useState(false);
	const target = useRef(null);

	return (
		<span className="tap-to-copied h6 me-3">
			<span ref={target} className="small copied-text text-break">
				{props.children}
			</span>
			{" | "}
			<Badge
				bg="info"
				className="tap-to-copied-badge"
				onClick={() => {
					navigator.clipboard.writeText(props.children);
					setShow(!show);
                    setTimeout(() => {
                        setShow(show);
                    }, 3000);
				}}
			>
				Copy
			</Badge>
			<Overlay target={target.current} show={show} placement="top">
				{(props) => (
					<Tooltip id="overlay-example" {...props}>
						Copied!
					</Tooltip>
				)}
			</Overlay>
		</span>
	);
};

export default TapToCopied;
