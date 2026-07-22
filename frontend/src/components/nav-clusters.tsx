"use client"

import { Link, useLocation } from "react-router-dom"
import { Boxes } from "lucide-react"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavClusters() {
  const location = useLocation();
  const isClustersActive = location.pathname.startsWith('/clusters');

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton
          tooltip="Clusters"
          asChild
          isActive={isClustersActive}
        >
          <Link to="/clusters">
            <Boxes/>
            <span>Clusters</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
