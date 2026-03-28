import { generatePKCE, generateState } from "../lib/pkce";
import { setPkceVerifier, setOAuthState } from "../lib/auth";

const CLIENT_ID = "sample-client";

async function startLogin() {
  const { verifier, challenge } = await generatePKCE();
  const state = generateState();

  setPkceVerifier(verifier);
  setOAuthState(state);

  const redirectUri = window.location.origin + "/sample-app/callback";

  const params = new URLSearchParams({
    response_type: "code",
    client_id: CLIENT_ID,
    redirect_uri: redirectUri,
    scope: "openid email profile",
    code_challenge: challenge,
    code_challenge_method: "S256",
    state: state,
  });

  window.location.href = `/api/v1/oauth/authorize?${params.toString()}`;
}

export default function Home() {
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
            This is a demonstration of the gate OAuth 2.0 flow with PKCE. Click
            the button below to authenticate via gate.
          </p>
          <button
            className="btn btn-primary"
            onClick={startLogin}
            style={{ fontSize: 14, padding: "12px 28px" }}
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
