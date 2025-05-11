import { Link, useLocation, useParams } from "react-router-dom";
import { 
  BarChart3, 
  Activity, 
  Settings, 
  Database, 
  Server, 
  Globe,
  Layers
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useStacks } from "@/pages/stacks/contexts/stack-context";

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { NavUser } from "@/components/nav-user";

export function StackSidebar() {
  const { id } = useParams();
  const location = useLocation();
  const { stacks } = useStacks();
  
  // Find the current stack
  const currentStack = stacks.find(stack => stack.id === id);
  
  // Get path segments for determining active state
  const path = location.pathname;
  const isOverview = path === `/stacks/${id}`;
  const isActivity = path === `/stacks/${id}/activity`;
  const isSettings = path === `/stacks/${id}/settings`;
  
  // Mock service data (in real app, this would come from stack data)
  const services = [
    { name: "frontend", type: "web" },
    { name: "backend", type: "api" },
    { name: "database", type: "db" }
  ];
  
  const getServiceIcon = (type: string) => {
    switch(type) {
      case "web":
        return <Globe className="h-4 w-4" />;
      case "api":
        return <Server className="h-4 w-4" />;
      case "db":
        return <Database className="h-4 w-4" />;
      default:
        return <Server className="h-4 w-4" />;
    }
  };

  const data = {
    user: {
      name: "shadcn",
      email: "m@example.com",
      avatar: "/avatars/shadcn.jpg",
    },
  };
  
  return (
    <Sidebar variant="inset">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <Link to="/">
                <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                  <Layers className="size-4" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">StackDome</span>
                  <span className="truncate text-xs">PaaS Platform</span>
                </div>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      
      <SidebarContent>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton asChild>
              <Link to="/stacks" className="bg-primary/10 text-primary">
                <Layers className="size-4" />
                <span>Stacks</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
          
          {currentStack && (
            <>
              <div className="mt-6 mb-2 px-4">
                <h3 className="font-medium text-xs uppercase text-muted-foreground">Current Stack</h3>
              </div>
              
              <SidebarMenuItem>
                <SidebarMenuButton asChild>
                  <Link 
                    to={`/stacks/${id}`} 
                    className={cn(isOverview && !location.search ? "bg-primary/10 text-primary" : "")}
                  >
                    <BarChart3 className="size-4" />
                    <span>Overview</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
              
              {services.map(service => (
                <SidebarMenuItem key={service.name}>
                  <SidebarMenuButton asChild>
                    <Link 
                      to={`/stacks/${id}?service=${service.name}`}
                      className={cn(
                        location.search.includes(`service=${service.name}`) ? "bg-primary/10 text-primary" : ""
                      )}
                    >
                      {getServiceIcon(service.type)}
                      <span>{service.name}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
              
              <SidebarMenuItem>
                <SidebarMenuButton asChild>
                  <Link 
                    to={`/stacks/${id}/activity`} 
                    className={cn(isActivity ? "bg-primary/10 text-primary" : "")}
                  >
                    <Activity className="size-4" />
                    <span>Activity</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
              
              <SidebarMenuItem>
                <SidebarMenuButton asChild>
                  <Link 
                    to={`/stacks/${id}/settings`} 
                    className={cn(isSettings ? "bg-primary/10 text-primary" : "")}
                  >
                    <Settings className="size-4" />
                    <span>Settings</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </>
          )}
        </SidebarMenu>
      </SidebarContent>
      
      <SidebarFooter>
        <NavUser user={data.user} />
      </SidebarFooter>
    </Sidebar>
  );
}
