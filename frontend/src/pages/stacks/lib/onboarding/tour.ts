import { driver, type Driver, type DriveStep } from "driver.js";
import "driver.js/dist/driver.css";
import "./tour.css";

const DONE_KEY = "stackdome.onboarding-tour.done";

/** Module state is enough — a full page reload abandons the tour, which is the
    intended quiet failure mode. */
type Stage = "idle" | "canvas" | "deploying" | "done";
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

function run(steps: DriveStep[], opts?: { onDestroyed?: () => void }): Driver {
  active?.destroy();
  const d = driver({
    popoverClass: "stackdome-tour",
    showProgress: steps.length > 1,
    showButtons: ["next", "close"],
    allowClose: true,
    overlayClickBehavior: () => {},
    disableActiveInteraction: true,
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

/** Must run before the draft is seeded, so the canvas beats pick up on arrival. */
export function startCanvasStage(): void {
  stage = "canvas";
}

/** Drawer tabs in render order; the tour drives them by position. */
const DRAWER_TABS = ["configuration", "deployment", "environment"] as const;

/** Radix tabs activate on mousedown, so a bare click() does nothing. */
function showDrawerTab(tab: (typeof DRAWER_TABS)[number]): void {
  const tabs = document.querySelectorAll<HTMLButtonElement>(
    '[data-testid="resource-drawer"] [role="tab"]',
  );
  tabs[DRAWER_TABS.indexOf(tab)]?.dispatchEvent(new MouseEvent("mousedown", { bubbles: true }));
}

/** The Deploy step stays interactive: clicking it deploys, and the editor's own
    switch to the deployments tab is what advances the tour. */
export function runCanvasTour(): void {
  if (stage !== "canvas") return;
  const d = run(
    [
      {
        // The canvas fills the viewport, so anchoring pushes the popover off the top.
        popover: {
          title: "This is the Stack Canvas",
          description:
            "Your whole app lives here as one stack. Each card is a resource. The lines show who talks to whom.",
        },
      },
      {
        element: '.react-flow__node[data-id="resource:web"]',
        popover: {
          title: "The web resource",
          description:
            "This one runs a ready-made container image. Click the card to look inside.",
          side: "bottom",
          showButtons: ["close"],
        },
        disableActiveInteraction: false,
        advanceOnClick: true,
      },
      {
        element: '[data-testid="resource-drawer"]',
        waitForElement: 2000,
        popover: {
          title: "Configuration",
          description:
            "The basics live here: the name, where the image comes from, and the port it listens on. All filled in for the demo.",
          side: "left",
          align: "center",
          onNextClick: () => {
            showDrawerTab("deployment");
            setTimeout(() => d.moveNext(), 250);
          },
        },
      },
      {
        element: '[data-testid="resource-drawer"]',
        popover: {
          title: "Deployment",
          description:
            "This controls how the container starts. The demo sets its start command here. Leave it empty and the image's own default is used.",
          side: "left",
          align: "center",
          onNextClick: () => {
            showDrawerTab("environment");
            setTimeout(() => d.moveNext(), 250);
          },
        },
      },
      {
        element: '[data-testid="resource-drawer"]',
        popover: {
          title: "Environment",
          description:
            "The variables the app reads at runtime. A value can be typed in, or point at another resource: REDIS_URL picks up redis's address, and PUBLIC_URL points back at this resource. Both fill in at deploy time, so no addresses are hardcoded.",
          side: "left",
          align: "center",
          onNextClick: () => {
            document
              .querySelector<HTMLButtonElement>('[data-testid="resource-drawer"] [aria-label="Close"]')
              ?.click();
            setTimeout(() => d.moveNext(), 300);
          },
        },
      },
      {
        element: '.react-flow__node[data-id="resource:worker"]',
        popover: {
          title: "The worker stays private",
          description:
            "It has no port, so nothing outside can reach it. Only web can, over the private network.",
          side: "right",
        },
      },
      {
        element: '[data-testid="deploy-pill"]',
        popover: {
          title: "Time to deploy",
          description:
            "Click Deploy. Your cluster will pull the images and start everything up.",
          side: "top",
          showButtons: ["close"],
        },
        disableActiveInteraction: false,
      },
    ],
    // Advancing to the timeline moves the stage first; anything else here is
    // the user closing the tour, which retires it.
    { onDestroyed: () => { if (stage === "canvas") markTourDone(); } },
  );
}

export function runTimelineStep(): void {
  stage = "deploying";
  let grow: ResizeObserver | null = null;
  const d = run(
    [
      {
        element: '[data-tour="deploy-timeline"]',
        popover: {
          title: "Watch it get deployed",
          description:
            "Every step shows up here as it happens. This takes a minute or two. When the app is live, we will point you to it.",
          side: "left",
          doneBtnText: "Got it",
        },
        // The feed grows as events stream in; the spotlight has to follow it.
        onHighlighted: (el) => {
          grow = new ResizeObserver(() => d.refresh());
          grow.observe(el!);
        },
      },
    ],
    { onDestroyed: () => grow?.disconnect() },
  );
}

/** Both header variants render an endpoint row: a compact chip and a wide row
    that spells out the address. */
function widestEndpointRow(): Element | undefined {
  return [...document.querySelectorAll('[data-tour="public-endpoints"]')]
    .filter((el) => (el as HTMLElement).offsetParent !== null)
    .sort((a, b) => b.getBoundingClientRect().width - a.getBoundingClientRect().width)[0];
}

export function runLiveStep(): void {
  stage = "done";
  run(
    [
      {
        element: widestEndpointRow(),
        popover: {
          title: "Your app is live",
          description: "Here is its public address. Open it and say hello.",
          side: "bottom",
          doneBtnText: "Finish",
        },
        disableActiveInteraction: false,
      },
    ],
    { onDestroyed: markTourDone },
  );
}
