// @vitest-environment jsdom
import { beforeEach, describe, expect, it, vi } from "vitest";
import { openWebResource } from "@/pages/stacks/lib/onboarding/tour";

describe("openWebResource", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
  });

  function renderWebNode() {
    const node = document.createElement("div");
    node.className = "react-flow__node";
    node.dataset.id = "resource:web";
    const onClick = vi.fn();
    node.addEventListener("click", onClick);
    document.body.append(node);
    return onClick;
  }

  it("clicks the web card when the drawer is closed", () => {
    const onClick = renderWebNode();
    openWebResource();
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it("does nothing when the drawer is already open", () => {
    const onClick = renderWebNode();
    const drawer = document.createElement("div");
    drawer.dataset.testid = "resource-drawer";
    document.body.append(drawer);

    openWebResource();

    expect(onClick).not.toHaveBeenCalled();
  });
});
