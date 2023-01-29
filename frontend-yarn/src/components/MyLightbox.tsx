import { FunctionComponent, useState } from "react";
import Lightbox from "react-image-lightbox";
import 'react-image-lightbox/style.css';

interface MyLightboxProps {
	images: string[];
    device: string;

}

const MyLightbox: FunctionComponent<MyLightboxProps> = (props) => {
	const [isOpen, setIsOpen] = useState(false);
	const [photoIndex, setPhotoIndex] = useState(0);

	return (
		<div className="px-10">
			{props.device === "desktop" && (
				<img
					src={props.images[photoIndex]}
					alt=""
					onClick={() => setIsOpen(true)}
					className="bg-center bg-cover"
					style={{ width: "560px", height: "100%" }}
				/>
			)}
			{props.device === "mobile" && (
				<img
					src={props.images[photoIndex]}
					alt=""
					onClick={() => setIsOpen(true)}
					className="bg-center bg-cover"
					style={{ width: "300px", height: "100%" }}
				/>
			)}

			{isOpen && (
				<Lightbox
					mainSrc={props.images[photoIndex]}
					nextSrc={props.images[(photoIndex + 1) % props.images.length]}
					prevSrc={
						props.images[
							(photoIndex + props.images.length - 1) % props.images.length
						]
					}
					onCloseRequest={() => setIsOpen(false)}
					onMovePrevRequest={() =>
						setPhotoIndex(
							(photoIndex + props.images.length - 1) % props.images.length
						)
					}
					onMoveNextRequest={() =>
						setPhotoIndex((photoIndex + 1) % props.images.length)
					}
				/>
			)}
		</div>
	);
};

export default MyLightbox;
