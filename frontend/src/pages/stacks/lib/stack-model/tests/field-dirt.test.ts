import { describe, it, expect } from "vitest";
import { isFieldDirty } from "../field-dirt";

/**
 * A drawer field tints when it differs from the baseline. Two paths reach that
 * answer: a named resource canonicalizes and compares in canonical terms;
 * anything else falls back to a raw compare at the path.
 */

const gitResource = (git: Record<string, unknown> = {}, rest: Record<string, unknown> = {}) => ({
  name: "worker",
  sourceType: "git" as const,
  source: {
    git: {
      repo_url: "https://github.com/acme/demo.git",
      dockerfile_path: "Dockerfile",
      build_context: ".",
      ...git,
    },
  },
  ...rest,
});

describe("isFieldDirty on a named resource", () => {
  it("does not tint a build path the API defaulted and the form spelled out", () => {
    const fromServer = {
      name: "worker",
      sourceType: "git" as const,
      source: { git: { repo_url: "https://github.com/acme/demo.git" } },
    };
    expect(isFieldDirty(gitResource(), fromServer, "source.git.dockerfile_path")).toBe(false);
    expect(isFieldDirty(gitResource(), fromServer, "source.git.build_context")).toBe(false);
  });

  it("tints a build path the user actually changed", () => {
    expect(
      isFieldDirty(gitResource({ dockerfile_path: "docker/Dockerfile" }), gitResource(), "source.git.dockerfile_path"),
    ).toBe(true);
  });

  it("maps the revision helpers onto whichever of branch/tag carries the value", () => {
    const onBranch = gitResource({}, { gitRevisionType: "branch", gitRevisionValue: "main" });
    const onTag = gitResource({}, { gitRevisionType: "tag", gitRevisionValue: "v1" });
    expect(isFieldDirty(onBranch, onBranch, "gitRevisionValue")).toBe(false);
    expect(isFieldDirty(onTag, onBranch, "gitRevisionValue")).toBe(true);
  });

  it("tints the commit pin only when it changes", () => {
    const pinned = gitResource({}, {
      gitRevisionType: "branch",
      gitRevisionValue: "main",
      gitCommitPin: "abc123",
    });
    expect(isFieldDirty(pinned, pinned, "gitCommitPin")).toBe(false);
    expect(isFieldDirty({ ...pinned, gitCommitPin: undefined }, pinned, "gitCommitPin")).toBe(true);
  });

  /**
   * Mounts are keyed by target path in canonical terms but addressed by row
   * index in the drawer, so the index resolves against the draft.
   */
  describe("volume mounts", () => {
    const mounted = (rows: Array<Record<string, unknown>>) =>
      gitResource({}, { volume_mounts: rows });
    const data = { source_volume_name: "data", source_sub_path: "", target_path: "/data" };
    const logs = { source_volume_name: "logs", source_sub_path: "", target_path: "/logs" };

    it("tints the row whose volume changed", () => {
      const draft = mounted([{ ...data, source_volume_name: "data-v2" }]);
      expect(isFieldDirty(draft, mounted([data]), "volume_mounts.0")).toBe(true);
    });

    it("tints the row whose sub path changed", () => {
      const draft = mounted([{ ...data, source_sub_path: "inner" }]);
      expect(isFieldDirty(draft, mounted([data]), "volume_mounts.0")).toBe(true);
    });

    it("tints the row that was retargeted", () => {
      const draft = mounted([{ ...data, target_path: "/var/data" }]);
      expect(isFieldDirty(draft, mounted([data]), "volume_mounts.0")).toBe(true);
    });

    it("leaves the untouched sibling row alone", () => {
      const draft = mounted([{ ...data, source_volume_name: "data-v2" }, logs]);
      expect(isFieldDirty(draft, mounted([data, logs]), "volume_mounts.1")).toBe(false);
    });

    /** Guards the exact-match lookup: a prefix test would let /data light /data.bak. */
    it("does not let a mount light its dotted neighbour", () => {
      const bak = { ...data, source_volume_name: "backup", target_path: "/data.bak" };
      const draft = mounted([{ ...data, source_volume_name: "data-v2" }, bak]);
      expect(isFieldDirty(draft, mounted([data, bak]), "volume_mounts.1")).toBe(false);
    });

    it("does not treat a reorder as an edit", () => {
      expect(isFieldDirty(mounted([logs, data]), mounted([data, logs]), "volume_mounts.0")).toBe(false);
    });

    it("says nothing about a half-typed row that has no target yet", () => {
      const draft = mounted([data, { source_volume_name: "logs" }]);
      expect(isFieldDirty(draft, mounted([data]), "volume_mounts.1")).toBe(false);
    });

    it("stays clean when nothing about the mounts changed", () => {
      expect(isFieldDirty(mounted([data]), mounted([data]), "volume_mounts.0")).toBe(false);
    });
  });

  it("lights the source toggle for a change of kind, not a change inside the source", () => {
    const asImage = { name: "worker", sourceType: "image" as const, source: { image: { ref: "nginx:1" } } };
    expect(
      isFieldDirty(gitResource({ repo_url: "https://other/repo.git" }), gitResource(), "sourceType"),
    ).toBe(false);
    expect(isFieldDirty(asImage, gitResource(), "sourceType")).toBe(true);
  });
});

/**
 * The form and the API disagree constantly about how an absent value is
 * spelled. None of those spellings is an edit. These carry no name, so they
 * exercise the raw fallback.
 */
describe("isFieldDirty — structurally-empty equivalence", () => {
  const equivalent: Array<[string, object, object, string]> = [
    ["empty array vs undefined", { init_spec: { command: [] } }, { init_spec: undefined }, "init_spec.command"],
    ["empty object vs undefined", { execution_config: {} }, { execution_config: undefined }, "execution_config"],
    ["empty string vs undefined", { name: "" }, { name: undefined }, "name"],
    [
      "deeply-nested all-empty vs undefined",
      { init_spec: { command: [], args: "" } },
      { init_spec: undefined },
      "init_spec",
    ],
  ];

  for (const [label, draft, baseline, path] of equivalent) {
    it(`treats ${label} as not dirty`, () => {
      expect(isFieldDirty(draft as never, baseline as never, path)).toBe(false);
    });
  }

  it("flags a real value against an empty one", () => {
    expect(
      isFieldDirty({ init_spec: { command: ["migrate"] } } as never, { init_spec: undefined } as never, "init_spec"),
    ).toBe(true);
  });

  it("flags a value that differs from a non-empty baseline", () => {
    expect(isFieldDirty({ name: "" } as never, { name: "redis" } as never, "name")).toBe(true);
  });
});
