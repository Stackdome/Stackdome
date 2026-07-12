import { describe, it, expect } from "vitest";
import { buildGitSeed } from "../git-source-seed";

const repo = {
  fullName: "acme/My_Web.App",
  cloneUrl: "https://github.com/acme/My_Web.App.git",
  defaultBranch: "main",
  integrationId: "int-app",
};

describe("buildGitSeed", () => {
  it("builds a one-service seed with a git build source and sanitized names", () => {
    const seed = buildGitSeed(repo, {
      serviceName: "my-web-app",
      branch: "main",
      dockerfilePath: "Dockerfile",
      buildContext: ".",
      port: 3000,
      exposePublic: true,
    });
    expect(seed.name).toBe("my-web-app");
    expect(seed.resources).toHaveLength(1);
    const r = seed.resources[0];
    expect(r.name).toBe("my-web-app");
    expect(r.workload_type).toBe("Service");
    expect(r.sourceType).toBe("git");
    expect(r.gitRevisionType).toBe("branch");
    expect(r.gitRevisionValue).toBe("main");
    expect(r.source).toEqual({
      git: {
        repo_url: "https://github.com/acme/My_Web.App.git",
        dockerfile_path: "Dockerfile",
        build_context: ".",
        integration_id: "int-app",
      },
    });
    expect(r.ports).toEqual([
      { name: "http-3000", number: 3000, protocol: "http", exposed_to_public: true },
    ]);
    expect(seed.volumes).toEqual([]);
    expect(seed.linkedAddonIds).toEqual([]);
  });

  it("derives a sanitized default service name from the repo", () => {
    expect(buildGitSeed(repo, {
      serviceName: "",
      branch: "main",
      dockerfilePath: "Dockerfile",
      buildContext: ".",
      port: 8080,
      exposePublic: false,
    }).resources[0].name).toBe("my-web-app");
  });

  it("omits integration_id when the repo has no integration (public URL)", () => {
    const publicRepo = { ...repo, integrationId: null };
    const seed = buildGitSeed(publicRepo, {
      serviceName: "my-web-app",
      branch: "main",
      dockerfilePath: "Dockerfile",
      buildContext: ".",
      port: 3000,
      exposePublic: true,
    });
    expect(seed.resources[0].source.git.integration_id).toBeUndefined();
  });
});
