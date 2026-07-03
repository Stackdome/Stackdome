import { describe, it, expect } from "vitest";
import { classifyIngressUrl, pickBestIngress } from "../public-endpoints";

const ORG = ["acme.stackdome.app"];
const GENERATED = "https://k7x2m3qp4rt6w3ab.web.acme.stackdome.app";
const PREFIX = "https://web.acme.stackdome.app";
const CUSTOM = "https://app.mycompany.com";

describe("classifyIngressUrl", () => {
  it("host outside org domains → custom", () => {
    expect(classifyIngressUrl(CUSTOM, ORG)).toBe("custom");
  });
  it("16-char base32 first label under org domain → generated", () => {
    expect(classifyIngressUrl(GENERATED, ORG)).toBe("generated");
  });
  it("friendly first label under org domain → prefix", () => {
    expect(classifyIngressUrl(PREFIX, ORG)).toBe("prefix");
  });
  it("unparseable url → custom (shown as-is, never hidden)", () => {
    expect(classifyIngressUrl("not a url", ORG)).toBe("custom");
  });
});

describe("pickBestIngress", () => {
  it("prefers custom > prefix > generated", () => {
    const ingresses = [
      { url: GENERATED, target_port: 80 },
      { url: PREFIX, target_port: 80 },
      { url: CUSTOM, target_port: 80 },
    ];
    expect(pickBestIngress(ingresses, ORG)?.url).toBe(CUSTOM);
    expect(pickBestIngress(ingresses.slice(0, 2), ORG)?.url).toBe(PREFIX);
    expect(pickBestIngress(ingresses.slice(0, 1), ORG)?.url).toBe(GENERATED);
  });
  it("breaks ties by array order and skips url-less entries", () => {
    expect(pickBestIngress([{ target_port: 1 }, { url: PREFIX }, { url: "https://web2.acme.stackdome.app" }], ORG)?.url).toBe(PREFIX);
  });
  it("returns null when nothing has a url", () => {
    expect(pickBestIngress([{}, { target_port: 80 }], ORG)).toBeNull();
  });
});
