"use client"

import { Link, useLocation } from "react-router-dom"
import { Github } from "lucide-react"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavGitIntegrations() {
  const location = useLocation();
  const isGitIntegrationsActive = location.pathname.startsWith('/git-integrations');

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton
          tooltip="Git Integrations"
          asChild
          isActive={isGitIntegrationsActive}
        >
          <Link to="/git-integrations">
            <Github />
            <span>Git Integrations</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
