function Footer() {
	const year = new Date().getFullYear();

	return (
		<footer className="mt-auto py-3 bg-dark">
			<div className="flex content-center justify-center">
				<span className="text-muted text-light">
					Logv2 App <span>&copy;</span> {year}
				</span>
			</div>
		</footer>
	);
}

export default Footer;
