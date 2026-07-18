// @vitest-environment jsdom
import { describe, it, expect, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useResourceTabProps, type ResourceTabContext } from "../use-resource-tab-props";
import type { UseSecretsReturn } from "@/pages/stacks/hooks/use-secrets";
import type { FormStackResourceData } from "@/pages/stacks/schemas/form-schema";

const ctx: ResourceTabContext = {
  errors: {},
  volumes: [],
  allResources: [
    { name: "web", index: 0, outputs: [] },
    { name: "api", index: 1, outputs: ["URL"] },
  ],
  secrets: { secrets: [], isLoading: false } as unknown as UseSecretsReturn,
  addons: [],
  addonNameById: new Map(),
};

describe("useResourceTabProps", () => {
  it("projects the resource into configuration props", () => {
    const resource: Partial<FormStackResourceData> = { name: "web", sourceType: "image", source: { image: { ref: "nginx" } } };
    const { result } = renderHook(() => useResourceTabProps({ resource, index: 0, onChange: vi.fn(), context: ctx }));
    expect(result.current.configurationProps.draft.name).toBe("web");
    expect(result.current.environmentProps.index).toBe(0);
  });

  it("excludes the current resource from env resourceOptions", () => {
    const resource: Partial<FormStackResourceData> = { name: "web" };
    const { result } = renderHook(() => useResourceTabProps({ resource, index: 0, onChange: vi.fn(), context: ctx }));
    expect(result.current.environmentProps.resourceOptions.map((o) => o.name)).toEqual(["api"]);
  });

  it("onPatchResource forwards a shallow-merged resource to onChange", () => {
    const onChange = vi.fn();
    const resource: Partial<FormStackResourceData> = { name: "web", source: { image: { ref: "nginx" } } };
    const { result } = renderHook(() => useResourceTabProps({ resource, index: 2, onChange, context: ctx }));
    act(() => result.current.configurationProps.onPatchResource({ source: { image: { ref: "redis" } } }));
    expect(onChange).toHaveBeenCalledWith(2, expect.objectContaining({ name: "web", source: { image: { ref: "redis" } } }));
  });

  it("reports dirty tabs against a baseline", () => {
    const resource: Partial<FormStackResourceData> = { name: "web", source: { image: { ref: "redis" } } };
    const baselineResource: Partial<FormStackResourceData> = { name: "web", source: { image: { ref: "nginx" } } };
    const { result } = renderHook(() =>
      useResourceTabProps({ resource, index: 0, baselineResource, onChange: vi.fn(), context: ctx }),
    );
    expect(result.current.isDirty).toBe(true);
    expect(result.current.dirtyTabs.configuration).toBe(true);
  });
});
