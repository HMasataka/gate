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

    const result = await api("GET", "/api/v1/admin/clients");
    expect(result).toEqual(mockData);
  });

  it("includes authorization header when token exists", async () => {
    localStorage.setItem("admin_access_token", "admin-token");

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({}),
    });

    await api("GET", "/api/v1/admin/users");

    expect(fetch).toHaveBeenCalledWith("/api/v1/admin/users", {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer admin-token",
      },
      body: undefined,
    });
  });

  it("throws ApiError on non-ok response", async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 403,
      json: () => Promise.resolve({ message: "Forbidden" }),
    });

    await expect(api("GET", "/api/v1/admin/fail")).rejects.toThrow(ApiError);
    await expect(api("GET", "/api/v1/admin/fail")).rejects.toThrow("Forbidden");
  });

  it("returns null for 204 responses", async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
    });

    const result = await api("DELETE", "/api/v1/admin/users/1");
    expect(result).toBeNull();
  });
});
