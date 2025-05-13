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
import { Button } from "@/components/ui/button";
import { PlusCircle } from "lucide-react";
import { StackCreationModal } from "@/pages/stacks/components/shared/stack-creation-modal";

export function AppLayout({
  children,
}: {
  children?: React.ReactNode;
}) {
  const location = useLocation();

  // Parse the current path for breadcrumbs
  const pathSegments = location.pathname.split('/').filter(Boolean);
  
  // Create breadcrumb items based on the current path
  const breadcrumbItems = [
    { name: 'Home', path: '/' },
    ...pathSegments.map((segment, index) => {
      // Create a path up to this segment
      const path = '/' + pathSegments.slice(0, index + 1).join('/');
      // Capitalize the segment for display
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
          </div>
          <div className="flex-1 p-4">
            {children ? children : <Outlet />}
          </div>
        </SidebarInset>
      </div>
    </SidebarProvider>
  );
}
