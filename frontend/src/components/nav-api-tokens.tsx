"use client"

import { Link, useLocation } from "react-router-dom"
import { KeySquare } from "lucide-react"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavApiTokens() {
  const location = useLocation();
  const isApiTokensActive = location.pathname.startsWith('/settings/api-tokens');

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton
          tooltip="API Tokens"
          asChild
          isActive={isApiTokensActive}
        >
          <Link to="/settings/api-tokens">
            <KeySquare />
            <span>API Tokens</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
