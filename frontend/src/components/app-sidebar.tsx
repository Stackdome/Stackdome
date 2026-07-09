import * as React from "react"
import { Link } from "react-router-dom"
import { StackdomeMark } from "@/components/branded"

import { NavStacks } from "@/components/nav-stacks"
import { NavPreviews } from "@/components/nav-previews"
import { NavClusters } from "@/components/nav-clusters"
import { NavSecrets } from "@/components/nav-secrets"
import { NavObjectStores } from "@/components/nav-object-stores"
import { NavDomains } from "@/components/nav-domains"
import { NavAddons } from "@/components/nav-addons"
import { NavUser } from "@/components/nav-user"
import { getCurrentUser } from "@/helpers/common"
import { useCurrentUser } from "@/hooks/use-current-user"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarGroupContent,
} from "@/components/ui/sidebar"

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const user = getCurrentUser();
  const { isOrgAdmin } = useCurrentUser();

  // Create required user data for NavUser component
  const userData = {
    name: user?.name || user?.username || "User",
    email: user?.email || "user@example.com",
    avatar: "",
  };

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader className="px-2 py-3">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild tooltip="Stackdome">
              <Link to="/">
                <div className="flex aspect-square size-8 shrink-0 items-center justify-center rounded-md border border-brand-border bg-brand-bg">
                  <StackdomeMark size={18} />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold lowercase" style={{ letterSpacing: "0.04em" }}>stackdome</span>
                  <span className="truncate text-[10px] font-mono uppercase tracking-[0.5px] text-muted-foreground">{user?.organisation}</span>
                </div>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel className="font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground/70">Platform</SidebarGroupLabel>
          <SidebarGroupContent>
            <NavStacks />
            <NavPreviews />
            <NavSecrets />
            <NavObjectStores />
            <NavAddons />
            {/* Org-scoped, admin-only resources. Hidden for members so they
                don't see (or fetch) endpoints that return 403 for them. */}
            {isOrgAdmin && <NavClusters />}
            {isOrgAdmin && <NavDomains />}
          </SidebarGroupContent>
        </SidebarGroup>
        {/* Settings group (Users + Teams) is shelved — nav hidden and routes
            redirected in App.tsx. Components/pages remain in the repo. */}
      </SidebarContent>

      <SidebarFooter>
        <NavUser user={userData} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
