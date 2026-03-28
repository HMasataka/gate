import { useEffect, useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { getOAuthState, getPkceVerifier, clearPkceState, setToken, setRefreshToken, getClientId } from "../lib/auth";
import { tokenExchange } from "../lib/api";

export default function Callback() {
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    async function handleCallback() {
      const params = new URLSearchParams(window.location.search);
      const code = params.get("code");
      const state = params.get("state");
      const errorParam = params.get("error");
      const errorDescription = params.get("error_description");

      if (errorParam) {
        setError(errorDescription || errorParam || "Authorization was denied.");
        return;
      }

      if (!code) {
        setError("No authorization code received.");
        return;
      }

      const storedState = getOAuthState();
      if (!storedState) {
        setError("Missing stored state. Please try logging in again.");
        return;
      }
      if (state !== storedState) {
        setError("State mismatch. Possible CSRF attack detected.");
        return;
      }

      const codeVerifier = getPkceVerifier();
      if (!codeVerifier) {
        setError("Missing PKCE code verifier. Please try again.");
        return;
      }

      const clientId = getClientId();
      if (!clientId) {
        setError("Missing Client ID. Please try logging in again.");
        return;
      }

      const redirectUri = window.location.origin + "/sample-app/callback";

      try {
        const body = new URLSearchParams({
          grant_type: "authorization_code",
          code: code,
          redirect_uri: redirectUri,
          client_id: clientId,
          code_verifier: codeVerifier,
        });

        const data = await tokenExchange(body);

        clearPkceState();
        setToken(data.access_token);
        if (data.refresh_token) {
          setRefreshToken(data.refresh_token);
        }

        navigate("/dashboard", { replace: true });
      } catch (err) {
        const message = err instanceof Error ? err.message : "Network error during token exchange. Please try again.";
        setError(message);
      }
    }

    handleCallback();
  }, [navigate]);

  if (error) {
    return (
      <div className="sample-layout">
        <nav className="sample-nav">
          <Link className="sample-nav-logo" to="/">
            <strong>sample app</strong>
          </Link>
        </nav>

        <div className="sample-hero">
          <div className="processing-card">
            <div className="alert alert-error" style={{ textAlign: "left" }}>
              <div>
                <strong>Authentication failed</strong>
                <div style={{ marginTop: 6 }}>{error}</div>
              </div>
            </div>
            <div style={{ marginTop: 20 }}>
              <Link to="/" className="btn btn-secondary">Back to home</Link>
            </div>
          </div>
        </div>
      </div>
    );
  }

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
          <div className="processing-title">Exchanging authorization code...</div>
          <div className="processing-subtitle">
            Please wait while we complete authentication.
          </div>
        </div>
      </div>
    </div>
  );
}
