import { Layers, PlusCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { AppSidebar } from "@/components/app-sidebar";
import { SidebarProvider, SidebarInset, SidebarTrigger } from "@/components/ui/sidebar";
import { Separator } from "@/components/ui/separator";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList } from "@/components/ui/breadcrumb";
import { StackCard } from "@/pages/stacks/components/shared/stack-card";
import { StackCreationModal } from "@/pages/stacks/components/shared/stack-creation-modal";
import { CreateSampleStackButton } from "@/pages/stacks/components/shared/create-sample-stack-button";
import { useStacks } from "@/pages/stacks/contexts/stack-context";

export default function StacksPage() {
  const { stacks, removeStack } = useStacks();
  
  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-2 h-4" />
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem>
                  <BreadcrumbLink href="#">
                    Stacks
                  </BreadcrumbLink>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
          <div className="ml-auto mr-4">
            <StackCreationModal 
              trigger={
                <Button size="sm">
                  <PlusCircle className="mr-2 h-4 w-4" />
                  New Stack
                </Button>
              } 
            />
          </div>
        </header>
        
        <div className="flex flex-1 flex-col p-4 pt-0 h-full">
          {stacks.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-[80vh] text-center">
              <div className="flex flex-col items-center max-w-md">
                <div className="rounded-full bg-primary/10 p-4 mb-4">
                  <Layers className="h-10 w-10 text-primary" />
                </div>
                <h2 className="text-2xl font-bold mb-2">No stacks deployed yet</h2>
                <p className="text-muted-foreground mb-6">
                  Deploy your first stack to get started. 
                </p>
                <StackCreationModal 
                  trigger={
                    <Button>
                      <PlusCircle className="mr-2 h-4 w-4" />
                      Create New Stack
                    </Button>
                  } 
                />
                
                {/* Add our test button for direct sample stack creation */}
                <CreateSampleStackButton />
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">
              {stacks.map((stack) => (
                <StackCard 
                  key={stack.id} 
                  stack={stack} 
                  onDelete={removeStack}
                />
              ))}
              
              <div className="flex items-center justify-center border-2 border-dashed rounded-lg p-8 h-full">
                <StackCreationModal 
                  trigger={
                    <Button variant="outline" className="w-full h-full py-8 flex flex-col items-center">
                      <PlusCircle className="h-8 w-8 mb-2" />
                      <span>Add New Stack</span>
                    </Button>
                  } 
                />
              </div>
            </div>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
