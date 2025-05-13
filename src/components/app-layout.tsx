import * as React from "react";
import { SidebarProvider, SidebarTrigger, SidebarInset } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import { Outlet, useLocation } from "react-router-dom";
import { 
  Breadcrumb, 
  BreadcrumbItem, 
  BreadcrumbLink, 
  BreadcrumbList, 
  BreadcrumbPage, 
  BreadcrumbSeparator 
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import { BreadcrumbProvider } from "@/contexts/breadcrumb-context";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { ThemeProvider } from "@/contexts/theme-provider";
import { ThemeToggle } from "@/components/theme-toggle";

interface BreadcrumbItemType {
  name: string;
  path: string;
}

function AppLayoutContent({
  children,
}: {
  children?: React.ReactNode;
}) {
  const location = useLocation();
  const { customLabels, loadingLabels } = useBreadcrumb(); // Added loadingLabels

  // Parse the current path for breadcrumbs
  const pathSegments = location.pathname.split('/').filter(Boolean);
  
  // Create breadcrumb items based on the current path
  const breadcrumbItems: BreadcrumbItemType[] = [
    { name: 'Home', path: '/' },
    ...pathSegments.map((segment, index): BreadcrumbItemType => {
      const path = '/' + pathSegments.slice(0, index + 1).join('/');
      // If it's the last segment and loading, show "..."
      if (index === pathSegments.length - 1 && loadingLabels && loadingLabels[path]) {
        return { name: "...", path };
      }
      // If there's a custom label for the last segment, use it
      if (index === pathSegments.length - 1 && customLabels[path]) {
        return { name: customLabels[path], path };
      }
      // Otherwise, capitalize the segment
      const name = segment.charAt(0).toUpperCase() + segment.slice(1);
      return { name, path };
    }),
  ];

  return (
    <SidebarProvider>
      <div className="flex min-h-screen w-full">
        <AppSidebar />
        <SidebarInset>
          <div className="flex items-center justify-between p-4">
            <div className="flex items-center gap-2">
              <SidebarTrigger />
              <Separator orientation="vertical" className="h-6" />
              <Breadcrumb>
                <BreadcrumbList>
                  {breadcrumbItems.map((item, index) => (
                    <React.Fragment key={index}>
                      {index > 0 && <BreadcrumbSeparator />}
                      {index === breadcrumbItems.length - 1 ? (
                        <BreadcrumbItem>
                          <BreadcrumbPage>{item.name}</BreadcrumbPage>
                        </BreadcrumbItem>
                      ) : (
                        <BreadcrumbItem>
                          <BreadcrumbLink href={item.path}>{item.name}</BreadcrumbLink>
                        </BreadcrumbItem>
                      )}
                    </React.Fragment>
                  ))}
                </BreadcrumbList>
              </Breadcrumb>
            </div>
            <ThemeToggle />
          </div>
          <div className="flex-1 p-4">
            {children ? children : <Outlet />}
          </div>
        </SidebarInset>
      </div>
    </SidebarProvider>
  );
}

export function AppLayout({
  children,
}: {
  children?: React.ReactNode;
}) {
  return (
    <ThemeProvider defaultTheme="system" storageKey="stackdome-ui-theme">
      <BreadcrumbProvider>
        <AppLayoutContent children={children} />
      </BreadcrumbProvider>
    </ThemeProvider>
  );
}
