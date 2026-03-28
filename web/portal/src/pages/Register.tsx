import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useToast } from "../components/Toast";

export default function Register() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { showToast } = useToast();

  async function handleRegister() {
    setError("");

    if (!email.trim() || !password || !confirmPassword) {
      setError("All fields are required.");
      return;
    }

    if (password !== confirmPassword) {
      setError("Passwords do not match.");
      return;
    }

    if (password.length < 8) {
      setError("Password must be at least 8 characters.");
      return;
    }

    setLoading(true);

    try {
      const res = await fetch("/api/v1/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: email.trim(), password }),
      });

      const data = await res.json().catch(() => ({}));

      if (!res.ok) {
        setError(data.error?.message ?? data.message ?? "Registration failed. Please try again.");
        return;
      }

      showToast("Account created successfully. Please sign in.");
      setTimeout(() => navigate("/login"), 1200);
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setLoading(false);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter") handleRegister();
  }

  return (
    <div className="auth-layout" onKeyDown={handleKeyDown}>
      <div className="auth-brand">
        <div className="auth-brand-title">gate</div>
        <div className="auth-brand-subtitle">Identity Provider</div>
        <div className="auth-brand-desc">
          Create your gate account to start building OAuth applications.
          Access the client portal and manage your integrations.
        </div>
      </div>

      <div className="auth-form-area">
        <div className="auth-form-card">
          <h2>Create account</h2>
          <p className="subtitle">Register to get started with gate.</p>

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
              placeholder="user@example.com"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </div>

          <div className="form-group">
            <label className="form-label" htmlFor="password">Password</label>
            <input
              className="form-input"
              type="password"
              id="password"
              placeholder="Choose a password"
              autoComplete="new-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>

          <div className="form-group">
            <label className="form-label" htmlFor="confirm-password">Confirm password</label>
            <input
              className="form-input"
              type="password"
              id="confirm-password"
              placeholder="Repeat your password"
              autoComplete="new-password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
            />
          </div>

          <button
            className="btn btn-primary btn-full"
            disabled={loading}
            onClick={handleRegister}
          >
            {loading ? "Creating account..." : "Create account"}
          </button>

          <div className="auth-footer">
            Already have an account?{" "}
            <Link to="/login">Sign in</Link>
          </div>
        </div>
      </div>
    </div>
  );
}
