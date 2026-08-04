"use client"

import { Link, useLocation } from "react-router-dom"
import { Layers } from "lucide-react"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"
import { usePreviewLineage } from "@/hooks/use-preview-lineage"

export function NavStacks() {
  const location = useLocation();
  const { lineage } = usePreviewLineage();
  // A preview stack hands its highlight to Previews.
  const isStacksActive = location.pathname.startsWith('/stacks') && !lineage;

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
