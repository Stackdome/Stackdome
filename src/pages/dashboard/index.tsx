import { AppSidebar } from "@/components/app-sidebar"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { Button } from "@/components/ui/button"
import { PlusCircle, Layers } from "lucide-react"
import { StackCreationModal } from "@/components/stacks/stack-creation-modal"

export default function Dashboard() {
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
          <div className="flex flex-col items-center justify-center h-[80vh] text-center">
            <div className="flex flex-col items-center max-w-md">
              <div className="rounded-full bg-primary/10 p-4 mb-4">
                <Layers className="h-10 w-10 text-primary" />
              </div>
              <h2 className="text-2xl font-bold mb-2">No stacks deployed yet</h2>
              <p className="text-muted-foreground mb-6">
                Deploy your first stack to get started. Create a stack-compose.yaml file and deploy
                your application with ease.
              </p>
              <StackCreationModal 
                trigger={
                  <Button>
                    <PlusCircle className="mr-2 h-4 w-4" />
                    Create New Stack
                  </Button>
                } 
              />
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
