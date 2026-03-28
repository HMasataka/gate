import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { setToken, setEmail } from "../lib/auth";

export default function Login() {
  const [email, setEmailValue] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  async function handleLogin() {
    setError("");

    if (!email.trim() || !password) {
      setError("Email and password are required.");
      return;
    }

    setLoading(true);

    try {
      const res = await fetch("/api/v1/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: email.trim(), password }),
      });

      const data = await res.json().catch(() => ({}));

      if (!res.ok) {
        setError(data.error?.message ?? data.message ?? "Invalid credentials.");
        return;
      }

      if (data.access_token) {
        setToken(data.access_token);
        setEmail(email.trim());
      }

      navigate("/clients");
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setLoading(false);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter") handleLogin();
  }

  return (
    <div className="auth-layout" onKeyDown={handleKeyDown}>
      <div className="auth-brand">
        <div className="auth-brand-title">gate admin</div>
        <div className="auth-brand-subtitle">Server Management Console</div>
        <div className="auth-brand-desc">
          Administer OAuth clients, user accounts, tokens, and server
          configuration. Restricted access only.
        </div>
      </div>

      <div className="auth-form-area">
        <div className="auth-form-card">
          <h2>Administrator Sign in</h2>
          <p className="subtitle">
            Enter your admin credentials to access the management console.
          </p>

          {error && (
            <div className="alert alert-error" style={{ marginBottom: 16 }}>
              {error}
            </div>
          )}

          <div className="form-group">
            <label className="form-label" htmlFor="email">Email</label>
            <input
              className="form-input"
              type="email"
              id="email"
              placeholder="admin@example.com"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmailValue(e.target.value)}
            />
          </div>

          <div className="form-group">
            <label className="form-label" htmlFor="password">Password</label>
            <input
              className="form-input"
              type="password"
              id="password"
              placeholder="Enter your password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>

          <button
            className="btn btn-primary btn-full"
            disabled={loading}
            onClick={handleLogin}
          >
            {loading ? "Signing in..." : "Sign in as Administrator"}
          </button>

          <div className="auth-footer">
            <a href="/portal/login" rel="noopener">Back to user login</a>
          </div>
        </div>
      </div>
    </div>
  );
}
