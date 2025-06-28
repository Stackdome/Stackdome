"use client"

import { Link, useLocation } from "react-router-dom"
import { KeyRound } from "lucide-react"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavSecrets() {
  const location = useLocation();
  const isSecretsActive = location.pathname.startsWith('/secrets');

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton
          tooltip="Secrets"
          asChild
          isActive={isSecretsActive}
        >
          <Link to="/secrets">
            <KeyRound />
            <span>Secrets</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
