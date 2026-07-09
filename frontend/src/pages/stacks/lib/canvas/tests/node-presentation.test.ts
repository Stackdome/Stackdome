import { describe, it, expect } from "vitest";
import { nodePresentation } from "../node-presentation";

describe("nodePresentation", () => {
  it("labels a managed addon as Postgres with a managed summary", () => {
    expect(nodePresentation({ isAddon: true })).toEqual({
      kindLabel: "Postgres",
      glyph: "postgres",
      summary: "managed postgres",
    });
  });

  it("detects redis and marks it in-memory", () => {
    const p = nodePresentation({ isAddon: false, image: "redis:6.2", ports: [{ number: 6379 }] });
    expect(p.kindLabel).toBe("Redis");
    expect(p.glyph).toBe("redis");
    expect(p.summary).toBe("redis:6.2 · in-memory");
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

  it("detects minio as an Object store, S3-compatible", () => {
    const p = nodePresentation({ isAddon: false, image: "minio/minio:latest" });
    expect(p).toMatchObject({ kindLabel: "Object", glyph: "object" });
    expect(p.summary).toBe("minio/minio:latest · S3-compatible");
  });

  it("treats a generic image with a public port as Web and shows :port · public", () => {
    const p = nodePresentation({
      isAddon: false,
      image: "acme/web-api:1.2.3",
      ports: [{ number: 8080, protocol: "http", exposedToPublic: true }],
    });
    expect(p.kindLabel).toBe("Web");
    expect(p.glyph).toBe("web");
    expect(p.summary).toBe("web-api · :8080 · public");
  });

  it("treats a generic image with only an internal port as a Service", () => {
    const p = nodePresentation({
      isAddon: false,
      image: "acme/mailhog",
      ports: [{ number: 1025, exposedToPublic: false }],
    });
    expect(p.kindLabel).toBe("Service");
    expect(p.glyph).toBe("service");
    expect(p.summary).toBe("acme/mailhog");
  });

  it("falls back to git build when there is no image", () => {
    const p = nodePresentation({ isAddon: false, hasBuild: true });
    expect(p.kindLabel).toBe("Service");
    expect(p.summary).toBe("git build");
  });

  it("strips the registry and tag when building a Web summary base", () => {
    const p = nodePresentation({
      isAddon: false,
      image: "registry.example.com:5000/team/frontend:v2",
      ports: [{ number: 3000, exposedToPublic: true }],
    });
    expect(p.summary).toBe("frontend · :3000 · public");
  });
});
