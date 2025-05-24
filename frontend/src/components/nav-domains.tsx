"use client"

import { Link, useLocation } from "react-router-dom"
import { Globe } from "lucide-react"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavDomains() {
  const location = useLocation();
  const isDomainsActive = location.pathname.startsWith('/domains');

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton
          tooltip="Domains"
          asChild
          isActive={isDomainsActive}
        >
          <Link to="/domains">
            <Globe />
            <span>Domains</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
