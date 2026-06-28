// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import TemplatesBrowserDialog from "../templates-browser-dialog";
import type { Template } from "@/data/templates/types";

const tooljet: Template = {
  id: "tooljet",
  name: "ToolJet",
  initials: "Tj",
  icon: "/tj.svg",
  category: "Dev Tools",
  shortDescription: "Low-code internal tools.",
  longDescription: "Build internal tools without a frontend team.",
  website: "https://tooljet.com",
  docs: "https://docs.tooljet.com",
  version: "ee-lts-latest",
  stackYaml: "name: tooljet",
};

const ghost: Template = {
  id: "ghost",
  name: "Ghost",
  initials: "Gh",
  icon: "/gh.svg",
  category: "Publishing",
  shortDescription: "Publishing platform.",
  longDescription: "Blogs, newsletters and memberships.",
  website: "https://ghost.org",
  docs: "https://ghost.org/docs",
  version: "5",
  stackYaml: "name: ghost",
};

// Radix focus management occasionally calls scrollIntoView, which jsdom lacks.
beforeAll(() => {
  Element.prototype.scrollIntoView = vi.fn();
});

afterEach(cleanup);

function setup() {
  const onUse = vi.fn();
  const onOpenChange = vi.fn();
  render(
    <TemplatesBrowserDialog
      open
      onOpenChange={onOpenChange}
      templates={[tooljet, ghost]}
      onUse={onUse}
    />,
  );
  return { onUse, onOpenChange };
}

describe("TemplatesBrowserDialog", () => {
  it("lists templates and shows the first template's detail by default", () => {
    setup();
    expect(screen.getByRole("option", { name: /ToolJet/ })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: /Ghost/ })).toBeInTheDocument();
    // detail pane reflects the default (first) selection
    expect(
      screen.getByText("Build internal tools without a frontend team."),
    ).toBeInTheDocument();
    expect(screen.getByText(/ee-lts-latest/)).toBeInTheDocument();
    expect(screen.getByText(/2 templates/i)).toBeInTheDocument();
  });

  it("filters the list by search query", async () => {
    const user = userEvent.setup();
    setup();
    await user.type(
      screen.getByPlaceholderText(/search templates/i),
      "ghost",
    );
    expect(
      screen.queryByRole("option", { name: /ToolJet/ }),
    ).not.toBeInTheDocument();
    expect(screen.getByRole("option", { name: /Ghost/ })).toBeInTheDocument();
  });

  it("calls onUse with the clicked template", async () => {
    const user = userEvent.setup();
    const { onUse } = setup();
    await user.click(screen.getByRole("option", { name: /Ghost/ }));
    await user.click(screen.getByRole("button", { name: /use template/i }));
    expect(onUse).toHaveBeenCalledWith(expect.objectContaining({ id: "ghost" }));
  });

  it("moves selection with arrow keys and uses on Enter", async () => {
    const user = userEvent.setup();
    const { onUse } = setup();
    await user.click(screen.getByPlaceholderText(/search templates/i));
    await user.keyboard("{ArrowDown}{Enter}");
    expect(onUse).toHaveBeenCalledWith(expect.objectContaining({ id: "ghost" }));
  });

  it("shows an empty state when nothing matches", async () => {
    const user = userEvent.setup();
    setup();
    await user.type(
      screen.getByPlaceholderText(/search templates/i),
      "zzzzzz",
    );
    expect(screen.getByText(/no templates match/i)).toBeInTheDocument();
  });
});
