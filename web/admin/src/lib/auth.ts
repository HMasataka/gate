const TOKEN_KEY = "admin_access_token";
const EMAIL_KEY = "admin_user_email";

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function setEmail(email: string): void {
  localStorage.setItem(EMAIL_KEY, email);
}

export function getEmail(): string {
  return localStorage.getItem(EMAIL_KEY) ?? "";
}

export function clearAuth(): void {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(EMAIL_KEY);
}

export function isAuthenticated(): boolean {
  return getToken() !== null;
}
