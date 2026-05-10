"use client"

import { Link, useLocation } from "react-router-dom"
import { Cloud } from "lucide-react"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavObjectStores() {
  const location = useLocation();
  const isObjectStoresActive = location.pathname.startsWith('/object-stores');

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton tooltip="Object Stores" asChild isActive={isObjectStoresActive}>
          <Link to="/object-stores">
            <Cloud />
            <span>Object Stores</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
