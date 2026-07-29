import { describe, it, expect } from "vitest";
import { nodePresentation } from "../node-presentation";

describe("nodePresentation", () => {
  it("labels a managed addon as Postgres with a managed summary", () => {
    expect(nodePresentation({ isAddon: true })).toEqual({
      kindLabel: "Postgres",
      glyph: "postgres",
      brandSlug: "postgres",
      summary: "managed postgres",
      details: [],
    });
  });

  it("resolves a brand slug from the image for known software", () => {
    const slug = (image: string) => nodePresentation({ isAddon: false, image }).brandSlug;
    expect(slug("redis:6.2")).toBe("redis");
    expect(slug("postgres:16")).toBe("postgres");
    expect(slug("tooljet/tooljet:v3.20.18")).toBe("tooljet");
    expect(slug("grafana/otel-lgtm:0.9")).toBe("opentelemetry"); // otel wins over grafana
    expect(slug("grafana/grafana:11")).toBe("grafana");
    expect(slug("minio/minio:latest")).toBe("minio");
    expect(slug("acme/web-api:1")).toBeUndefined();
  });

  it("does not mistake postgrest for postgres", () => {
    const p = nodePresentation({
      isAddon: false,
      image: "postgrest/postgrest:v12.2.12",
      ports: [{ number: 3000, exposedToPublic: false }],
    });
    expect(p.kindLabel).toBe("Service");
    expect(p.brandSlug).toBe("postgrest");
  });

  it("detects redis and shows the uniform image · port · access line", () => {
    const p = nodePresentation({ isAddon: false, image: "redis:6.2", ports: [{ number: 6379 }] });
    expect(p.kindLabel).toBe("Redis");
    expect(p.glyph).toBe("redis");
    expect(p.summary).toBe("redis:6.2");
    expect(p.details).toEqual([{ text: "port 6379 · internal", port: 6379, public: false }]);
  });

  it("detects postgres from a service image", () => {
    const p = nodePresentation({ isAddon: false, image: "postgres:16" });
    expect(p.kindLabel).toBe("Postgres");
    expect(p.glyph).toBe("postgres");
    expect(p.summary).toBe("postgres:16");
  });

  it("detects mariadb/mysql as MySQL with a database glyph", () => {
    expect(nodePresentation({ isAddon: false, image: "mariadb:11" }).kindLabel).toBe("MySQL");
    expect(nodePresentation({ isAddon: false, image: "mysql:8" }).glyph).toBe("database");
  });

  it("detects minio as an Object store", () => {
    const p = nodePresentation({ isAddon: false, image: "minio/minio:latest" });
    expect(p).toMatchObject({ kindLabel: "Object", glyph: "object" });
    expect(p.summary).toBe("minio:latest");
  });

  it("treats a generic image with a public port as Web and shows :port · public", () => {
    const p = nodePresentation({
      isAddon: false,
      image: "acme/web-api:1.2.3",
      ports: [{ number: 8080, protocol: "http", exposedToPublic: true }],
    });
    expect(p.kindLabel).toBe("Web");
    expect(p.glyph).toBe("web");
    expect(p.summary).toBe("web-api:1.2.3");
    expect(p.details).toEqual([{ text: "port 8080 · public", port: 8080, public: true }]);
  });

  it("treats a generic image with only an internal port as a Service", () => {
    const p = nodePresentation({
      isAddon: false,
      image: "acme/mailhog",
      ports: [{ number: 1025, exposedToPublic: false }],
    });
    expect(p.kindLabel).toBe("Service");
    expect(p.glyph).toBe("service");
    expect(p.summary).toBe("mailhog");
    expect(p.details).toEqual([{ text: "port 1025 · internal", port: 1025, public: false }]);
  });

  it("falls back to git build when there is no image", () => {
    const p = nodePresentation({ isAddon: false, hasBuild: true });
    expect(p.kindLabel).toBe("Service");
    expect(p.summary).toBe("git build");
  });

  it("lists every declared port, one line each, in declared order", () => {
    const p = nodePresentation({
      isAddon: false,
      image: "grafana/otel-lgtm:0.29.1",
      ports: [
        { number: 4318, exposedToPublic: false },
        { number: 3000, exposedToPublic: true },
      ],
    });
    expect(p.details).toEqual([{ text: "port 4318 · internal", port: 4318, public: false }, { text: "port 3000 · public", port: 3000, public: true }]);
  });

  it("returns every declared port — the card collapses them to one line itself", () => {
    const p = nodePresentation({
      isAddon: false,
      image: "acme/kitchen-sink",
      ports: [
        { number: 80, exposedToPublic: true },
        { number: 443, exposedToPublic: true },
        { number: 9090, exposedToPublic: false },
        { number: 9091, exposedToPublic: false },
        { number: 9092, exposedToPublic: false },
      ],
    });
    expect(p.details.map((d) => d.port)).toEqual([80, 443, 9090, 9091, 9092]);
  });

  it("strips the registry but keeps the tag in the summary", () => {
    const p = nodePresentation({
      isAddon: false,
      image: "registry.example.com:5000/project/frontend:v2",
      ports: [{ number: 3000, exposedToPublic: true }],
    });
    expect(p.summary).toBe("frontend:v2");
    expect(p.details).toEqual([{ text: "port 3000 · public", port: 3000, public: true }]);
  });
});
