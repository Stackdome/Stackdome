import * as React from "react"
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
} from "@/components/ui/sidebar"
import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from "@/components/ui/accordion"
import { NavUser } from "@/components/nav-user"

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar variant="inset" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <a href="#">
                <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                  <Command className="size-4" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">StackDome</span>
                  <span className="truncate text-xs">PaaS Platform</span>
                </div>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <div className="h-7" />
        <Accordion type="multiple" className="w-full" defaultValue={["stacks"]}>
          <AccordionItem value="stacks">
            <AccordionTrigger className="px-2 py-1 rounded-md">
              <span className="flex items-center gap-2">
                <Layers className="size-4" />
                <span>Stacks</span>
              </span>
            </AccordionTrigger>
            <AccordionContent>
              <SidebarMenuSub>
                <SidebarMenuSubItem>
                  <SidebarMenuSubButton asChild>
                    <a href="/stacks/create">
                      <Plus className="size-2" />
                      <span>Add new Stack</span>
                    </a>
                  </SidebarMenuSubButton>
                </SidebarMenuSubItem>
                {/* Existing stacks would be mapped here in the future */}
              </SidebarMenuSub>
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="clusters">
            <AccordionTrigger className="px-2 py-1 rounded-md">
              <span className="flex items-center gap-2">
                <Cloud className="size-4" />
                <span>Clusters</span>
              </span>
            </AccordionTrigger>
            <AccordionContent>
              <SidebarMenuSub>
                <SidebarMenuSubItem>
                  <SidebarMenuSubButton asChild>
                    <a href="/clusters">
                      <Plus className="size-4" />
                      <span>Add new Cluster</span>
                    </a>
                  </SidebarMenuSubButton>
                </SidebarMenuSubItem>
                {/* Existing clusters would be mapped here in the future */}
              </SidebarMenuSub>
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </SidebarContent>
      <SidebarFooter>
        <NavUser />
      </SidebarFooter>
    </Sidebar>
  )
}
