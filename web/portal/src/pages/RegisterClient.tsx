import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { isAuthenticated, getEmail, clearAuth } from "../lib/auth";
import { api } from "../lib/api";
import { useToast } from "../components/Toast";

export default function RegisterClient() {
  const [step, setStep] = useState(1);
  const [appName, setAppName] = useState("");
  const [clientType, setClientType] = useState("confidential");
  const [redirectUris, setRedirectUris] = useState("");
  const [scopes, setScopes] = useState<Set<string>>(new Set(["openid", "email"]));
  const [grantTypes, setGrantTypes] = useState<Set<string>>(new Set(["authorization_code"]));
  const [loading, setLoading] = useState(false);
  const [resultClientId, setResultClientId] = useState("");
  const [resultClientSecret, setResultClientSecret] = useState("");
  const [success, setSuccess] = useState(false);
  const navigate = useNavigate();
  const { showToast } = useToast();

  useEffect(() => {
    if (!isAuthenticated()) {
      navigate("/login");
    }
  }, [navigate]);

  function toggleScope(scope: string) {
    setScopes((prev) => {
      const next = new Set(prev);
      if (next.has(scope)) next.delete(scope);
      else next.add(scope);
      return next;
    });
  }

  function toggleGrantType(gt: string) {
    setGrantTypes((prev) => {
      const next = new Set(prev);
      if (next.has(gt)) next.delete(gt);
      else next.add(gt);
      return next;
    });
  }

  function goToStep(target: number) {
    if (target > step) {
      if (step === 1 && !appName.trim()) {
        showToast("Application name is required.", "error");
        return;
      }
      if (step === 2 && !redirectUris.trim()) {
        showToast("At least one redirect URI is required.", "error");
        return;
      }
    }
    setStep(target);
  }

  async function submitRegistration() {
    setLoading(true);

    const uris = redirectUris
      .split("\n")
      .map((u) => u.trim())
      .filter((u) => u.length > 0);

    try {
      const data = await api<{
        client_id?: string;
        client_secret?: string;
      }>("POST", "/api/v1/clients", {
        name: appName.trim(),
        client_type: clientType,
        redirect_uris: uris,
        scopes: [...scopes],
        grant_types: [...grantTypes],
      });

      setResultClientId(data.client_id ?? "");
      setResultClientSecret(data.client_secret ?? "");
      setSuccess(true);
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Registration failed.", "error");
    } finally {
      setLoading(false);
    }
  }

  function copyText(text: string) {
    navigator.clipboard.writeText(text).then(
      () => showToast("Copied to clipboard."),
      () => showToast("Copy failed.", "error"),
    );
  }

  function signOut() {
    clearAuth();
    navigate("/login");
  }

  const email = getEmail();

  const stepClass = (n: number) => {
    if (success) return "wizard-step done";
    if (n < step) return "wizard-step done";
    if (n === step) return "wizard-step active";
    return "wizard-step";
  };

  return (
    <div className="portal-layout">
      <nav className="portal-nav">
        <Link to="/clients" className="portal-nav-logo">
          <span>gate</span> / clients
        </Link>
        <div className="portal-nav-links">
          <Link to="/clients" className="portal-nav-link">My Clients</Link>
          <Link to="/register-client" className="portal-nav-link active">+ Register New</Link>
        </div>
        <div className="portal-nav-right">
          <span className="portal-user">{email}</span>
          <button className="btn btn-ghost btn-sm" onClick={signOut}>Sign out</button>
        </div>
      </nav>

      <div className="portal-content">
        <div className="portal-header">
          <h1>Register Application</h1>
          <p>Register a new OAuth 2.0 client application.</p>
        </div>

        <div className="wizard-container">
          <div className="wizard-steps">
            <div className={stepClass(1)}>
              <div className="step-number">1</div>
              <div className="step-label">App Details</div>
            </div>
            <div className={stepClass(2)}>
              <div className="step-number">2</div>
              <div className="step-label">Redirect URIs</div>
            </div>
            <div className={stepClass(3)}>
              <div className="step-number">3</div>
              <div className="step-label">Scopes</div>
            </div>
          </div>

          {!success && step === 1 && (
            <div>
              <div className="card">
                <div className="card-header">
                  <div>
                    <div className="card-title">Application Details</div>
                    <div className="card-subtitle">Basic information about your OAuth client.</div>
                  </div>
                </div>
                <div className="form-group">
                  <label className="form-label" htmlFor="app-name">Application Name</label>
                  <input
                    className="form-input"
                    type="text"
                    id="app-name"
                    placeholder="My Application"
                    value={appName}
                    onChange={(e) => setAppName(e.target.value)}
                  />
                  <div className="form-hint">A human-readable name for your application.</div>
                </div>
                <div className="form-group">
                  <label className="form-label">Client Type</label>
                  <div className="radio-group">
                    <label className="radio-option">
                      <input
                        type="radio"
                        name="client-type"
                        value="confidential"
                        checked={clientType === "confidential"}
                        onChange={() => setClientType("confidential")}
                      />
                      <div>
                        <div className="radio-option-label">Confidential</div>
                        <div className="radio-option-desc">
                          Clients that can securely store a client secret. Suitable for server-side web applications.
                        </div>
                      </div>
                    </label>
                    <label className="radio-option">
                      <input
                        type="radio"
                        name="client-type"
                        value="public"
                        checked={clientType === "public"}
                        onChange={() => setClientType("public")}
                      />
                      <div>
                        <div className="radio-option-label">Public</div>
                        <div className="radio-option-desc">
                          Clients that cannot store secrets. Suitable for SPAs, mobile apps, and CLI tools. Uses PKCE.
                        </div>
                      </div>
                    </label>
                  </div>
                </div>
              </div>
              <div className="wizard-actions">
                <div />
                <div className="right">
                  <button className="btn btn-primary" onClick={() => goToStep(2)}>Next</button>
                </div>
              </div>
            </div>
          )}

          {!success && step === 2 && (
            <div>
              <div className="card">
                <div className="card-header">
                  <div>
                    <div className="card-title">Redirect URIs</div>
                    <div className="card-subtitle">Allowed callback endpoints for the OAuth flow.</div>
                  </div>
                </div>
                <div className="form-group">
                  <label className="form-label" htmlFor="redirect-uris">Redirect URIs</label>
                  <textarea
                    className="form-input"
                    id="redirect-uris"
                    rows={5}
                    placeholder={"http://localhost:3000/callback\nhttps://myapp.example.com/callback"}
                    value={redirectUris}
                    onChange={(e) => setRedirectUris(e.target.value)}
                  />
                  <div className="form-hint">Enter one redirect URI per line.</div>
                </div>
              </div>
              <div className="wizard-actions">
                <button className="btn btn-secondary" onClick={() => goToStep(1)}>Back</button>
                <div className="right">
                  <button className="btn btn-primary" onClick={() => goToStep(3)}>Next</button>
                </div>
              </div>
            </div>
          )}

          {!success && step === 3 && (
            <div>
              <div className="card">
                <div className="card-header">
                  <div>
                    <div className="card-title">Scopes &amp; Grant Types</div>
                    <div className="card-subtitle">Permissions and flows supported by your client.</div>
                  </div>
                </div>
                <div className="form-group">
                  <label className="form-label">Scopes</label>
                  <div className="checkbox-group">
                    {["openid", "email", "profile"].map((s) => (
                      <label key={s} className="checkbox-option">
                        <input
                          type="checkbox"
                          checked={scopes.has(s)}
                          onChange={() => toggleScope(s)}
                        />
                        <span>{s}</span>
                        <span className="checkbox-option-desc">
                          {s === "openid" ? "Required for OIDC" : s === "email" ? "User email address" : "User profile info"}
                        </span>
                      </label>
                    ))}
                  </div>
                </div>
                <hr className="divider" />
                <div className="form-group">
                  <label className="form-label">Grant Types</label>
                  <div className="checkbox-group">
                    {["authorization_code", "refresh_token"].map((gt) => (
                      <label key={gt} className="checkbox-option">
                        <input
                          type="checkbox"
                          checked={grantTypes.has(gt)}
                          onChange={() => toggleGrantType(gt)}
                        />
                        <span>{gt}</span>
                        <span className="checkbox-option-desc">
                          {gt === "authorization_code" ? "Standard OAuth flow" : "Long-lived sessions"}
                        </span>
                      </label>
                    ))}
                  </div>
                </div>
              </div>
              <div className="wizard-actions">
                <button className="btn btn-secondary" onClick={() => goToStep(2)}>Back</button>
                <div className="right">
                  <button
                    className="btn btn-primary"
                    disabled={loading}
                    onClick={submitRegistration}
                  >
                    {loading ? "Registering..." : "Register Application"}
                  </button>
                </div>
              </div>
            </div>
          )}

          {success && (
            <div className="card">
              <div className="success-screen">
                <div className="success-icon">+</div>
                <div className="success-title">Application registered successfully</div>
                <div className="success-subtitle">
                  Your OAuth client has been created. Save your credentials below.
                </div>

                <div className="alert alert-warning" style={{ marginBottom: 24, textAlign: "left" }}>
                  Save your client secret now. It will not be shown again.
                </div>

                <div className="credential-group">
                  <div className="credential-label">Client ID</div>
                  <div className="secret-box">
                    <button
                      className="btn btn-secondary btn-sm copy-btn"
                      onClick={() => copyText(resultClientId)}
                    >
                      Copy
                    </button>
                    <span>{resultClientId}</span>
                  </div>
                </div>

                {resultClientSecret && (
                  <div className="credential-group">
                    <div className="credential-label">Client Secret</div>
                    <div className="secret-box">
                      <button
                        className="btn btn-secondary btn-sm copy-btn"
                        onClick={() => copyText(resultClientSecret)}
                      >
                        Copy
                      </button>
                      <span>{resultClientSecret}</span>
                    </div>
                  </div>
                )}

                <div style={{ marginTop: 24 }}>
                  <Link to="/clients" className="btn btn-primary">
                    Go to My Clients
                  </Link>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
