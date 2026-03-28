import { describe, it, expect } from "vitest";
import { base64urlEncode, generatePKCE, generateState } from "./pkce";

describe("base64urlEncode", () => {
  it("encodes an ArrayBuffer to base64url without padding", () => {
    const buffer = new Uint8Array([72, 101, 108, 108, 111]).buffer;
    const result = base64urlEncode(buffer);
    expect(result).toBe("SGVsbG8");
    expect(result).not.toMatch(/[+/=]/);
  });

  it("replaces + with - and / with _", () => {
    const buffer = new Uint8Array([251, 255, 254]).buffer;
    const result = base64urlEncode(buffer);
    expect(result).not.toContain("+");
    expect(result).not.toContain("/");
    expect(result).not.toContain("=");
  });
});

describe("generatePKCE", () => {
  it("returns verifier and challenge as non-empty strings", async () => {
    const { verifier, challenge } = await generatePKCE();
    expect(verifier).toBeTruthy();
    expect(challenge).toBeTruthy();
    expect(typeof verifier).toBe("string");
    expect(typeof challenge).toBe("string");
  });

  it("returns different verifiers on each call", async () => {
    const a = await generatePKCE();
    const b = await generatePKCE();
    expect(a.verifier).not.toBe(b.verifier);
  });

  it("verifier and challenge are base64url-safe", async () => {
    const { verifier, challenge } = await generatePKCE();
    expect(verifier).toMatch(/^[A-Za-z0-9_-]+$/);
    expect(challenge).toMatch(/^[A-Za-z0-9_-]+$/);
  });
});

describe("generateState", () => {
  it("returns a non-empty string", () => {
    const state = generateState();
    expect(state).toBeTruthy();
    expect(typeof state).toBe("string");
  });

  it("returns different values on each call", () => {
    const a = generateState();
    const b = generateState();
    expect(a).not.toBe(b);
  });

  it("is base64url-safe", () => {
    const state = generateState();
    expect(state).toMatch(/^[A-Za-z0-9_-]+$/);
  });
});
