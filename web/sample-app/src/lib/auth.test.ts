import { describe, it, expect, beforeEach } from "vitest";
import {
  getToken,
  setToken,
  getRefreshToken,
  setRefreshToken,
  clearAuth,
  isAuthenticated,
  setPkceVerifier,
  getPkceVerifier,
  setOAuthState,
  getOAuthState,
  clearPkceState,
} from "./auth";

describe("auth - localStorage token management", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("getToken returns null when no token is stored", () => {
    expect(getToken()).toBeNull();
  });

  it("setToken and getToken round-trip", () => {
    setToken("abc123");
    expect(getToken()).toBe("abc123");
  });

  it("setRefreshToken and getRefreshToken round-trip", () => {
    setRefreshToken("refresh_xyz");
    expect(getRefreshToken()).toBe("refresh_xyz");
  });

  it("clearAuth removes both tokens", () => {
    setToken("abc");
    setRefreshToken("xyz");
    clearAuth();
    expect(getToken()).toBeNull();
    expect(getRefreshToken()).toBeNull();
  });

  it("isAuthenticated returns false when no token", () => {
    expect(isAuthenticated()).toBe(false);
  });

  it("isAuthenticated returns true when token exists", () => {
    setToken("token");
    expect(isAuthenticated()).toBe(true);
  });
});

describe("auth - sessionStorage PKCE management", () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  it("getPkceVerifier returns null when not set", () => {
    expect(getPkceVerifier()).toBeNull();
  });

  it("setPkceVerifier and getPkceVerifier round-trip", () => {
    setPkceVerifier("verifier123");
    expect(getPkceVerifier()).toBe("verifier123");
  });

  it("getOAuthState returns null when not set", () => {
    expect(getOAuthState()).toBeNull();
  });

  it("setOAuthState and getOAuthState round-trip", () => {
    setOAuthState("state_abc");
    expect(getOAuthState()).toBe("state_abc");
  });

  it("clearPkceState removes verifier and state", () => {
    setPkceVerifier("v");
    setOAuthState("s");
    clearPkceState();
    expect(getPkceVerifier()).toBeNull();
    expect(getOAuthState()).toBeNull();
  });
});
