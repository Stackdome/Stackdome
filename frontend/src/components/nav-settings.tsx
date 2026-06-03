"use client"

import { Link, useLocation } from "react-router-dom"
import { Users, FolderKanban } from "lucide-react"

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
        <SidebarMenuButton tooltip="Teams" asChild isActive={location.pathname.startsWith("/settings/teams")}>
          <Link to="/settings/teams">
            <FolderKanban />
            <span>Teams</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
