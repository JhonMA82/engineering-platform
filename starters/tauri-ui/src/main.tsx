import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";

function App() { return <main><p>Engineering Platform</p><h1>Aplicación de escritorio lista</h1><p>Implementa el primer incremento descrito en GENTLE.md.</p></main>; }
createRoot(document.getElementById("root")!).render(<React.StrictMode><App /></React.StrictMode>);
