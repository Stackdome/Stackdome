import { Link, useLocation } from "react-router-dom"
import { GitPullRequest } from "lucide-react"

import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"
import { usePreviewLineage } from "@/hooks/use-preview-lineage"

export function NavPreviews() {
  const location = useLocation();
  const { lineage } = usePreviewLineage();
  // A preview stack renders under /stacks, but it belongs to Previews.
  const isActive = location.pathname.startsWith("/previews") || !!lineage;

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton tooltip="Previews" asChild isActive={isActive}>
          <Link to="/previews">
            <GitPullRequest />
            <span>Previews</span>
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
