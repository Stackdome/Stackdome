import * as React from "react";
import { SidebarProvider, SidebarTrigger, SidebarInset } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import { Outlet, useLocation } from "react-router-dom";
import { BreadcrumbProvider } from "@/contexts/breadcrumb-context";
import { SheetHeader } from "@/components/sheet-header";
import { useGithubSetupLanding } from "@/hooks/use-github-setup-landing";

function AppLayoutContent({
  children,
  defaultSidebarOpen = true,
}: {
  children?: React.ReactNode;
  /** Start with the sidebar collapsed to its 56px rail. Stories use this to
   *  show the collapsed shell; the user's own toggle takes over from there. */
  defaultSidebarOpen?: boolean;
}) {
  useGithubSetupLanding();
  const location = useLocation();

  // Full-bleed: the canvas editor at /stacks/draft and /stacks/<id>, and the
  // New stack journey at /stacks/new — its starting-point strip runs the full
  // width of the sheet and it pins its own footer to the bottom, so neither can
  // sit inside the standard 16px page padding. A single trailing segment only,
  // so /stacks itself is unaffected.
  const isFullBleed = /^\/stacks\/[^/]+$/.test(location.pathname);

  return (
    <SidebarProvider defaultOpen={defaultSidebarOpen}>
      {/* 12px gutter on every free edge (§12). The sidebar sits flush to the
          window's left edge; the sheet is inset from the other three. */}
      <div className="flex h-screen max-h-screen w-full overflow-hidden bg-background py-3 pr-3">
        <AppSidebar />
        {/* The content plane is a white sheet floating on the paper frame —
            white floats, grey recedes. The sidebar needs no divider: the
            sheet's own edge draws the seam. */}
        {/* The sheet's hairline is an OUTLINE, not a border — the board draws
            it as an outside stroke, which is not part of the frame's 1186×876.
            A `border` would be, and it pushed the header's row down by 1px,
            which is exactly what put the two planes' centrelines out of step.
            `outline` paints outside the box, follows the radius, and costs the
            layout nothing.

            `ml-0.5` is the board's 2px gap between the rail and the sheet, and
            it is load-bearing rather than decorative: an outline is painted
            OUTSIDE the box, so with the two columns flush the left edge of it
            landed under the `fixed` sidebar and was clipped away. The gap is
            what lets the card be a card on all four sides.

            The line is `border-subtle` (6%), not the 11% hairline — the shadow
            now does the separating, so the edge only has to describe the shape. */}
        <SidebarInset className="ml-0.5 min-h-0 overflow-hidden rounded-lg bg-card shadow-md outline-1 outline-border-subtle">
          {/* The scroll container is a flex column so the header can be a
              sticky block of ANY height and nothing downstream needs to know
              what that height is. The old layout hardcoded `top-[52px]` in
              three places and `calc(100% - 52px)` in a fourth; the header is
              now 64px or 108px depending on whether the page has a toolbar, so
              every one of those numbers was about to become wrong. */}
          <div className="flex min-h-0 flex-grow flex-col overflow-auto scrollbar-hide">
            {/* Header, the page's sticky bar and the fade travel together as
                one sticky block pinned to the top of the sheet. */}
            <div className="sticky top-0 z-40 shrink-0">
              {/* Chrome, not content: 32px hit area, 16px glyph, fg-2. */}
              <SheetHeader leading={<SidebarTrigger className="size-8 text-fg-2" />} />
              {/* Forms pin a save bar directly beneath the header. */}
              <div id="page-sticky-bar" />
              {/* No fade. The band now carries a 1px hairline, and a dissolve
                  under a crisp line is two answers to the same question — the
                  gradient only blurred the 8px directly beneath the rule and
                  weakened it. Content is cut by the line instead. */}
            </div>

            {isFullBleed ? (
              <div className="min-h-0 flex-1">{children ? children : <Outlet />}</div>
            ) : (
              /* The sheet's content edge is 16px — the SAME edge the header
                 uses (§12a). It ran at 32px, so the page title sat on one edge
                 and everything under it on another, 20px in, with nothing on
                 screen explaining why.

                 `max-w-6xl` is gone with it. It capped the body at 1152 while
                 the header spanned the full sheet, so above 1280 the two
                 planes drifted apart — a second alignment bug waiting for a
                 wider monitor. A row's BOX lands on the edge and its text sits
                 8px inside, so the hover wash extends past the name. */
              <div className="px-4 py-4">{children ? children : <Outlet />}</div>
            )}
          </div>
        </SidebarInset>
      </div>
    </SidebarProvider>
  );
}

export function AppLayout({
  children,
  defaultSidebarOpen,
}: {
  children?: React.ReactNode;
  defaultSidebarOpen?: boolean;
}) {
  return (
    <BreadcrumbProvider>
      <AppLayoutContent defaultSidebarOpen={defaultSidebarOpen}>{children}</AppLayoutContent>
    </BreadcrumbProvider>
  );
}
