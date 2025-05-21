"use client"

import { Link, useLocation } from "react-router-dom"
import { Layers, Plus } from "lucide-react"

import {
  SidebarGroup,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"
import { cn } from "@/lib/utils"

export function NavStacks() {
  const location = useLocation();
  const isStacksActive = location.pathname.startsWith('/stacks');

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton
          tooltip="Stacks"
          asChild
          isActive={isStacksActive && !location.pathname.includes('/stacks/create')}
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
