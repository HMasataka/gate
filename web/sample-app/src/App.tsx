import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { ToastProvider } from "./components/Toast";
import Home from "./pages/Home";
import Callback from "./pages/Callback";
import Dashboard from "./pages/Dashboard";

export default function App() {
  return (
    <BrowserRouter basename="/sample-app">
      <ToastProvider>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/callback" element={<Callback />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </ToastProvider>
    </BrowserRouter>
  );
}
