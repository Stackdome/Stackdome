import { describe, it, expect } from "vitest";
import { secretOutputAccessor, parseSecretOutput } from "../connection-mapping";

describe("secretOutputAccessor", () => {
  it("uses dot form for simple keys", () => {
    expect(secretOutputAccessor("LOCKBOX_MASTER_KEY")).toBe("key.LOCKBOX_MASTER_KEY");
    expect(secretOutputAccessor("a1_B2")).toBe("key.a1_B2");
    // Backend `^[A-Za-z0-9_]+$` treats digit-leading keys as simple.
    expect(secretOutputAccessor("1leading")).toBe("key.1leading");
  });

  it("uses bracket form for keys with special characters", () => {
    expect(secretOutputAccessor("my-key")).toBe("key['my-key']");
    expect(secretOutputAccessor("has space")).toBe("key['has space']");
  });

  it("escapes single quotes and backslashes in bracket form", () => {
    expect(secretOutputAccessor("a'b")).toBe("key['a\\'b']");
    expect(secretOutputAccessor("a\\b")).toBe("key['a\\\\b']");
  });
});

describe("parseSecretOutput", () => {
  it("reverses the dot form", () => {
    expect(parseSecretOutput("key.LOCKBOX_MASTER_KEY")).toBe("LOCKBOX_MASTER_KEY");
    expect(parseSecretOutput("key.1leading")).toBe("1leading");
  });
  it("reverses the bracket form, unescaping", () => {
    expect(parseSecretOutput("key['my-key']")).toBe("my-key");
    expect(parseSecretOutput("key['a\\'b']")).toBe("a'b");
    expect(parseSecretOutput("key['a\\\\b']")).toBe("a\\b");
  });
  it("returns null for unrecognized accessors", () => {
    expect(parseSecretOutput("host")).toBeNull();
    expect(parseSecretOutput("")).toBeNull();
    expect(parseSecretOutput("key")).toBeNull();
    expect(parseSecretOutput("key.")).toBeNull();
    expect(parseSecretOutput("key['unclosed")).toBeNull();
  });

  it("round-trips every key form", () => {
    const keys = [
      "LOCKBOX_MASTER_KEY",
      "tls.crt",
      "my-key",
      "has space",
      "a'b",
      "a\\b",
      "1leading",
    ];
    for (const k of keys) {
      expect(parseSecretOutput(secretOutputAccessor(k))).toBe(k);
    }
  });
});
