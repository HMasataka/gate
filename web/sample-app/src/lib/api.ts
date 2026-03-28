import { getToken, clearAuth } from "./auth";

export class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

export async function api<T>(
  method: string,
  path: string,
  body?: unknown,
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  const token = getToken();
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  if (res.status === 401) {
    clearAuth();
    window.location.href = "/sample-app/";
    throw new ApiError("Unauthorized", 401);
  }

  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    const message =
      data.error?.message ?? data.message ?? data.error_description ?? `HTTP ${res.status}`;
    throw new ApiError(message, res.status);
  }

  if (res.status === 204) {
    return null as T;
  }

  return res.json();
}

interface TokenResponse {
  access_token: string;
  refresh_token?: string;
  token_type: string;
  expires_in?: number;
}

export async function tokenExchange(params: URLSearchParams): Promise<TokenResponse> {
  const res = await fetch("/api/v1/oauth/token", {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: params.toString(),
  });

  const data = await res.json().catch(() => ({}));

  if (!res.ok) {
    const message =
      data.error_description ?? data.error ?? data.message ?? "Token exchange failed.";
    throw new ApiError(message, res.status);
  }

  if (!data.access_token) {
    throw new ApiError("No access token received from server.", 0);
  }

  return data as TokenResponse;
}
