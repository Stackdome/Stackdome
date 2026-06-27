import { describe, it, expect, vi } from "vitest";
import { buildImportOptions } from "../import-options";

describe("buildImportOptions", () => {
  it("includes an enabled Templates option wired to onTemplates", () => {
    const onTemplates = vi.fn();
    const onDockerCompose = vi.fn();
    const opts = buildImportOptions({ onDockerCompose, onTemplates });

    const templates = opts.find((o) => o.id === "templates");
    expect(templates).toBeDefined();
    expect(templates!.disabled).toBeFalsy();

    templates!.onClick();
    expect(onTemplates).toHaveBeenCalledTimes(1);
  });

  it("wires the Docker Compose option to onDockerCompose", () => {
    const onTemplates = vi.fn();
    const onDockerCompose = vi.fn();
    const opts = buildImportOptions({ onDockerCompose, onTemplates });

    const dockerCompose = opts.find((o) => o.id === "docker-compose");
    expect(dockerCompose).toBeDefined();

    dockerCompose!.onClick();
    expect(onDockerCompose).toHaveBeenCalledTimes(1);
  });
});
