import { describe, it, expect, beforeEach, vi } from "vitest";
import { api, ApiError } from "./api";

describe("api", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it("sends a GET request and returns JSON", async () => {
    const mockData = { clients: [] };
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockData),
    });

    const result = await api("GET", "/api/v1/clients");
    expect(result).toEqual(mockData);
    expect(fetch).toHaveBeenCalledWith("/api/v1/clients", {
      method: "GET",
      headers: { "Content-Type": "application/json" },
      body: undefined,
    });
  });

  it("includes authorization header when token exists", async () => {
    localStorage.setItem("portal_access_token", "my-token");

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({}),
    });

    await api("GET", "/api/v1/test");

    expect(fetch).toHaveBeenCalledWith("/api/v1/test", {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer my-token",
      },
      body: undefined,
    });
  });

  it("sends request body for POST", async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ id: "1" }),
    });

    const body = { name: "test" };
    await api("POST", "/api/v1/clients", body);

    expect(fetch).toHaveBeenCalledWith("/api/v1/clients", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
  });

  it("throws ApiError on non-ok response", async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 400,
      json: () => Promise.resolve({ message: "Bad request" }),
    });

    await expect(api("GET", "/api/v1/fail")).rejects.toThrow(ApiError);
    await expect(api("GET", "/api/v1/fail")).rejects.toThrow("Bad request");
  });

  it("returns null for 204 responses", async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
    });

    const result = await api("DELETE", "/api/v1/clients/1");
    expect(result).toBeNull();
  });
});
