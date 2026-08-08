import * as React from "react";
import { Link, useLocation, useNavigate, useNavigationType } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";

interface BreadcrumbItemType {
  name: string;
  path: string;
  clickable: boolean;
}

/**
 * The top of the sheet (§12a) — **two parts, one component.**
 *
 *   1. The **title row** identifies the section: the collapse toggle, the page
 *      title, the page's one fact and its actions.
 *   2. The **toolbar row** holds the tools for that section — search, filters,
 *      sort, a view toggle. Whatever the section needs.
 *
 * The toolbar row is **conditional**, and it collapses itself: pages portal
 * into `#sheet-toolbar`, and `empty:hidden` removes the row when nothing
 * arrives. A `display:none` child takes no part in the flex gap either, so the
 * band falls from 108px to 64px with no height math anywhere.
 *
 *   16 + 32 + 16           = 64   title row only
 *   16 + 32 + 12 + 32 + 16 = 108  with a toolbar
 *
 * **The padding is 16 and the gap between rows is 12.** They are deliberately
 * not the same number: the outer inset is the sheet's own margin, the inner one
 * is the distance between two rows of controls.
 *
 * Measurements come from the `app shell` Figma board. The band carries a **1px
 * bottom hairline**, full sheet width, drawn INSIDE the 108 — content is cut by
 * it as it scrolls under, rather than dissolving into it.
 *
 * ### Journeys get an exit; nested pages get a trail
 *
 * A page that calls `useJourney(origin)` is declaring itself a **task launched
 * from a main screen** — `New stack` — rather than a place you navigated down
 * into. The header then renders a back arrow and **the title alone.**
 *
 * The trail is dropped on purpose. The sidebar is always on screen with the
 * section highlighted, so `Stacks /` would be the *third* way back off one
 * screen, and it pushes the page title into third position. The two marks were
 * never doing the same job: a crumb says *where you are*, an arrow says *this
 * is a task you can leave* — and only the second was unmet.
 *
 * Back steps through **history**, so it returns you the way you came. Deep
 * linked, with nothing to step back to, it falls back to the journey's origin.
 */
export function SheetHeader({ leading }: { leading?: React.ReactNode }) {
  const location = useLocation();
  const navigate = useNavigate();
  const navigationType = useNavigationType();
  const { customLabels, loadingLabels, nonClickablePaths, journeyOrigin } = useBreadcrumb();

  const pathSegments = location.pathname.split("/").filter(Boolean);

  // There is no "Home" crumb. Every top-level destination is one click away in
  // the sidebar, so a Home hop says nothing the frame isn't already saying —
  // and it pushed the real page title out of first position. The trail's LAST
  // segment is the page title (§12a).
  const breadcrumbItems: BreadcrumbItemType[] = pathSegments.map(
    (segment, index): BreadcrumbItemType => {
      const path = "/" + pathSegments.slice(0, index + 1).join("/");
      const clickable = !nonClickablePaths[path];
      if (index === pathSegments.length - 1 && loadingLabels?.[path]) {
        return { name: "...", path, clickable };
      }
      if (customLabels[path]) {
        return { name: customLabels[path], path, clickable };
      }
      const name = segment.charAt(0).toUpperCase() + segment.slice(1);
      return { name, path, clickable };
    },
  );

  // A journey shows its title alone (§12a). The trail's last segment IS the
  // title, so keep that and drop the wayfinding in front of it: the sidebar
  // already says which section you are in, and the arrow is the way out.
  const trailItems = journeyOrigin ? breadcrumbItems.slice(-1) : breadcrumbItems;

  function goBack() {
    // Did we arrive here by navigating inside the app? Only then is there an
    // entry to step back to. A deep link, a fresh tab or a reload all arrive as
    // POP, and stepping back from those leaves the product entirely — so those
    // fall back to wherever the journey is launched from.
    if (navigationType === "PUSH") navigate(-1);
    else navigate(journeyOrigin ?? "/");
  }

  return (
    <div className="sheet-edge-b flex flex-col gap-3 bg-card px-4 py-4">
      {/* ── 1. Title row — what section am I in, and what can I do to it? ── */}
      <div className="flex h-8 items-center gap-1.5">
        {leading}
        {journeyOrigin && (
          <>
            {/* The divider separates chrome belonging to the SHELL (collapse)
                from chrome belonging to THIS JOURNEY (back). */}
            <div aria-hidden className="bg-border h-5 w-px flex-none" />
            <Button
              variant="ghost"
              size="icon"
              className="size-8 text-fg-2"
              onClick={goBack}
              aria-label="Back"
            >
              <ArrowLeft />
            </Button>
          </>
        )}
        <Breadcrumb>
          <BreadcrumbList className="text-body gap-2 sm:gap-2">
            {trailItems.map((item, index) => (
              <React.Fragment key={index}>
                {index > 0 && (
                  <BreadcrumbSeparator className="text-fg-2/50 [&>svg]:hidden">
                    <span>/</span>
                  </BreadcrumbSeparator>
                )}
                {index === trailItems.length - 1 ? (
                  <BreadcrumbItem>
                    {/* The page title. The trail before it is 13px wayfinding.
                        14/20 at weight 500, per the Shape + Hierarchy board
                        (node 110:4030). It has been 16/24 at 600 and 20/28 at
                        500; both made the header shout over the content it
                        introduces. The title is a label on the sheet, not a
                        headline — one rung above the trail, one below a card's
                        own name, and hierarchy is carried by the card. */}
                    <BreadcrumbPage className="text-name font-medium text-foreground">
                      {item.name}
                    </BreadcrumbPage>
                  </BreadcrumbItem>
                ) : !item.clickable ? (
                  <BreadcrumbItem>
                    <span className="text-fg-2">{item.name}</span>
                  </BreadcrumbItem>
                ) : (
                  <BreadcrumbItem>
                    <BreadcrumbLink asChild className="text-fg-2 hover:text-foreground transition-colors">
                      <Link to={item.path}>{item.name}</Link>
                    </BreadcrumbLink>
                  </BreadcrumbItem>
                )}
              </React.Fragment>
            ))}
          </BreadcrumbList>
        </Breadcrumb>

        {/* The page's one fact and its actions. Only things scoped to the page
            you are looking at may land here; global helpers stay in the grey
            frame. Bar buttons run at `text-name`, enforced here so call sites
            cannot drift. 8px between actions; the fact adds 4px more to reach
            the 12px the board specifies. */}
        <div id="topnav-actions" className="ml-auto flex items-center gap-2 [&_button]:text-name" />
      </div>

      {/* ── 2. Toolbar row — the tools for that section. Conditional. ── */}
      <div id="sheet-toolbar" className="flex h-8 items-center gap-1.5 empty:hidden" />
    </div>
  );
}
