"use client"

import { Link, useLocation } from "react-router-dom"
import { Puzzle } from "lucide-react"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavAddons() {
  const location = useLocation();
  const isAddonsActive = location.pathname.startsWith('/addons');

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton
          tooltip="Addons"
          asChild
          isActive={isAddonsActive}
        >
          <Link to="/addons">
            <Puzzle />
            <span>Addons</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
