"use client"

import { Link, useLocation } from "react-router-dom"
import { GitPullRequest } from "lucide-react"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavPreviews() {
  const location = useLocation();
  const isActive = location.pathname.startsWith('/previews');

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton
          tooltip="Preview Environments"
          asChild
          isActive={isActive}
        >
          <Link to="/previews">
            <GitPullRequest />
            <span>Preview Environments</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
