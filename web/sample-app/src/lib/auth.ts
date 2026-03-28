const CLIENT_ID_KEY = "sample_client_id";
const ACCESS_TOKEN_KEY = "sample_access_token";
const REFRESH_TOKEN_KEY = "sample_refresh_token";
const PKCE_VERIFIER_KEY = "sample_pkce_verifier";
const OAUTH_STATE_KEY = "sample_oauth_state";

export function getToken(): string | null {
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(ACCESS_TOKEN_KEY, token);
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

export function setRefreshToken(token: string): void {
  localStorage.setItem(REFRESH_TOKEN_KEY, token);
}

export function clearAuth(): void {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
}

export function isAuthenticated(): boolean {
  return getToken() !== null;
}

export function setPkceVerifier(verifier: string): void {
  sessionStorage.setItem(PKCE_VERIFIER_KEY, verifier);
}

export function getPkceVerifier(): string | null {
  return sessionStorage.getItem(PKCE_VERIFIER_KEY);
}

export function setOAuthState(state: string): void {
  sessionStorage.setItem(OAUTH_STATE_KEY, state);
}

export function getOAuthState(): string | null {
  return sessionStorage.getItem(OAUTH_STATE_KEY);
}

export function clearPkceState(): void {
  sessionStorage.removeItem(PKCE_VERIFIER_KEY);
  sessionStorage.removeItem(OAUTH_STATE_KEY);
}

export function getClientId(): string | null {
  return localStorage.getItem(CLIENT_ID_KEY);
}

export function setClientId(clientId: string): void {
  localStorage.setItem(CLIENT_ID_KEY, clientId);
}
