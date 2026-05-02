import React, { useState, useEffect, useRef } from "react";
import { useSelector } from "react-redux";
import { BrowserRouter, Routes, Route, Navigate, Link } from "react-router-dom";
import "./App.css";
import Login from "./components/login";
import User from "./components/user";
import Menu from "./components/menu";
import Macos from "./components/macos";
import Windows from "./components/windows";
import Iphone from "./components/iphone";
import Android from "./components/android";
import Footer from "./components/footer";
import Mypanel from "./components/mypanel";
import Nodes from "./components/nodes";
import AddNode from "./components/addNode";
import PaymentInput from "./components/paymentInput";
import PaymentStatistics from "./components/paymentStatistics";
import PaymentRecords from "./components/paymentRecords";
import {logoBase64} from "./components/logoImage"

const services = [
	{
		icon: (
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5} className="w-7 h-7">
				<rect x="2" y="3" width="20" height="14" rx="2" />
				<path d="M8 21h8M12 17v4" />
			</svg>
		),
		title: "Website Development",
		desc: "Custom WordPress & React applications built for performance. From landing pages to full-stack web platforms tailored to your brand.",
		tags: ["WordPress", "React", "TypeScript"],
	},
	{
		icon: (
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5} className="w-7 h-7">
				<path d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" />
			</svg>
		),
		title: "Cloud Deployment",
		desc: "Zero-downtime deployments, container orchestration, CI/CD pipelines. Your product ships faster, stays up longer.",
		tags: ["Docker", "Kubernetes", "GitHub Actions"],
	},
	{
		icon: (
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5} className="w-7 h-7">
				<path d="M5 12H3l9-9 9 9h-2M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7" />
				<path d="M9 21V12h6v9" />
			</svg>
		),
		title: "Infrastructure Optimization",
		desc: "Cut cloud costs, harden security, and scale with confidence. We tune your stack so you pay only for what you need.",
		tags: ["Golang", "Terraform", "Monitoring"],
	},
];

const stats = [
	{ value: "5+", label: "Years Remote" },
	{ value: "200+", label: "Users Served" },
	{ value: "100%", label: "Remote Delivery" },
	{ value: "24/7", label: "Uptime Focus" },
];

const faqs = [
	{
		q: "How do you deliver services remotely?",
		a: "All work is delivered digitally — configured systems, deployment access, managed software solutions, or subscription-based maintenance — with clear handoff documentation.",
	},
	{
		q: "What types of clients do you work with?",
		a: "Primarily small businesses and individual clients who need reliable technical support for their online presence and digital infrastructure.",
	},
	{
		q: "Do you offer ongoing support?",
		a: "Yes. We offer subscription-based maintenance and support tiers, keeping your systems monitored, updated, and secure.",
	},
];

