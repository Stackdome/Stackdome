import { describe, it, expect } from "vitest";
import { mapFieldErrors, fieldTab } from "./map-field-errors";
import type { ParsedFieldError } from "@/api/errors";

function fe(field: string, message = "msg", code = "x"): ParsedFieldError {
  return { field, code, message };
}

describe("mapFieldErrors — fat dialect", () => {
  it("strips the spec.stack_resources[N]. prefix and converts brackets to dots", () => {
    const out = mapFieldErrors([fe("spec.stack_resources[0].ports[1].protocol", "bad protocol")], { dialect: "fat" });
    expect(out.resources[0]).toEqual({ "ports.1.protocol": "bad protocol" });
  });

  it("routes bare `name` to the stack name, not a resource", () => {
    const out = mapFieldErrors([fe("name", "stack name invalid")], { dialect: "fat" });
    expect(out.stackName).toBe("stack name invalid");
    expect(out.resources).toEqual({});
  });

  it("routes spec.stack_resources[N].name to that resource's name", () => {
    const out = mapFieldErrors([fe("spec.stack_resources[2].name", "dup")], { dialect: "fat" });
    expect(out.resources[2]).toEqual({ name: "dup" });
    expect(out.stackName).toBeUndefined();
  });

  it("translates env paths to the frontend env-tab key with .name / .value leaf", () => {
    const out = mapFieldErrors(
      [
        fe("spec.stack_resources[0].execution_config.env[2].name", "name required"),
        fe("spec.stack_resources[0].execution_config.env[3]", "value missing"),
        fe("spec.stack_resources[0].execution_config.env[4].secret_key_ref", "secret missing"),
        fe("spec.stack_resources[0].execution_config.env[5].self_output", "unknown output"),
      ],
      { dialect: "fat" },
    );
    expect(out.resources[0]).toEqual({
      "execution_config.environment_variables.2.name": "name required",
      "execution_config.environment_variables.3.value": "value missing",
      "execution_config.environment_variables.4.value": "secret missing",
      "execution_config.environment_variables.5.value": "unknown output",
    });
  });

  it("collects settings and connections separately", () => {
    const out = mapFieldErrors(
      [fe("spec.settings", "retention too high"), fe("spec.connections[0]", "bad connection")],
      { dialect: "fat" },
    );
    expect(out.settings).toEqual(["retention too high"]);
    expect(out.connections).toEqual(["bad connection"]);
  });

  it("puts unrecognized paths in unmapped", () => {
    const err = fe("spec.something_new", "huh");
    const out = mapFieldErrors([err], { dialect: "fat" });
    expect(out.unmapped).toEqual([err]);
  });
});

describe("mapFieldErrors — thin dialect", () => {
  it("assigns every bare path to the edited resource's index", () => {
    const out = mapFieldErrors(
      [fe("name", "resource name required"), fe("ports[0].number", "bad number")],
      { dialect: "thin", resourceIndex: 3 },
    );
    expect(out.resources[3]).toEqual({ name: "resource name required", "ports.0.number": "bad number" });
    expect(out.stackName).toBeUndefined();
  });

  it("translates thin env paths too", () => {
    const out = mapFieldErrors([fe("execution_config.env[0].name", "required")], { dialect: "thin", resourceIndex: 1 });
    expect(out.resources[1]).toEqual({ "execution_config.environment_variables.0.name": "required" });
  });

  it("routes an unexpected spec-prefixed path to unmapped", () => {
    const err = fe("spec.settings", "unexpected");
    const out = mapFieldErrors([err], { dialect: "thin", resourceIndex: 0 });
    expect(out.unmapped).toEqual([err]);
    expect(out.resources).toEqual({});
  });
});

describe("fieldTab", () => {
  it("routes env paths to the environment tab", () => {
    expect(fieldTab("execution_config.env[0].name")).toBe("environment");
  });

  it("routes everything else to configuration", () => {
    expect(fieldTab("source.image.ref")).toBe("configuration");
    expect(fieldTab("ports[0].protocol")).toBe("configuration");
  });
});
