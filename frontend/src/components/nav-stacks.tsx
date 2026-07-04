"use client"

import { Link, useLocation } from "react-router-dom"
import { Layers } from "lucide-react"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavStacks() {
  const location = useLocation();
  const isStacksActive = location.pathname.startsWith('/stacks');

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton
          tooltip="Stacks"
          asChild
          isActive={isStacksActive && !location.pathname.includes('/stacks/new')}
        >
          <Link to="/stacks">
            <Layers />
            <span>Stacks</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
