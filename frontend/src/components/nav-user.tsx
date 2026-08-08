"use client"

import {
  ChevronsUpDown,
  LogOut,
} from "lucide-react"

import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@/components/ui/avatar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar"
import { logoutAndRedirect } from "@/lib/common"

/** First letter of the first two words, e.g. "Ada Lovelace" → "AL". Falls back
 *  to the first two characters for single-word names, and to "?" for an empty
 *  one — this used to render a hardcoded "CN" for every user. */
function initials(name: string): string {
  const words = name.trim().split(/\s+/).filter(Boolean)
  if (words.length === 0) return "?"
  if (words.length === 1) return words[0].slice(0, 2).toUpperCase()
  return (words[0][0] + words[words.length - 1][0]).toUpperCase()
}

export function NavUser({
  user,
}: {
  user: {
    name: string
    email: string
    avatar: string
    /** Shown under the name. The org is the switchable context, and this menu
     *  is where switching lives — an email is not switchable, so it does not
     *  earn a permanent line. */
    organisation?: string
  }
}) {
  const { isMobile } = useSidebar()

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            {/* 48px, so the two lines and the 28px avatar sit without crowding
                — the one row in the frame that is not on the 32px ladder,
                because it carries two lines of type rather than one. */}
            <SidebarMenuButton
              size="lg"
              aria-label={`${user.name}${user.organisation ? `, ${user.organisation}` : ""} — account menu`}
              // The avatar must not move when the rail collapses: same height,
              // same left inset, same padding. Only the name and the chevron
              // vacate.
              //
              // The primitive forces every collapsed row to a 32px square with
              // no padding (`size-8!` + the `lg` size's `p-0!`). That is right
              // for a one-line nav row and wrong here — this row is 48px and
              // carries a 28px avatar, so shrinking it moved the avatar and
              // changed the row's height in the same frame. These override it.
              className="h-12 gap-2.5 px-2 data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground group-data-[collapsible=icon]:h-12! group-data-[collapsible=icon]:w-full! group-data-[collapsible=icon]:px-2!"
            >
              {/* A control's own fill and a hairline, not a photo well — the
                  avatar is a fallback monogram far more often than an image. */}
              <Avatar className="size-7 shrink-0 rounded-sm border border-border bg-control">
                <AvatarImage src={user.avatar} alt="" />
                <AvatarFallback className="rounded-sm bg-transparent text-label font-medium text-fg-2">
                  {initials(user.name)}
                </AvatarFallback>
              </Avatar>
              <div className="rail-x grid flex-1 overflow-hidden text-left">
                <span className="truncate text-body font-medium leading-5">{user.name}</span>
                <span className="truncate text-meta font-normal leading-4 text-fg-muted">
                  {user.organisation || user.email}
                </span>
              </div>
              <ChevronsUpDown className="rail-x ml-auto size-4 shrink-0 text-fg-2" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
            side={isMobile ? "bottom" : "right"}
            align="end"
            sideOffset={4}
          >
            <DropdownMenuItem onClick={() => logoutAndRedirect()}>
              <LogOut />
              Log out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
