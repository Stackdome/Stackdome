import { driver, type Driver, type DriveStep } from "driver.js";
import "driver.js/dist/driver.css";
import "./tour.css";

const DONE_KEY = "stackdome.onboarding-tour.done";

/** Where the tour is across route changes. Module state is enough — a full
    page reload abandons the tour, which is the intended quiet failure mode. */
type Stage = "idle" | "canvas" | "deploying" | "live" | "done";
let stage: Stage = "idle";
let active: Driver | null = null;

export function isTourDone(): boolean {
  return localStorage.getItem(DONE_KEY) === "1";
}

export function markTourDone(): void {
  localStorage.setItem(DONE_KEY, "1");
  stage = "done";
}

export function tourStage(): Stage {
  return stage;
}

export function endTour(): void {
  active?.destroy();
  active = null;
  if (stage !== "done") markTourDone();
}

function run(steps: DriveStep[], opts?: { onDestroyed?: () => void }): Driver {
  active?.destroy();
  const d = driver({
    popoverClass: "stackdome-tour",
    showProgress: steps.length > 1,
    allowClose: true,
    overlayOpacity: 0.55,
    stagePadding: 6,
    stageRadius: 10,
    steps,
    onDestroyed: () => {
      if (active === d) active = null;
      opts?.onDestroyed?.();
    },
  });
  active = d;
  d.drive();
  return d;
}

/** Beat 1 — element-less welcome popover on the stacks list. Accepting seeds
    the demo draft; closing marks the tour done and never shows it again. */
export function startWelcome(onAccept: () => void): void {
  let accepted = false;
  run(
    [
      {
        popover: {
          title: "Deploy your first stack",
          description:
            "We prepared a small demo app — a web service, a queue, and a background worker. Take a two-minute tour and watch it go live.",
          nextBtnText: "Show me",
          onNextClick: () => {
            accepted = true;
            stage = "canvas";
            active?.destroy();
            onAccept();
          },
        },
      },
    ],
    { onDestroyed: () => { if (!accepted) markTourDone(); } },
  );
}

/** Beats 2–5 on the draft canvas. The last step leaves the Deploy pill
    interactive — clicking it deploys and the editor tab switch advances us. */
export function runCanvasTour(): void {
  if (stage !== "canvas") return;
  run([
    {
      element: '[data-testid="stack-canvas"]',
      popover: {
        title: "This is your stack",
        description:
          "Each card is a container. Lines show who talks to whom: web pushes jobs to redis, the worker picks them up.",
        side: "top",
      },
    },
    {
      element: '.react-flow__node[data-id="resource:web"]',
      popover: {
        title: "The web service",
        description:
          "Built straight from a Git repo — no image to push. Click any card later to see its configuration. Everything here is already set up.",
        side: "right",
      },
    },
    {
      element: '.react-flow__node[data-id="resource:worker"]',
      popover: {
        title: "The worker has no URL",
        description:
          "No port, no ingress — on purpose. Only web is reachable from outside; the worker just watches the queue.",
        side: "right",
      },
    },
    {
      element: '[data-testid="deploy-pill"]',
      popover: {
        title: "Ship it",
        description:
          "Click Deploy. Your cluster will clone the repo, build both images, and roll everything out.",
        side: "top",
        showButtons: ["close"],
      },
      disableActiveInteraction: false,
    },
  ]);
}

/** Beat 6 — the activity timeline right after Deploy. Dismissing hands off to
    the build; we come back when the release converges. */
export function runTimelineStep(): void {
  stage = "deploying";
  run([
    {
      element: '[data-tour="deploy-timeline"]',
      popover: {
        title: "Watch it build",
        description:
          "This feed is live: cloning, image builds, rollout. Builds take a few minutes — when the stack is live we'll show you where to click.",
        side: "left",
        doneBtnText: "I'll watch",
      },
    },
  ]);
}

/** Both header variants (collapsed/expanded) render an endpoint row; only one
    is visible. Spotlight the visible one. */
function visibleEndpointRow(): Element {
  const rows = [...document.querySelectorAll('[data-tour="public-endpoints"]')];
  return rows.find((el) => (el as HTMLElement).offsetParent !== null) ?? rows[0];
}

/** Beat 8 — the public URL chip once the release has converged. */
export function runLiveStep(): void {
  stage = "live";
  run(
    [
      {
        element: visibleEndpointRow,
        popover: {
          title: "Your app is live",
          description: "This is its public URL. Open it — it has been waiting to celebrate with you.",
          side: "bottom",
          doneBtnText: "Finish",
        },
        disableActiveInteraction: false,
      },
    ],
    { onDestroyed: markTourDone },
  );
}
