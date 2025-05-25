import * as React from "react";
import {
  Command,
  PlusCircle,
} from "lucide-react";
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
} from "@/components/ui/sidebar";
import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from "@/components/ui/accordion";
import { NavUser } from "@/components/nav-user";
import { Link } from "react-router-dom";
import { cn } from "@/lib/utils";
import { getCurrentUser } from "@/helpers/common";

// SidebarSection type for navigation
export interface SidebarSection {
  label: string;
  icon: React.ReactNode;
  items: Array<{
    label: string;
    icon?: React.ReactNode;
    href: string;
    active?: boolean;
    onClick?: () => void;
  }>;
  addHref?: string;
  addLabel?: string;
}

interface ProjectSidebarProps {
  sections: SidebarSection[];
  children?: React.ReactNode;
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function ProjectSidebar({ sections, children }: ProjectSidebarProps) {
  return (
    <Sidebar variant="inset">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <Link to="/">
                <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                  <Command className="size-4" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">StackDome</span>
                  <span className="truncate text-xs">{getCurrentUser()?.organisation}</span>
                </div>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <div className="h-7" />
        <Accordion type="multiple" className="w-full" defaultValue={sections.map(s => s.label.toLowerCase())}>
          {sections.map(section => (
            <AccordionItem value={section.label.toLowerCase()} key={section.label}>
              <AccordionTrigger className="px-2 py-1 rounded-md">
                <span className="flex items-center gap-2">
                  {section.icon}
                  <span>{section.label}</span>
                </span>
              </AccordionTrigger>
              <AccordionContent>
                <SidebarMenuSub>
                  {section.addHref && (
                    <SidebarMenuSubItem>
                      <SidebarMenuSubButton asChild>
                        <Link to={section.addHref}>
                          <PlusCircle className="size-4" />
                          <span>{section.addLabel || `Add new ${section.label}`}</span>
                        </Link>
                      </SidebarMenuSubButton>
                    </SidebarMenuSubItem>
                  )}
                  {section.items.map(item => (
                    <SidebarMenuSubItem key={item.label}>
                      <SidebarMenuSubButton asChild>
                        <Link to={item.href} className={cn(item.active ? "bg-primary/10 text-primary" : "")}
                          onClick={item.onClick}
                        >
                          {item.icon}
                          <span>{item.label}</span>
                        </Link>
                      </SidebarMenuSubButton>
                    </SidebarMenuSubItem>
                  ))}
                </SidebarMenuSub>
              </AccordionContent>
            </AccordionItem>
          ))}
        </Accordion>
      </SidebarContent>
      <SidebarFooter>
        <NavUser />
      </SidebarFooter>
    </Sidebar>
  );
}
