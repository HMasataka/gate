import { useEffect, useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { getToken, clearAuth } from "../lib/auth";
import { api } from "../lib/api";
import { useToast } from "../components/Toast";

interface UserInfo {
  [key: string]: unknown;
}

const KNOWN_FIELDS = [
  "sub",
  "email",
  "email_verified",
  "updated_at",
];

export default function Dashboard() {
  const [userInfo, setUserInfo] = useState<UserInfo | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const { showToast } = useToast();

  useEffect(() => {
    const token = getToken();
    if (!token) {
      navigate("/", { replace: true });
      return;
    }

    async function loadUserInfo() {
      try {
        const data = await api<UserInfo>("GET", "/oauth/userinfo");
        setUserInfo(data);
      } catch (err) {
        const message = err instanceof Error ? err.message : "Failed to load user info.";
        setError(message);
      } finally {
        setLoading(false);
      }
    }

    loadUserInfo();
  }, [navigate]);

  function signOut() {
    clearAuth();
    navigate("/", { replace: true });
  }

  function copyToken() {
    const token = getToken() ?? "";
    navigator.clipboard.writeText(token).then(
      () => showToast("Access token copied to clipboard."),
      () => showToast("Failed to copy token.", "error"),
    );
  }

  if (loading) {
    return (
      <div className="sample-layout">
        <nav className="sample-nav">
          <Link className="sample-nav-logo" to="/">
            <strong>sample app</strong>
          </Link>
        </nav>
        <div className="sample-hero">
          <div className="processing-card">
            <div className="spinner"></div>
            <div className="processing-subtitle">Loading your profile...</div>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="sample-layout">
        <nav className="sample-nav">
          <Link className="sample-nav-logo" to="/">
            <strong>sample app</strong>
          </Link>
          <button className="btn btn-secondary btn-sm" onClick={signOut}>
            Sign out
          </button>
        </nav>
        <div className="sample-hero">
          <div className="processing-card">
            <div className="alert alert-error" style={{ textAlign: "left" }}>
              <div>
                <strong>Failed to load profile</strong>
                <div style={{ marginTop: 6 }}>{error}</div>
              </div>
            </div>
            <div style={{ marginTop: 20 }}>
              <button className="btn btn-secondary" onClick={signOut}>
                Sign out
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  const token = getToken() ?? "";
  const tokenPreview = token.substring(0, 40) + "...";
  const email = String(userInfo?.email ?? userInfo?.sub ?? "user");

  const knownEntries = KNOWN_FIELDS.filter((key) => userInfo?.[key] !== undefined);
  const extraEntries = Object.keys(userInfo ?? {}).filter(
    (key) => !KNOWN_FIELDS.includes(key),
  );

  return (
    <div className="sample-layout">
      <nav className="sample-nav">
        <Link className="sample-nav-logo" to="/">
          <strong>sample app</strong>
        </Link>
        <button className="btn btn-secondary btn-sm" onClick={signOut}>
          Sign out
        </button>
      </nav>

      <div className="dashboard-content">
        <div className="user-greeting">
          Hello, <strong>{email}</strong>
        </div>

        <div className="info-card">
          <h3>User Information</h3>
          {knownEntries.map((key) => (
            <div className="info-row" key={key}>
              <span className="info-key">{key}</span>
              <span className="info-value">{String(userInfo![key])}</span>
            </div>
          ))}
          {extraEntries.map((key) => (
            <div className="info-row" key={key}>
              <span className="info-key">{key}</span>
              <span className="info-value">{String(userInfo![key])}</span>
            </div>
          ))}
        </div>

        <div className="info-card">
          <h3>Token Information</h3>
          <div className="info-row">
            <span className="info-key">access_token</span>
            <div className="token-row">
              <span className="token-truncated">{tokenPreview}</span>
              <button className="btn btn-ghost btn-sm" onClick={copyToken}>
                Copy
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
