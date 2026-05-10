"use client"

import { Link, useLocation } from "react-router-dom"
import { Database } from "lucide-react"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavObjectStores() {
  const location = useLocation();
  const isActive = location.pathname.startsWith('/object-stores');

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton tooltip="Object Stores" asChild isActive={isActive}>
          <Link to="/object-stores">
            <Database />
            <span>Object Stores</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