function LandingPage() {
	const [scrolled, setScrolled] = useState(false);
	const [openFaq, setOpenFaq] = useState(null);
	const heroRef = useRef(null);

	useEffect(() => {
		const onScroll = () => setScrolled(window.scrollY > 20);
		window.addEventListener("scroll", onScroll);
		return () => window.removeEventListener("scroll", onScroll);
	}, []);

	return (
		<>
			<style>{`
        @import url('https://fonts.googleapis.com/css2?family=Space+Mono:wght@400;700&family=DM+Sans:ital,wght@0,300;0,400;0,500;0,700&display=swap');
 
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
 
        :root {
          --bg: #050c18;
          --bg2: #08111f;
          --bg3: #0d1a2d;
          --cyan: #00d4ff;
          --cyan-dim: #00d4ff22;
          --cyan-glow: #00d4ff55;
          --border: rgba(0, 212, 255, 0.15);
          --border-hover: rgba(0, 212, 255, 0.4);
          --text: #e2eaf5;
          --text-muted: #6b849e;
          --mono: 'Space Mono', monospace;
          --sans: 'DM Sans', sans-serif;
        }
 
        html { scroll-behavior: smooth; }
 
        body {
          background: var(--bg);
          color: var(--text);
          font-family: var(--sans);
          line-height: 1.6;
        }
 
        .grid-bg {
          background-image:
            linear-gradient(rgba(0,212,255,0.04) 1px, transparent 1px),
            linear-gradient(90deg, rgba(0,212,255,0.04) 1px, transparent 1px);
          background-size: 48px 48px;
        }
 
        .glow-dot {
          width: 480px; height: 480px;
          border-radius: 50%;
          background: radial-gradient(circle, rgba(0,212,255,0.08) 0%, transparent 70%);
          pointer-events: none;
        }
 
        @keyframes fadeUp {
          from { opacity: 0; transform: translateY(24px); }
          to   { opacity: 1; transform: translateY(0); }
        }
        @keyframes pulse-ring {
          0%, 100% { opacity: 0.4; transform: scale(1); }
          50%       { opacity: 0.9; transform: scale(1.06); }
        }
        @keyframes blink { 0%,100%{opacity:1} 50%{opacity:0} }
        @keyframes scan {
          0%   { transform: translateY(-100%); }
          100% { transform: translateY(100vh); }
        }
 
        .fade-up { animation: fadeUp 0.7s ease forwards; }
        .delay-1 { animation-delay: 0.1s; opacity: 0; }
        .delay-2 { animation-delay: 0.25s; opacity: 0; }
        .delay-3 { animation-delay: 0.4s; opacity: 0; }
        .delay-4 { animation-delay: 0.55s; opacity: 0; }
 
        .cursor::after {
          content: '|';
          animation: blink 1s infinite;
          color: var(--cyan);
        }
 
        .service-card {
          background: var(--bg2);
          border: 1px solid var(--border);
          border-radius: 12px;
          padding: 2rem;
          transition: border-color 0.3s, background 0.3s, transform 0.2s;
        }
        .service-card:hover {
          border-color: var(--border-hover);
          background: var(--bg3);
          transform: translateY(-4px);
        }
 
        .tag {
          display: inline-block;
          font-family: var(--mono);
          font-size: 11px;
          padding: 2px 10px;
          border-radius: 4px;
          background: var(--cyan-dim);
          color: var(--cyan);
          border: 1px solid var(--cyan-glow);
          letter-spacing: 0.03em;
        }
 
        .btn-primary {
          display: inline-flex;
          align-items: center;
          gap: 8px;
          background: var(--cyan);
          color: #050c18;
          font-family: var(--mono);
          font-weight: 700;
          font-size: 14px;
          letter-spacing: 0.08em;
          padding: 14px 32px;
          border-radius: 6px;
          border: none;
          cursor: pointer;
          transition: opacity 0.2s, transform 0.15s, box-shadow 0.2s;
          text-decoration: none;
        }
        .btn-primary:hover {
          opacity: 0.9;
          transform: translateY(-1px);
          box-shadow: 0 0 32px rgba(0,212,255,0.4);
        }
 
        .btn-outline {
          display: inline-flex;
          align-items: center;
          gap: 8px;
          background: transparent;
          color: var(--cyan);
          font-family: var(--mono);
          font-size: 13px;
          font-weight: 700;
          letter-spacing: 0.06em;
          padding: 10px 22px;
          border-radius: 6px;
          border: 1px solid var(--border-hover);
          cursor: pointer;
          transition: background 0.2s, border-color 0.2s;
          text-decoration: none;
        }
        .btn-outline:hover {
          background: var(--cyan-dim);
          border-color: var(--cyan);
        }
 
        .stat-card {
          background: var(--bg2);
          border: 1px solid var(--border);
          border-radius: 10px;
          padding: 1.5rem;
          text-align: center;
          transition: border-color 0.3s;
        }
        .stat-card:hover { border-color: var(--border-hover); }
 
        .faq-item {
          border-bottom: 1px solid var(--border);
          padding: 1.25rem 0;
        }
        .faq-btn {
          width: 100%;
          background: none;
          border: none;
          color: var(--text);
          font-family: var(--sans);
          font-size: 16px;
          font-weight: 500;
          text-align: left;
          cursor: pointer;
          display: flex;
          justify-content: space-between;
          align-items: center;
          gap: 1rem;
          padding: 0;
        }
        .faq-chevron {
          color: var(--cyan);
          transition: transform 0.25s;
          flex-shrink: 0;
        }
        .faq-chevron.open { transform: rotate(180deg); }
        .faq-answer {
          overflow: hidden;
          transition: max-height 0.3s ease, opacity 0.3s;
          max-height: 0;
          opacity: 0;
        }
        .faq-answer.open { max-height: 200px; opacity: 1; }
 
        nav {
          position: fixed; top: 0; left: 0; right: 0; z-index: 50;
          transition: background 0.3s, border-color 0.3s, backdrop-filter 0.3s;
        }
        nav.scrolled {
          background: rgba(5,12,24,0.85);
          backdrop-filter: blur(12px);
          border-bottom: 1px solid var(--border);
        }
 
        .scan-line {
          position: absolute;
          left: 0; right: 0;
          height: 2px;
          background: linear-gradient(90deg, transparent, var(--cyan), transparent);
          opacity: 0.06;
          animation: scan 8s linear infinite;
          pointer-events: none;
        }
 
        section { padding: 96px 0; }
 
        .container {
          max-width: 1120px;
          margin: 0 auto;
          padding: 0 24px;
        }
 
        .section-label {
          font-family: var(--mono);
          font-size: 12px;
          letter-spacing: 0.2em;
          color: var(--cyan);
          text-transform: uppercase;
          margin-bottom: 12px;
        }
 
        h1, h2, h3 { line-height: 1.15; }
        h1 { font-size: clamp(2.4rem, 5vw, 4rem); font-weight: 700; letter-spacing: -0.02em; }
        h2 { font-size: clamp(1.8rem, 3.5vw, 2.6rem); font-weight: 700; letter-spacing: -0.02em; }
        h3 { font-size: 1.15rem; font-weight: 600; }
 
        .gradient-text {
          background: linear-gradient(135deg, #e2eaf5 30%, var(--cyan));
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
          background-clip: text;
        }
 
        footer {
          border-top: 1px solid var(--border);
          padding: 40px 0;
        }
 
        @media (max-width: 768px) {
          .services-grid { grid-template-columns: 1fr !important; }
          .stats-grid { grid-template-columns: 1fr 1fr !important; }
        }
      `}</style>

			{/* NAV */}
			<nav className={scrolled ? "scrolled" : ""}>
				<div className="container" style={{ display: "flex", alignItems: "center", justifyContent: "space-between", height: 72 }}>
					<a href="/" style={{ display: "flex", alignItems: "center", gap: 10, textDecoration: "none" }}>
						<img
							src={logoBase64}
							alt="logo"
							style={{ width: 36, height: 36, objectFit: "contain", filter: "brightness(1.1)" }}
						/>
						<span style={{ fontFamily: "var(--mono)", fontSize: 15, fontWeight: 700, color: "var(--text)", letterSpacing: "0.04em" }}>
							UVP<span style={{ color: "var(--cyan)" }}>{" "}DevOps</span>
						</span>
					</a>

					<div style={{ display: "flex", alignItems: "center", gap: 12 }}>
						<div style={{ display: "flex", gap: 28, marginRight: 16 }}>
							{["Services", "About", "FAQ"].map((item) => (
								<a key={item} href={`#${item.toLowerCase()}`} style={{ fontFamily: "var(--sans)", fontSize: 14, color: "var(--text-muted)", textDecoration: "none", transition: "color 0.2s" }}
									onMouseEnter={e => e.target.style.color = "var(--text)"}
									onMouseLeave={e => e.target.style.color = "var(--text-muted)"}
								>{item}</a>
							))}
						</div>
						<a href="/login" className="btn-outline" style={{ fontSize: 12, padding: "8px 18px" }}>
							Login
						</a>
					</div>
				</div>
			</nav>

			{/* HERO */}
			<section id="hero" className="grid-bg" style={{ position: "relative", overflow: "hidden", paddingTop: 160, paddingBottom: 120 }}>
				<div className="scan-line" />
				{/* Glow blobs */}
				<div className="glow-dot" style={{ position: "absolute", top: -120, right: -80, pointerEvents: "none" }} />
				<div className="glow-dot" style={{ position: "absolute", bottom: -160, left: -120, pointerEvents: "none", opacity: 0.5 }} />

				<div className="container" style={{ position: "relative" }}>
					<p className="section-label fade-up delay-1">// remote software & devops services</p>

					<h1 className="gradient-text fade-up delay-2" style={{ maxWidth: 780, marginBottom: "1.5rem" }}>
						Infrastructure That&nbsp;
						<br />
						<span className="cursor">Works While You Sleep</span>
					</h1>

					<p className="fade-up delay-3" style={{ fontSize: "1.15rem", color: "var(--text-muted)", maxWidth: 560, marginBottom: "2.5rem", lineHeight: 1.7 }}>
						Website development, cloud deployment, and infrastructure optimization — delivered entirely remotely for small businesses and individual clients.
					</p>

					<div className="fade-up delay-4" style={{ display: "flex", gap: 16, flexWrap: "wrap" }}>
						<a href="mailto:warley8013@gmail.com" className="btn-primary">
							Contact Us
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.5}>
								<path d="M5 12h14M12 5l7 7-7 7" />
							</svg>
						</a>
						<a href="#services" className="btn-outline">See Services</a>
					</div>

					{/* Terminal badge */}
					<div style={{ marginTop: "3.5rem", display: "inline-flex", alignItems: "center", gap: 10, background: "var(--bg2)", border: "1px solid var(--border)", borderRadius: 8, padding: "10px 18px" }}>
						<span style={{ width: 8, height: 8, borderRadius: "50%", background: "#22c55e", animation: "pulse-ring 2s infinite", display: "inline-block" }} />
						<span style={{ fontFamily: "var(--mono)", fontSize: 12, color: "var(--text-muted)" }}>
							<span style={{ color: "var(--cyan)" }}>$</span> uptime --all-systems
							<span style={{ color: "#22c55e", marginLeft: 12 }}>● operational</span>
						</span>
					</div>
				</div>
			</section>

			{/* STATS */}
			<div style={{ borderTop: "1px solid var(--border)", borderBottom: "1px solid var(--border)", padding: "48px 0", background: "var(--bg2)" }}>
				<div className="container">
					<div className="stats-grid" style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 16 }}>
						{stats.map((s) => (
							<div key={s.label} className="stat-card">
								<div style={{ fontFamily: "var(--mono)", fontSize: "2rem", fontWeight: 700, color: "var(--cyan)", marginBottom: 4 }}>{s.value}</div>
								<div style={{ fontSize: 13, color: "var(--text-muted)", letterSpacing: "0.05em" }}>{s.label}</div>
							</div>
						))}
					</div>
				</div>
			</div>

			{/* SERVICES */}
			<section id="services">
				<div className="container">
					<p className="section-label">// what we build</p>
					<h2 style={{ marginBottom: "0.75rem" }}>Our Services</h2>
					<p style={{ color: "var(--text-muted)", fontSize: "1.05rem", maxWidth: 500, marginBottom: "3.5rem", lineHeight: 1.7 }}>
						Full-cycle delivery from design to production — everything your online presence needs, maintained and optimized.
					</p>

					<div className="services-grid" style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 20 }}>
						{services.map((svc) => (
							<div key={svc.title} className="service-card">
								<div style={{ color: "var(--cyan)", marginBottom: "1.25rem" }}>{svc.icon}</div>
								<h3 style={{ marginBottom: "0.75rem" }}>{svc.title}</h3>
								<p style={{ color: "var(--text-muted)", fontSize: "0.95rem", lineHeight: 1.7, marginBottom: "1.25rem" }}>{svc.desc}</p>
								<div style={{ display: "flex", flexWrap: "wrap", gap: 6 }}>
									{svc.tags.map((t) => <span key={t} className="tag">{t}</span>)}
								</div>
							</div>
						))}
					</div>
				</div>
			</section>

			{/* ABOUT / HOW WE WORK */}
			<section id="about" style={{ background: "var(--bg2)", borderTop: "1px solid var(--border)", borderBottom: "1px solid var(--border)" }}>
				<div className="container" style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "4rem", alignItems: "center" }}>
					<div>
						<p className="section-label">// how we work</p>
						<h2 style={{ marginBottom: "1rem" }}>100% Remote,<br />Zero Compromise</h2>
						<p style={{ color: "var(--text-muted)", lineHeight: 1.8, marginBottom: "1.5rem" }}>
							Every engagement is delivered remotely. You receive configured systems, deployment access, or managed software solutions — with documentation clear enough to hand off to anyone.
						</p>
						<p style={{ color: "var(--text-muted)", lineHeight: 1.8, marginBottom: "2rem" }}>
							Need ongoing peace of mind? Subscription maintenance keeps your systems monitored, patched, and performing — without hiring in-house.
						</p>
						<a href="mailto:warley8013@gmail.com" className="btn-primary">Contact Us →</a>
					</div>

					{/* Code block decoration */}
					<div style={{ background: "var(--bg)", border: "1px solid var(--border)", borderRadius: 12, padding: "1.5rem", fontFamily: "var(--mono)", fontSize: 13, lineHeight: 2 }}>
						<div style={{ display: "flex", gap: 6, marginBottom: "1rem" }}>
							{["#ff5f57", "#ffbd2e", "#28c840"].map(c => (
								<span key={c} style={{ width: 12, height: 12, borderRadius: "50%", background: c, display: "inline-block" }} />
							))}
						</div>
						<div><span style={{ color: "var(--text-muted)" }}>// deploy.yml</span></div>
						<div><span style={{ color: "#c792ea" }}>on:</span> <span style={{ color: "var(--cyan)" }}>push</span></div>
						<div><span style={{ color: "#c792ea" }}>jobs:</span></div>
						<div style={{ paddingLeft: "1rem" }}><span style={{ color: "#82aaff" }}>build:</span></div>
						<div style={{ paddingLeft: "2rem" }}><span style={{ color: "var(--text-muted)" }}>runs-on:</span> <span style={{ color: "#c3e88d" }}>ubuntu-latest</span></div>
						<div style={{ paddingLeft: "2rem" }}><span style={{ color: "var(--text-muted)" }}>steps:</span></div>
						<div style={{ paddingLeft: "3rem" }}><span style={{ color: "var(--cyan)" }}>- uses:</span> <span style={{ color: "#c3e88d" }}>actions/checkout@v4</span></div>
						<div style={{ paddingLeft: "3rem" }}><span style={{ color: "var(--cyan)" }}>- run:</span> <span style={{ color: "#f78c6c" }}>make deploy</span></div>
						<div style={{ paddingLeft: "3rem", color: "#22c55e" }}>✓ Deployed in 47s</div>
					</div>
				</div>
			</section>

			{/* FAQ */}
			<section id="faq">
				<div className="container" style={{ maxWidth: 720 }}>
					<p className="section-label">// common questions</p>
					<h2 style={{ marginBottom: "3rem" }}>FAQ</h2>
					{faqs.map((item, i) => (
						<div key={i} className="faq-item">
							<button className="faq-btn" onClick={() => setOpenFaq(openFaq === i ? null : i)}>
								{item.q}
								<svg className={`faq-chevron ${openFaq === i ? "open" : ""}`} width={18} height={18} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
									<path d="M6 9l6 6 6-6" />
								</svg>
							</button>
							<div className={`faq-answer ${openFaq === i ? "open" : ""}`}>
								<p style={{ paddingTop: "0.875rem", color: "var(--text-muted)", fontSize: "0.95rem", lineHeight: 1.75 }}>{item.a}</p>
							</div>
						</div>
					))}
				</div>
			</section>

			{/* CTA BANNER */}
			<section style={{ background: "var(--bg2)", borderTop: "1px solid var(--border)", paddingTop: 80, paddingBottom: 80 }}>
				<div className="container" style={{ textAlign: "center" }}>
					<p className="section-label" style={{ textAlign: "center" }}>// ready to ship?</p>
					<h2 style={{ marginBottom: "1rem", maxWidth: 560, margin: "0 auto 1rem" }}>Let's Build Something That Lasts</h2>
					<p style={{ color: "var(--text-muted)", marginBottom: "2.5rem", maxWidth: 440, margin: "0 auto 2.5rem" }}>
						Reach out and describe what you need — we'll scope it and get started.
					</p>
					<a href="mailto:warley8013@gmail.com" className="btn-primary" style={{ fontSize: 15, padding: "16px 40px" }}>
						Contact Us
					</a>
				</div>
			</section>

			{/* FOOTER */}
			<footer>
				<div className="container" style={{ display: "flex", justifyContent: "space-between", alignItems: "center", flexWrap: "wrap", gap: 16 }}>
					<div style={{ display: "flex", alignItems: "center", gap: 10 }}>
						<img
							src={logoBase64}
							alt="logo"
							style={{ width: 28, height: 28, objectFit: "contain", opacity: 0.8 }}
						/>
						<span style={{ fontFamily: "var(--mono)", fontSize: 13, color: "var(--text-muted)" }}>
							UVP<span style={{ color: "var(--cyan)" }}>{" "}DevOps</span>
						</span>
					</div>
					<p style={{ fontFamily: "var(--mono)", fontSize: 12, color: "var(--text-muted)" }}>
						© {new Date().getFullYear()} — All services delivered remotely
					</p>
					<div style={{ display: "flex", gap: 20 }}>
						{["Services", "FAQ", "Login"].map((l) => (
							<a key={l} href={l === "Login" ? "/login" : `#${l.toLowerCase()}`}
								style={{ fontFamily: "var(--mono)", fontSize: 12, color: "var(--text-muted)", textDecoration: "none", transition: "color 0.2s" }}
								onMouseEnter={e => e.target.style.color = "var(--cyan)"}
								onMouseLeave={e => e.target.style.color = "var(--text-muted)"}
							>{l}</a>
						))}
					</div>
				</div>
			</footer>
		</>
	);
}

function RequireAuth({ children }) {
	const loginState = useSelector((state) => state.login);

	return loginState.isLogin === true ? (
		children
	) : (
		<Navigate to="/login" replace />
	);
}

function App() {
	return (
		<BrowserRouter>
			<Routes>
				<Route path="/user" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<User />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/login" element={<Login />} />
				<Route path="/mypanel" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<Mypanel />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/addnode" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<AddNode />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/nodes" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<Nodes />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/macos" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<Macos />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/windows" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<Windows />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/iphone" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<Iphone />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/android" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<Android />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/paymentinput" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<PaymentInput />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/paymentstatistics" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<PaymentStatistics />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/paymentrecords" element={
					<RequireAuth>
						<div className="min-h-screen bg-gray-900 flex flex-col">
							<Menu />
							<div className="flex-1">
								<PaymentRecords />
							</div>
							<Footer />
						</div>
					</RequireAuth>
				} />
				<Route path="/" element={<LandingPage />} />
			</Routes>
		</BrowserRouter>
	);
}

export default App;
