"use client"

import { Link, useLocation } from "react-router-dom"
import { Users, FolderKanban, KeyRound } from "lucide-react"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavSettings() {
  const location = useLocation();
  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton tooltip="Users" asChild isActive={location.pathname.startsWith("/settings/users")}>
          <Link to="/settings/users">
            <Users />
            <span>Users</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
      <SidebarMenuItem>
        <SidebarMenuButton tooltip="Projects" asChild isActive={location.pathname.startsWith("/settings/projects")}>
          <Link to="/settings/projects">
            <FolderKanban />
            <span>Projects</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
      <SidebarMenuItem>
        <SidebarMenuButton tooltip="API Tokens" asChild isActive={location.pathname.startsWith("/settings/api-tokens")}>
          <Link to="/settings/api-tokens">
            <KeyRound />
            <span>API Tokens</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
