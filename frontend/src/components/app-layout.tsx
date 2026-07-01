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
import { ThemeToggle } from "@/components/theme-toggle";
import { Separator } from "@/components/ui/separator";
import { isCanvasEnabled } from "@/lib/feature-flags";

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

  // Full-bleed layout for the canvas stack editor: /stacks/<id> (a single id
  // segment — not /stacks or /stacks/new), and only when the flag is on.
  const isFullBleed =
    isCanvasEnabled() &&
    /^\/stacks\/[^/]+$/.test(location.pathname) &&
    !location.pathname.endsWith("/new");

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
                  <BreadcrumbList className="font-mono text-[12px] gap-2 sm:gap-2">
                    {breadcrumbItems.map((item, index) => (
                      <React.Fragment key={index}>
                        {index > 0 && (
                          <BreadcrumbSeparator className="text-muted-foreground/50 [&>svg]:hidden">
                            <span>/</span>
                          </BreadcrumbSeparator>
                        )}
                        {index === breadcrumbItems.length - 1 ? (
                          <BreadcrumbItem>
                            <BreadcrumbPage className="text-foreground">{item.name}</BreadcrumbPage>
                          </BreadcrumbItem>
                        ) : !item.clickable ? (
                          <BreadcrumbItem>
                            <span className="text-muted-foreground">{item.name}</span>
                          </BreadcrumbItem>
                        ) : (
                          <BreadcrumbItem>
                            <BreadcrumbLink asChild className="text-muted-foreground hover:text-brand transition-colors">
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

          {/* Scrollable content area with padding and max-width.
              The page-sticky-bar slot lives at the top of the scroll container
              so a sticky element inside it pins flush under the topnav and
              spans the full width of the inset (no max-w cap). Pages portal
              into it via #page-sticky-bar.

              The canvas stack editor opts out of the centered max-width column
              and renders full-bleed (edge-to-edge, full height). Gated on the
              feature flag + the stack-detail route so every other page keeps
              the standard constrained layout. */}
          <div className="flex-grow overflow-auto scrollbar-hide rounded-bl-lg rounded-br-lg">
            <div id="page-sticky-bar" className="sticky top-0 z-30" />
            {isFullBleed ? (
              <div className="h-full">{children ? children : <Outlet />}</div>
            ) : (
              <div className="flex justify-center items-start p-6">
                <div className="w-full max-w-6xl">
                  {children ? children : <Outlet />}
                </div>
              </div>
            )}
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
    <BreadcrumbProvider>
      <AppLayoutContent children={children} />
    </BreadcrumbProvider>
  );
}
