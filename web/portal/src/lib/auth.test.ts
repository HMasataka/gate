import { describe, it, expect, beforeEach } from "vitest";
import {
  getToken,
  setToken,
  setEmail,
  getEmail,
  clearAuth,
  isAuthenticated,
} from "./auth";

describe("auth", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("returns null when no token is stored", () => {
    expect(getToken()).toBeNull();
  });

  it("stores and retrieves a token", () => {
    setToken("test-token-123");
    expect(getToken()).toBe("test-token-123");
  });

  it("stores and retrieves an email", () => {
    setEmail("user@example.com");
    expect(getEmail()).toBe("user@example.com");
  });

  it("returns empty string when no email is stored", () => {
    expect(getEmail()).toBe("");
  });

  it("clears auth data", () => {
    setToken("some-token");
    setEmail("user@example.com");

    clearAuth();

    expect(getToken()).toBeNull();
    expect(getEmail()).toBe("");
  });

  it("reports authenticated when token exists", () => {
    expect(isAuthenticated()).toBe(false);
    setToken("token");
    expect(isAuthenticated()).toBe(true);
  });
});
