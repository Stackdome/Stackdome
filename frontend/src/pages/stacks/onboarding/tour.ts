import { driver, type Driver, type DriveStep } from "driver.js";
import "driver.js/dist/driver.css";
import "./tour.css";

const DONE_KEY = "stackdome.onboarding-tour.done";

/** Where the tour is across route changes. Module state is enough — a full
    page reload abandons the tour, which is the intended quiet failure mode. */
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
    // A guided tour: stray clicks on the dimmed page do nothing, and the
    // spotlit element is inert unless a step opts back in (click-me beats).
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

/** Beat 1 lives in WelcomeDialog (a real dialog, not a driver popover); on
    accept it calls this before seeding the draft so beat 2 can pick up. */
export function startCanvasStage(): void {
  stage = "canvas";
}

/** Clicks the drawer's nth tab (0 configuration, 1 deployment, 2 environment).
    Radix tabs activate on mousedown, so a bare click() does nothing. */
function clickDrawerTab(index: number): void {
  const tab = document.querySelectorAll<HTMLButtonElement>(
    '[data-testid="resource-drawer"] [role="tab"]',
  )[index];
  tab?.dispatchEvent(new MouseEvent("mousedown", { bubbles: true }));
  tab?.click();
}

/** Beats 2–8 on the draft canvas. The web-card step waits for the user to
    click the card; the opened drawer gets one beat per tab so each concept is
    explained where it lives. The last step leaves the Deploy pill interactive —
    clicking it deploys and the editor tab switch advances us. */
export function runCanvasTour(): void {
  if (stage !== "canvas") return;
  let drawerWatch: MutationObserver | null = null;
  const d = run(
    [
      {
        // No element: the canvas fills the viewport, so anchoring pushes the
        // popover off the top. Centred reads as the opening card it is.
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
        onHighlighted: () => {
          drawerWatch?.disconnect();
          drawerWatch = new MutationObserver(() => {
            if (document.querySelector('[data-testid="resource-drawer"]')) {
              drawerWatch?.disconnect();
              drawerWatch = null;
              setTimeout(() => d.moveNext(), 350);
            }
          });
          drawerWatch.observe(document.body, { childList: true, subtree: true });
        },
      },
      {
        element: '[data-testid="resource-drawer"]',
        popover: {
          title: "Configuration",
          description:
            "The basics live here: the name, where the image comes from, and the port it listens on. All filled in for the demo.",
          side: "left",
          align: "center",
          onNextClick: () => {
            clickDrawerTab(1);
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
            clickDrawerTab(2);
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
            "This one is built from a Git repo instead of an image. It has no port, so nothing outside can reach it.",
          side: "right",
        },
      },
      {
        element: '[data-testid="deploy-pill"]',
        popover: {
          title: "Time to deploy",
          description:
            "Click Deploy. Your cluster will fetch the code, build it, and start everything up.",
          side: "top",
          showButtons: ["close"],
        },
        disableActiveInteraction: false,
      },
    ],
    { onDestroyed: () => drawerWatch?.disconnect() },
  );
}

/** Beat 6 — the activity timeline right after Deploy. Dismissing hands off to
    the build; we come back when the release converges. The feed grows as
    events stream in, so the spotlight follows the element's size. */
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
            "Every step shows up here as it happens. The build takes a few minutes. When the app is live, we will point you to it.",
          side: "left",
          doneBtnText: "Got it",
        },
        onHighlighted: (el) => {
          if (!el) return;
          grow = new ResizeObserver(() => d.refresh());
          grow.observe(el);
        },
      },
    ],
    { onDestroyed: () => grow?.disconnect() },
  );
}

/** Both header variants render an endpoint row: a compact chip and a wide row
    that spells out the address. Spotlight whichever shows the address, i.e.
    the widest one currently on screen. */
function visibleEndpointRow(): Element {
  const rows = [...document.querySelectorAll('[data-tour="public-endpoints"]')].filter(
    (el) => (el as HTMLElement).offsetParent !== null,
  );
  return rows.reduce((widest, el) =>
    el.getBoundingClientRect().width > widest.getBoundingClientRect().width ? el : widest,
  );
}

/** Beat 8 — the public URL chip once the release has converged. */
export function runLiveStep(): void {
  stage = "done";
  run(
    [
      {
        element: visibleEndpointRow,
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
