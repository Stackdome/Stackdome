import * as React from "react"
import { Link } from "react-router-dom"
import { StackdomeMark } from "@/components/branded"

import { NavStacks } from "@/components/nav-stacks"
import { NavClusters } from "@/components/nav-clusters"
import { NavSecrets } from "@/components/nav-secrets"
import { NavDomains } from "@/components/nav-domains"
import { NavAddons } from "@/components/nav-addons"
import { NavUser } from "@/components/nav-user"
import { getCurrentUser } from "@/helpers/common"
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
                <div className="bg-brand-bg border border-brand-border flex aspect-square size-8 items-center justify-center rounded-md">
                  <StackdomeMark size={18} variant="tinted" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">Stackdome</span>
                  <span className="truncate text-[11px] font-mono uppercase tracking-[1.2px] text-muted-foreground">{user?.organisation}</span>
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
            <NavSecrets />
            <NavAddons />
            <NavClusters />
            <NavDomains />
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <NavUser user={userData} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
