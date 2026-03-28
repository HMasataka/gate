import { useState } from "react";
import { generatePKCE, generateState } from "../lib/pkce";
import { setPkceVerifier, setOAuthState, setClientId, getClientId } from "../lib/auth";

export default function Home() {
  const [clientId, setClientIdInput] = useState(getClientId() ?? "");
  const [error, setError] = useState("");

  async function startLogin() {
    const trimmed = clientId.trim();
    if (!trimmed) {
      setError("Client ID is required.");
      return;
    }

    setClientId(trimmed);

    const { verifier, challenge } = await generatePKCE();
    const state = generateState();

    setPkceVerifier(verifier);
    setOAuthState(state);

    const redirectUri = window.location.origin + "/sample-app/callback";

    const params = new URLSearchParams({
      response_type: "code",
      client_id: trimmed,
      redirect_uri: redirectUri,
      scope: "openid email profile",
      code_challenge: challenge,
      code_challenge_method: "S256",
      state: state,
    });

    window.location.href = `/api/v1/oauth/authorize?${params.toString()}`;
  }

  return (
    <div className="sample-layout">
      <nav className="sample-nav">
        <a className="sample-nav-logo" href="/sample-app/">
          <strong>sample app</strong>
        </a>
        <span style={{ fontSize: 11, color: "var(--text-muted)" }}>
          gate OAuth 2.0 demo
        </span>
      </nav>

      <div className="sample-hero">
        <div className="hero-card">
          <h1>sample app</h1>
          <p>
            This is a demonstration of the gate OAuth 2.0 flow with PKCE.
            Enter your OAuth Client ID and click the button to authenticate.
          </p>

          <div className="form-group" style={{ textAlign: "left", marginTop: 20 }}>
            <label className="form-label" htmlFor="client-id">Client ID</label>
            <input
              className="form-input"
              id="client-id"
              type="text"
              value={clientId}
              onChange={(e) => {
                setClientIdInput(e.target.value);
                setError("");
              }}
              placeholder="Paste your OAuth Client ID here"
            />
            {error && <div className="form-error">{error}</div>}
          </div>

          <button
            className="btn btn-primary"
            onClick={startLogin}
            style={{ fontSize: 14, padding: "12px 28px", marginTop: 12 }}
          >
            Login with gate
          </button>
          <div className="note">
            You will be redirected to gate for authentication.
          </div>
        </div>
      </div>
    </div>
  );
}
