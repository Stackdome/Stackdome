"use client"

import { Link, useLocation } from "react-router-dom"
import { Container } from "lucide-react"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavImageRegistries() {
  const location = useLocation();
  const isImageRegistriesActive = location.pathname.startsWith('/image-registries');

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton
          tooltip="Image Registries"
          asChild
          isActive={isImageRegistriesActive}
        >
          <Link to="/image-registries">
            <Container />
            <span>Image Registries</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
