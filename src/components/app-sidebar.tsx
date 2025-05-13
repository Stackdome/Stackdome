import * as React from "react"
import { useLocation } from "react-router-dom"
import {
  Command,
  Layers,
  Plus,
  Cloud
} from "lucide-react"

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubItem,
  SidebarMenuSubButton,
  SidebarMenuSkeleton,
} from "@/components/ui/sidebar"
import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from "@/components/ui/accordion"
import { useClusters } from "@/pages/clusters/hooks/use-clusters"
import { cn } from "@/lib/utils"
import { NavUser } from "@/components/nav-user"
import { getCurrentUser } from "@/helpers/common"

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  // Get clusters to conditionally show "Add new Cluster" option
  const { clusters, loading } = useClusters();
  const hasCluster = !loading && clusters.length > 0;
  const location = useLocation();
  
  return (
    <Sidebar variant="inset" className="bg-sidebar text-sidebar-foreground" {...props}>
      <SidebarHeader className="px-2 py-3">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <a href="/">
                <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                  <Command className="size-4" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">Stackdome</span>
                  <span className="truncate text-xs text-muted-foreground">{getCurrentUser()?.organisation}</span>
                </div>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      
      <SidebarContent className="px-2 py-2">
        <Accordion type="multiple" className="w-full" defaultValue={["stacks", "clusters"]}>
          <AccordionItem value="stacks" className="border-none">
            <AccordionTrigger className="px-2 py-1.5 text-sm text-sidebar-foreground hover:text-sidebar-accent-foreground hover:bg-sidebar-accent hover:no-underline rounded-md">
              <span className="flex items-center gap-2">
                <Layers className="size-4" />
                <span>Stacks</span>
              </span>
            </AccordionTrigger>
            <AccordionContent className="pb-1">
              <SidebarMenuSub>
                <SidebarMenuSubItem>
                  <SidebarMenuSubButton 
                    asChild
                    className={cn(
                      "text-sidebar-foreground/70 text-sm hover:text-sidebar-foreground hover:bg-sidebar-accent",
                      location.pathname.includes("/stacks/create") && "text-sidebar-foreground bg-sidebar-accent"
                    )}
                  >
                    <a href="/stacks/create">
                      <Plus className="size-3.5" />
                      <span>Add new Stack</span>
                    </a>
                  </SidebarMenuSubButton>
                </SidebarMenuSubItem>
                {/* Existing stacks would be mapped here in the future */}
              </SidebarMenuSub>
            </AccordionContent>
          </AccordionItem>
          
          <AccordionItem value="clusters" className="border-none">
            <AccordionTrigger className="px-2 py-1.5 text-sm text-sidebar-foreground hover:text-sidebar-accent-foreground hover:bg-sidebar-accent hover:no-underline rounded-md">
              <span className="flex items-center gap-2">
                <Cloud className="size-4" />
                <span>Clusters</span>
              </span>
            </AccordionTrigger>
            <AccordionContent className="pb-1">
              {loading ? (
                <SidebarMenuSkeleton />
              ) : (
                <SidebarMenuSub>
                  <SidebarMenuSubItem>
                    <SidebarMenuSubButton
                      asChild
                      className={cn(
                        "text-sidebar-foreground/70 text-sm hover:text-sidebar-foreground hover:bg-sidebar-accent",
                        (location.pathname === "/clusters" || location.pathname === "/clusters/") && "text-sidebar-foreground bg-sidebar-accent"
                      )}
                    >
                      <a href={hasCluster && clusters.length === 1 ? `/clusters/${clusters[0].id}` : "/clusters"}>
                        <span>Overview</span>
                      </a>
                    </SidebarMenuSubButton>
                  </SidebarMenuSubItem>
                  {hasCluster && clusters.length > 1 && clusters.map(cluster => (
                    <SidebarMenuSubItem key={cluster.id}>
                      <SidebarMenuSubButton 
                        asChild
                        className={cn(
                          "text-sidebar-foreground/70 text-sm hover:text-sidebar-foreground hover:bg-sidebar-accent",
                          location.pathname.includes(`/clusters/${cluster.id}`) && "text-sidebar-foreground bg-sidebar-accent"
                        )}
                      >
                        <a href={`/clusters/${cluster.id}`}>
                          <span>Detail</span>
                        </a>
                      </SidebarMenuSubButton>
                    </SidebarMenuSubItem>
                  ))}
                  <SidebarMenuSubItem>
                    {hasCluster ? (
                      <div
                        className={cn(
                          "text-sidebar-foreground/40 text-sm cursor-not-allowed flex items-center gap-2 px-2 py-1.5 rounded-md select-none",
                        )}
                        style={{ position: 'relative', overflow: 'visible' }}
                      >
                        <Plus className="size-3.5" />
                        <span>Add new Cluster</span>
                      </div>
                    ) : (
                      <SidebarMenuSubButton
                        asChild
                        className={cn(
                          "text-sidebar-foreground/70 text-sm hover:text-sidebar-foreground hover:bg-sidebar-accent",
                          location.pathname.includes("/clusters/create") && "text-sidebar-foreground bg-sidebar-accent"
                        )}
                      >
                        <a href="/clusters/create">
                          <Plus className="size-3.5" />
                          <span>Add new Cluster</span>
                        </a>
                      </SidebarMenuSubButton>
                    )}
                  </SidebarMenuSubItem>
                </SidebarMenuSub>
              )}
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </SidebarContent>
      
      <SidebarFooter className="p-2">
        <NavUser />
      </SidebarFooter>
    </Sidebar>
  )
}
