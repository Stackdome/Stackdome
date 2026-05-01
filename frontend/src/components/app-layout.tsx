import * as React from "react";
import { SidebarProvider, SidebarTrigger, SidebarInset } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import { Link, Outlet, useLocation } from "react-router-dom";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator
} from "@/components/ui/breadcrumb";
import { BreadcrumbProvider } from "@/contexts/breadcrumb-context";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { ThemeProvider } from "@/contexts/theme-provider";
import { ThemeToggle } from "@/components/theme-toggle";
import { Separator } from "@/components/ui/separator";

interface BreadcrumbItemType {
  name: string;
  path: string;
  clickable: boolean;
}

function AppLayoutContent({
  children,
}: {
  children?: React.ReactNode;
}) {
  const location = useLocation();
  const { customLabels, loadingLabels, nonClickablePaths } = useBreadcrumb();

  // Parse the current path for breadcrumbs
  const pathSegments = location.pathname.split('/').filter(Boolean);

  // Create breadcrumb items based on the current path
  const breadcrumbItems: BreadcrumbItemType[] = [
    { name: 'Home', path: '/', clickable: true },
    ...pathSegments.map((segment, index): BreadcrumbItemType => {
      const path = '/' + pathSegments.slice(0, index + 1).join('/');
      const clickable = !nonClickablePaths[path];
      // If it's the last segment and loading, show "..."
      if (index === pathSegments.length - 1 && loadingLabels && loadingLabels[path]) {
        return { name: "...", path, clickable };
      }
      // Custom label takes precedence for any segment that registers one
      if (customLabels[path]) {
        return { name: customLabels[path], path, clickable };
      }
      // Otherwise, capitalize the segment
      const name = segment.charAt(0).toUpperCase() + segment.slice(1);
      return { name, path, clickable };
    }),
  ];

  return (
    <SidebarProvider>
      <div className="flex h-screen max-h-screen w-full overflow-hidden bg-background">
        <AppSidebar />
        <SidebarInset>
          <div className="flex-shrink-0 bg-background rounded-tl-lg rounded-tr-lg">
            <div className="flex items-center justify-between p-4">
              <div className="flex items-center gap-2">
                <SidebarTrigger />
                <div className="border-l-2 h-4 w-0 mx-2" />
                <Breadcrumb>
                  <BreadcrumbList>
                    {breadcrumbItems.map((item, index) => (
                      <React.Fragment key={index}>
                        {index > 0 && <BreadcrumbSeparator />}
                        {index === breadcrumbItems.length - 1 ? (
                          <BreadcrumbItem>
                            <BreadcrumbPage>{item.name}</BreadcrumbPage>
                          </BreadcrumbItem>
                        ) : !item.clickable ? (
                          <BreadcrumbItem>
                            <span className="text-muted-foreground">{item.name}</span>
                          </BreadcrumbItem>
                        ) : (
                          <BreadcrumbItem>
                            <BreadcrumbLink asChild>
                              <Link to={item.path}>{item.name}</Link>
                            </BreadcrumbLink>
                          </BreadcrumbItem>
                        )}
                      </React.Fragment>
                    ))}
                  </BreadcrumbList>
                </Breadcrumb>
              </div>
              <ThemeToggle />
            </div>
            <Separator />
          </div>

          {/* Scrollable content area with padding and max-width */}
          <div className="flex-grow overflow-auto scrollbar-hide rounded-bl-lg rounded-br-lg flex justify-center items-start p-6">
            <div className="w-full max-w-6xl">
              {children ? children : <Outlet />}
            </div>
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
