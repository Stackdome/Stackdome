import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { ProjectSidebar } from "@/components/project-sidebar";
import type { SidebarSection } from "@/components/project-sidebar";
import { Outlet } from "react-router-dom";
import { Layers, Cloud } from "lucide-react";

// Default sidebar sections for the app (can be customized per page if needed)
const defaultSections: SidebarSection[] = [
  {
    label: "Stacks",
    icon: <Layers className="size-4" />,
    addHref: "/stacks/create",
    addLabel: "Add new Stack",
    items: [],
  },
  {
    label: "Clusters",
    icon: <Cloud className="size-4" />,
    addHref: "/clusters",
    addLabel: "Add new Cluster",
    items: [],
  },
];

export function AppLayout({
  children,
  sections = defaultSections,
}: {
  children?: React.ReactNode;
  sections?: SidebarSection[];
}) {
  return (
    <SidebarProvider>
      <div className="flex min-h-screen w-full">
        <ProjectSidebar sections={sections} />
        <main className="flex-1 flex flex-col">
          {/* Global sidebar trigger at the top left of the main content */}
          <div className="h-14 flex items-center px-4 border-b">
            <SidebarTrigger className="mr-2" />
            {/* You can add a global breadcrumb or title here if desired */}
          </div>
          <div className="flex-1">
            {children ? children : <Outlet />}
          </div>
        </main>
      </div>
    </SidebarProvider>
  );
}
