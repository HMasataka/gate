import { describe, it, expect, beforeEach, vi } from "vitest";
import { api, tokenExchange, ApiError } from "./api";
import * as auth from "./auth";

describe("api", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it("sends GET request with Bearer token when authenticated", async () => {
    auth.setToken("test-token");

    const mockResponse = { sub: "user1", email: "test@example.com" };
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify(mockResponse), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    const result = await api<typeof mockResponse>("GET", "/oauth/userinfo");

    expect(result).toEqual(mockResponse);
    expect(fetch).toHaveBeenCalledWith("/oauth/userinfo", {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer test-token",
      },
      body: undefined,
    });
  });

  it("sends request without Authorization when no token", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), { status: 200 }),
    );

    await api("GET", "/some/path");

    const callArgs = vi.mocked(fetch).mock.calls[0];
    const headers = callArgs[1]?.headers as Record<string, string>;
    expect(headers["Authorization"]).toBeUndefined();
  });

  it("throws ApiError on non-ok response", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ message: "Not found" }), { status: 404 }),
    );

    await expect(api("GET", "/missing")).rejects.toThrow(ApiError);
    await expect(
      api("GET", "/missing").catch((e: ApiError) => {
        expect(e.status).toBe(404);
        throw e;
      }),
    ).rejects.toThrow();
  });

  it("returns null for 204 response", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(null, { status: 204 }),
    );

    const result = await api("DELETE", "/resource");
    expect(result).toBeNull();
  });
});

describe("tokenExchange", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("sends form-urlencoded POST and returns token data", async () => {
    const tokenData = {
      access_token: "at_123",
      refresh_token: "rt_456",
      token_type: "Bearer",
      expires_in: 3600,
    };

    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify(tokenData), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    const params = new URLSearchParams({
      grant_type: "authorization_code",
      code: "test_code",
      redirect_uri: "http://localhost/callback",
      client_id: "sample-client",
      code_verifier: "test_verifier",
    });

    const result = await tokenExchange(params);

    expect(result.access_token).toBe("at_123");
    expect(result.refresh_token).toBe("rt_456");
    expect(fetch).toHaveBeenCalledWith("/api/v1/oauth/token", {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: params.toString(),
    });
  });

  it("throws ApiError on failure", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ error: "invalid_grant" }), { status: 400 }),
    );

    const params = new URLSearchParams({ grant_type: "authorization_code" });
    await expect(tokenExchange(params)).rejects.toThrow(ApiError);
  });

  it("throws ApiError when no access_token in response", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ token_type: "Bearer" }), { status: 200 }),
    );

    const params = new URLSearchParams({ grant_type: "authorization_code" });
    await expect(tokenExchange(params)).rejects.toThrow(
      "No access token received from server.",
    );
  });
});
