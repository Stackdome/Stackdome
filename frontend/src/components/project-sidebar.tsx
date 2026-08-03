import * as React from "react";
import { PlusCircle } from "lucide-react";
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
import { StackdomeMark } from "@/components/branded";
import { Link } from "react-router-dom";
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
}

export function ProjectSidebar({ sections }: ProjectSidebarProps) {
  const currentUser = getCurrentUser();
  const navUser = {
    name: currentUser?.name ?? "",
    email: currentUser?.email ?? "",
    avatar: "",
  };
  return (
    <Sidebar variant="inset">
      <SidebarHeader className="px-2 py-3">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild tooltip="Stackdome">
              <Link to="/">
                <div className="flex aspect-square size-8 shrink-0 items-center justify-center rounded-md border border-brand-border bg-brand-bg">
                  <StackdomeMark size={18} />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold lowercase" style={{ letterSpacing: "0.04em" }}>stackdome</span>
                  <span className="truncate text-[10px] font-mono uppercase tracking-[0.5px] text-muted-foreground">{getCurrentUser()?.organisation}</span>
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
                      <SidebarMenuSubButton asChild isActive={item.active}>
                        <Link to={item.href} onClick={item.onClick}>
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
        <NavUser user={navUser} />
      </SidebarFooter>
    </Sidebar>
  );
}
