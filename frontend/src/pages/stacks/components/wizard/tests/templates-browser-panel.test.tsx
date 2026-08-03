// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeAll, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TemplatesBrowserPanel } from "../templates-browser-panel";
import type { Template } from "@/pages/stacks/data/templates/types";

const tooljet: Template = {
  id: "tooljet",
  name: "ToolJet",
  initials: "TJ",
  icon: "box",
  category: "Website",
  shortDescription: "Low-code",
  longDescription: "Low-code platform",
  website: "https://tooljet.com",
  docs: "https://docs.tooljet.com",
  version: "1.0",
  stackYaml: "services:\n  tooljet:\n    image: tooljet/tooljet:latest\n",
};

// Radix focus management occasionally calls scrollIntoView, which jsdom lacks.
beforeAll(() => {
  Element.prototype.scrollIntoView = vi.fn();
});

afterEach(cleanup);

describe("TemplatesBrowserPanel", () => {
  it("calls onUse with the selected template", async () => {
    const user = userEvent.setup();
    const onUse = vi.fn();
    render(<TemplatesBrowserPanel templates={[tooljet]} onBack={vi.fn()} onUse={onUse} />);
    await user.click(screen.getByRole("button", { name: /Continue/i }));
    expect(onUse).toHaveBeenCalledWith(tooljet);
  });

  it("shows the template count marker", () => {
    render(<TemplatesBrowserPanel templates={[tooljet]} onBack={vi.fn()} onUse={vi.fn()} />);
    expect(screen.getByText("1 TEMPLATES")).toBeInTheDocument();
  });
});
