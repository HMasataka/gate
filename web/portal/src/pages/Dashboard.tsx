import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { getToken, getEmail, clearAuth } from "../lib/auth";

export default function Dashboard() {
  const [email, setEmail] = useState("");
  const [tokenPreview, setTokenPreview] = useState("");
  const [tokenExpiry, setTokenExpiry] = useState("");
  const navigate = useNavigate();

  useEffect(() => {
    const token = getToken();
    if (!token) {
      navigate("/login");
      return;
    }

    setEmail(getEmail());
    setTokenPreview(token.substring(0, 40) + "...");

    try {
      const payload = JSON.parse(atob(token.split(".")[1]));
      if (payload.email) setEmail(payload.email);
      if (payload.exp) {
        const exp = new Date(payload.exp * 1000);
        const diffMin = Math.round((exp.getTime() - Date.now()) / 60000);
        setTokenExpiry(diffMin > 0 ? `Expires in ${diffMin} minutes` : "Expired");
      }
    } catch {
      // token decode not critical
    }
  }, [navigate]);

  function handleLogout() {
    clearAuth();
    navigate("/login");
  }

  return (
    <div className="dashboard">
      <header className="dash-header">
        <div className="dash-header-left">
          <div className="dash-logo">gate</div>
          <nav className="dash-nav">
            <span className="dash-nav-item active">Overview</span>
            <Link to="/clients" className="dash-nav-item">My Applications</Link>
          </nav>
        </div>
        <div className="dash-header-right">
          <span className="dash-user">{email}</span>
          <button className="btn btn-secondary btn-sm" onClick={handleLogout}>
            Logout
          </button>
        </div>
      </header>

      <main className="dash-main">
        <div className="dash-greeting">
          <h1>Welcome back</h1>
          <p>Your account overview and active sessions.</p>
        </div>

        <div className="dash-stats">
          <div className="stat-card">
            <div className="stat-label">Active Sessions</div>
            <div className="stat-value">1</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Email Status</div>
            <div className="stat-value amber">Pending</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">MFA</div>
            <div className="stat-value">Off</div>
          </div>
        </div>

        <div className="dash-section">
          <div className="dash-section-header">
            <h2 className="dash-section-title">Account</h2>
          </div>
          <div className="account-grid">
            <div className="account-field">
              <div className="account-field-label">Email</div>
              <div className="account-field-value">{email}</div>
            </div>
            <div className="account-field">
              <div className="account-field-label">Email Verified</div>
              <div className="account-field-value">
                <span className="badge badge-amber">Pending</span>
              </div>
            </div>
            <div className="account-field">
              <div className="account-field-label">Account Status</div>
              <div className="account-field-value">
                <span className="badge badge-green">Active</span>
              </div>
            </div>
            <div className="account-field">
              <div className="account-field-label">Two-Factor Auth</div>
              <div className="account-field-value">
                <span className="badge badge-gray">Disabled</span>
              </div>
            </div>
          </div>
        </div>

        <div className="dash-section">
          <div className="dash-section-header">
            <h2 className="dash-section-title">Token Info</h2>
          </div>
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16 }}>
            <div className="stat-card">
              <div className="stat-label">Access Token</div>
              <div style={{ fontSize: 11, color: "var(--text-secondary)", wordBreak: "break-all", lineHeight: 1.8, background: "var(--bg-0)", border: "1px solid var(--border)", borderRadius: "var(--radius)", padding: "10px 12px", maxHeight: 60, overflow: "hidden", marginTop: 8 }}>
                {tokenPreview}
              </div>
              <div style={{ fontSize: 11, color: "var(--text-muted)", marginTop: 8 }}>
                {tokenExpiry}
              </div>
            </div>
            <div className="stat-card">
              <div className="stat-label">Refresh Token</div>
              <div style={{ fontSize: 11, color: "var(--text-secondary)", marginTop: 8 }}>
                Stored securely
              </div>
              <div style={{ fontSize: 11, color: "var(--text-muted)", marginTop: 8 }}>
                Expires in 7 days
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
