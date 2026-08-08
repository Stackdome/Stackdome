import * as React from "react"
import { Link } from "react-router-dom"
import { StackdomeWordmark } from "@/components/branded"

import { navGroups } from "@/components/nav-items"
import { NavItem } from "@/components/nav-item"
import { NavUser } from "@/components/nav-user"
import { ThemeToggle } from "@/components/theme-toggle"
import { getCurrentUser } from "@/lib/common"
import { useCurrentUser } from "@/hooks/use-current-user"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
  SidebarMenu,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarGroupContent,
} from "@/components/ui/sidebar"

/**
 * The frame's navigation (§12), built from `navGroups` rather than one component
 * per destination.
 *
 * Measurements come from the `app shell` Figma board: 240px expanded, 56px
 * collapsed, a 64px brand band with the lockup inset 16px, and a 12px gutter
 * around a 32px item pitched every 34px.
 *
 * The first group carries no label — Stacks and Previews are the product, and a
 * label above the first item is furniture. "Platform" holds what you attach to a
 * stack; "Infrastructure" holds what an admin configures once, which is already
 * the admin-gated set.
 */
export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const user = getCurrentUser();
  const { isOrgAdmin } = useCurrentUser();

  const userData = {
    name: user?.name || user?.username || "User",
    email: user?.email || "user@example.com",
    avatar: "",
    organisation: user?.organisation,
  };

  return (
    <Sidebar collapsible="icon" {...props}>
      {/* Brand band — product identity only, one line. The organisation is
          account context, so it lives in the account block at the foot, which
          is the only place it can be switched.

          The lockup shares a CENTRELINE with the page title across the seam.
          Both columns carry the shell's 12px gutter and then 16px of their own
          padding, so both 32px rows centre on 44. The sidebar is `fixed` and
          ignores the wrapper's padding, so its gutter is paid here instead:
          28px of lead-in (12 gutter + 16 padding), the 32px row, then 16px —
          76px total, which is where the board puts the first nav item.

          If the header padding changes, this number changes with it. Verified
          in the browser, not eyeballed. */}
      {/* The 16px left inset never changes. The board anchors the lockup to the
          left in BOTH states — collapsing narrows the rail around it, it does
          not re-centre the mark. */}
      <SidebarHeader className="h-[76px] pb-4 pl-4 pr-3 pt-[28px]">
        {/* The lockup is named for the PRODUCT, not its destination. Labelling
            it "…go to Stacks" gave this link and the Stacks nav row the same
            accessible name, so a screen reader announced two different
            controls identically. */}
        <Link
          to="/stacks"
          aria-label="Stackdome"
          className="focus-ring-edge flex h-8 items-center rounded-md"
        >
          {/* ONE drawing, clipped — never two SVGs swapped. The lockup is a cube
              followed by the wordmark, so closing the frame from the right stops
              at the cube and the cube never moves. `StackdomeWordmark` renders
              2× its `size` because the viewBox carries glow padding, so 13.5 is
              the board's 27px box. */}
          <span className="rail-logo block overflow-hidden">
            <StackdomeWordmark size={13.5} className="block max-w-none" />
          </span>
        </Link>
      </SidebarHeader>

      <SidebarContent className="gap-0.5 px-3">
        {navGroups.map((group, i) => {
          const items = group.items.filter((item) => !item.adminOnly || isOrgAdmin);
          if (items.length === 0) return null;

          return (
            <SidebarGroup key={group.label ?? `group-${i}`} className="gap-0 p-0">
              {group.label && (
                <>
                  {/* Not a centred 32px block: the board hangs the label off
                      the group below it — 12px of air above, 4px below. */}
                  <SidebarGroupLabel className="rail-y h-auto overflow-hidden pb-1 pl-2 pr-0 pt-3 text-label font-medium leading-4 text-fg-muted">
                    {group.label}
                  </SidebarGroupLabel>
                  {/* Collapsed, the label has no room — but the grouping still
                      has to survive, so the board replaces each one with a
                      16px centred hairline. Without it the rail is nine
                      undifferentiated glyphs.
                      Decorative: the groups are already named for assistive
                      tech by the expanded label, so this is aria-hidden. */}
                  <div aria-hidden className="rail-y-in mx-auto h-px w-4 bg-border" />
                </>
              )}
              <SidebarGroupContent>
                {/* 34px pitch — a 32px row and a 2px gap. */}
                <SidebarMenu className="gap-0.5">
                  {items.map((item) => (
                    <NavItem key={item.path} item={item} />
                  ))}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          );
        })}
      </SidebarContent>

      {/* Global helpers live in the frame, never on the content sheet — they are
          about the product, not about the page you are looking at. The account
          block is inset 7px rather than 12px: its 28px avatar carries a border,
          so squaring its optical left edge with the 16px nav glyphs above needs
          the extra 5px. */}
      <SidebarFooter className="gap-0.5 p-0 pb-3">
        <div className="px-3">
          <ThemeToggle presentation="row" />
        </div>
        {/* The account is inset 7px rather than 12px: its 28px avatar carries a
            hairline, so squaring its optical left edge with the 16px glyphs
            above needs the extra 5px.

            That inset is the SAME in both states — the avatar does not move
            when the rail collapses, it is simply the last thing left. Only the
            right inset closes, to the board's 44px collapsed width. */}
        <div className="pl-[7px] pr-3 group-data-[collapsible=icon]:pr-[5px]">
          <NavUser user={userData} />
        </div>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
