import { describe, it, expect } from "vitest";
import { providerIdForHost, REGISTRY_PROVIDERS, PURPOSE_LABELS, PURPOSE_BOTH } from "../providers";

describe("providerIdForHost", () => {
  it("maps known hosts to providers", () => {
    expect(providerIdForHost("docker.io")).toBe("dockerhub");
    expect(providerIdForHost("index.docker.io")).toBe("dockerhub");
    expect(providerIdForHost("ghcr.io")).toBe("ghcr");
    expect(providerIdForHost("registry.gitlab.com")).toBe("gitlab");
    expect(providerIdForHost("quay.io")).toBe("quay");
  });

  it("falls back to other for unknown or missing hosts", () => {
    expect(providerIdForHost("registry.example.com")).toBe("other");
    expect(providerIdForHost(undefined)).toBe("other");
  });
});

describe("REGISTRY_PROVIDERS", () => {
  it("prefills the documented hosts", () => {
    const byId = Object.fromEntries(REGISTRY_PROVIDERS.map((p) => [p.id, p.hostPrefill]));
    expect(byId.dockerhub).toBe("docker.io");
    expect(byId.ghcr).toBe("ghcr.io");
    expect(byId.gitlab).toBe("registry.gitlab.com");
    expect(byId.quay).toBe("quay.io");
    expect(byId.other).toBe("");
  });
});

describe("PURPOSE_LABELS", () => {
  it("labels every purpose", () => {
    expect(PURPOSE_LABELS[PURPOSE_BOTH]).toBe("Pull & push");
    expect(PURPOSE_LABELS.pull).toBe("Pull only");
    expect(PURPOSE_LABELS.push).toBe("Push only");
  });
});
