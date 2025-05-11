import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { ProjectSidebar } from "@/components/project-sidebar";
import type { SidebarSection } from "@/components/project-sidebar";
import { Outlet } from "react-router-dom";
import { Layers, Cloud } from "lucide-react";
import { useClusters } from "@/pages/clusters/hooks/use-clusters";

export function AppLayout({
  children,
}: {
  children?: React.ReactNode;
}) {
  const { clusters, loading } = useClusters();
  const hasCluster = !loading && clusters.length > 0;

  const sections: SidebarSection[] = [
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
      addHref: hasCluster ? undefined : "/clusters", // Only show Add new Cluster if no cluster exists
      addLabel: "Add new Cluster",
      items: hasCluster ? [
        {
          label: clusters[0]?.name || "Cluster",
          href: `/clusters/${clusters[0]?.id}`,
          active: true
        }
      ] : [],
    },
  ];
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
