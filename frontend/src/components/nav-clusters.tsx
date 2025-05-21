"use client"

import { Link, useLocation } from "react-router-dom"
import { Boxes } from "lucide-react"

import {
  SidebarGroup,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSkeleton,
} from "@/components/ui/sidebar"
import { useClusters } from "@/pages/clusters/hooks/use-clusters"
import { cn } from "@/lib/utils"

export function NavClusters() {
  const location = useLocation();
  const { clusters, loading } = useClusters();
  const hasCluster = !loading && clusters.length > 0;
  const isClustersActive = location.pathname.startsWith('/clusters');

  if (loading) {
    return (
      <SidebarGroup>
        <SidebarMenu>
          <SidebarMenuSkeleton />
        </SidebarMenu>
      </SidebarGroup>
    );
  }

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton
          tooltip="Clusters"
          asChild
          isActive={isClustersActive}
        >
          <Link to={hasCluster && clusters.length === 1 ? `/clusters/${clusters[0].id}` : "/clusters"}>
            <Boxes/>
            <span>Clusters</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>

      {hasCluster && clusters.length > 1 && clusters.map(cluster => (
        <SidebarMenuItem key={cluster.id}>
          <SidebarMenuButton
            tooltip={cluster.name || 'Detail'}
            asChild
            isActive={location.pathname.includes(`/clusters/${cluster.id}`)}
            className={cn("text-sidebar-foreground/70 text-sm")}
          >
            <Link to={`/clusters/${cluster.id}`}>
              <span>{cluster.name || 'Detail'}</span>
            </Link>
          </SidebarMenuButton>
        </SidebarMenuItem>
      ))}
    </SidebarMenu>
  );
}
