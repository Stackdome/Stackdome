import { Boxes, FileCode, GitBranch, Package, SquareDashed } from "lucide-react";
import { createElement } from "react";

import { blockCatalog } from "@/pages/stacks/data/blocks/registry";
import { templates } from "@/pages/stacks/data/templates/registry";

import type { StartingPoint } from "./starting-point-tabs";

/**
 * The five peers. **They are peers** — none is a default dressed as a
 * recommendation, and the order runs from "your own code" outwards to "nothing
 * at all", which is also roughly how often each is used.
 */
export type Source = "git" | "template" | "compose" | "blocks" | "blank";

/**
 * "n8n, Grafana, Immich and 4 more" — **named, then counted.**
 *
 * The names are chosen, not sliced off the front of the registry: the line's
 * job is to be recognised, and registry order is insertion order. The count is
 * derived from the real total, so the sentence cannot go stale as the registry
 * grows — only the three names are a copy decision.
 */
function namedThenCounted(shown: string[], total: number) {
  const rest = total - shown.length;
  const head = shown.join(", ");
  return rest > 0 ? `${head} and ${rest} more` : head;
}

/**
 * **Every title is a verb.** These were nouns — "A repository", "A compose
 * file" — with a `Start from` label above them doing the grammar. Three of the
 * five began with the article "A", which existed only to finish that label's
 * sentence, and the label sat 966px from the last tab it governed.
 *
 * The description now carries the **specifics** rather than a second
 * explanation: which provider, which apps, which parts.
 */
export const STARTING_POINTS: StartingPoint<Source>[] = [
  {
    value: "git",
    name: "Deploy your own code",
    description: "A connected provider, or a URL",
    icon: createElement(GitBranch),
  },
  {
    value: "template",
    name: "Run a ready-made app",
    description: namedThenCounted(["n8n", "Grafana", "Immich"], templates.length),
    icon: createElement(Package),
  },
  {
    value: "compose",
    name: "Import a compose file",
    description: "Paste or drop the YAML",
    icon: createElement(FileCode),
  },
  {
    value: "blocks",
    name: "Assemble from blocks",
    description: namedThenCounted(
      ["Postgres", "Redis"],
      // Data stores only — the two generic service shapes are not what anyone
      // comes to this tab for, and counting them overstates the catalogue.
      blockCatalog.filter((b) => b.category !== "services").length,
    ),
    icon: createElement(Boxes),
  },
  {
    value: "blank",
    name: "Start from nothing",
    description: "An empty canvas you fill",
    icon: createElement(SquareDashed),
    // The odd one out, and it says so: a dashed chip reports "nothing in it
    // yet", which is a fact rather than decoration.
    dashed: true,
  },
];

/**
 * The sentence under the strip. It says what the chosen starting point actually
 * does, because the tab's own line cannot carry it — and because §1's user
 * should not need documentation open beside the product.
 */
export const SOURCE_LEDE: Record<Source, string> = {
  git: "Stackdome builds the image from a branch and gives the service a URL.",
  template: "A known app, already wired. You can change anything afterwards.",
  compose: "Paste or drop a docker-compose file. We read the services out of it.",
  blocks: "Pick the parts. Known software lands already configured.",
  blank: "An empty canvas. You add every piece yourself.",
};
