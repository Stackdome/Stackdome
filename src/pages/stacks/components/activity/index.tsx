import { useParams, Link } from "react-router-dom";
import { useStacks } from "@/pages/stacks/contexts/stack-context";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import { Card, CardContent } from "@/components/ui/card";
import { 
  Activity as ActivityIcon, 
  Play, 
  Square, 
  RotateCw, 
  Settings, 
  Upload, 
  UserCircle2
} from "lucide-react";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import { StackSidebar } from "@/pages/stacks/components/shared/stack-sidebar";

export default function StackActivityPage() {
  const { id } = useParams();
  const { stacks } = useStacks();
  
  // Find the current stack
  const currentStack = stacks.find(stack => stack.id === id);
  
  if (!currentStack) {
    return (
      <div className="p-8 text-center">
        <h2 className="text-xl font-semibold mb-2">Stack not found</h2>
        <p className="text-muted-foreground mb-4">The stack you're looking for doesn't exist or has been deleted.</p>
        <Link to="/stacks" className="text-primary hover:underline">Return to Stacks</Link>
      </div>
    );
  }
  
  // Mock activity data
  const activities = [
    { 
      id: 1, 
      type: 'deploy', 
      user: 'Akshay',
      timestamp: 'May 6, 2025 14:30:25',
      icon: <Upload className="h-4 w-4" />,
      detail: 'Deployed stack from branch main'
    },
    { 
      id: 2, 
      type: 'start', 
      user: 'Akshay',
      timestamp: 'May 6, 2025 14:31:12',
      icon: <Play className="h-4 w-4" />,
      detail: 'Started all services'
    },
    { 
      id: 3, 
      type: 'config', 
      user: 'System',
      timestamp: 'May 6, 2025 14:35:00',
      icon: <Settings className="h-4 w-4" />,
      detail: 'Auto-scaling triggered: increased to 3 replicas'
    },
    { 
      id: 4, 
      type: 'restart', 
      user: 'Akshay',
      timestamp: 'May 6, 2025 16:42:18',
      icon: <RotateCw className="h-4 w-4" />,
      detail: 'Restarted backend service'
    },
    { 
      id: 5, 
      type: 'stop', 
      user: 'Akshay',
      timestamp: 'May 6, 2025 22:15:45',
      icon: <Square className="h-4 w-4" />,
      detail: 'Stopped database service for maintenance'
    },
    { 
      id: 6, 
      type: 'start', 
      user: 'Akshay',
      timestamp: 'May 6, 2025 22:30:10',
      icon: <Play className="h-4 w-4" />,
      detail: 'Restarted database service after maintenance'
    },
  ];
  
  const getIconColor = (type: string) => {
    switch(type) {
      case 'deploy':
        return 'text-blue-500 bg-blue-50';
      case 'start':
        return 'text-green-500 bg-green-50';
      case 'config':
        return 'text-orange-500 bg-orange-50';
      case 'restart':
        return 'text-purple-500 bg-purple-50';
      case 'stop':
        return 'text-red-500 bg-red-50';
      default:
        return 'text-gray-500 bg-gray-50';
    }
  };
  
  return (
    <SidebarProvider>
      <StackSidebar />
      <SidebarInset>
        <div className="p-6">
          <header className="mb-6">
            <Breadcrumb className="mb-3">
              <BreadcrumbList>
                <BreadcrumbItem>
                  <BreadcrumbLink href="/stacks">Stacks</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                  <BreadcrumbLink href={`/stacks/${id}`}>{currentStack.name}</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                  <BreadcrumbLink>Activity</BreadcrumbLink>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
            
            <div className="flex items-center">
              <ActivityIcon className="mr-2 h-5 w-5" />
              <h1 className="text-2xl font-bold">Activity Log</h1>
            </div>
            <p className="text-muted-foreground mt-1">
              Recent activity and events for {currentStack.name}
            </p>
            <Separator className="mt-4" />
          </header>
          
          <Card>
            <CardContent className="p-0">
              <div className="divide-y">
                {activities.map((activity) => (
                  <div key={activity.id} className="p-4 flex">
                    <div className={`${getIconColor(activity.type)} p-2 rounded-full mr-4 self-start`}>
                      {activity.icon}
                    </div>
                    <div className="flex-1">
                      <div className="flex justify-between">
                        <h3 className="font-medium">{activity.detail}</h3>
                        <span className="text-sm text-muted-foreground">{activity.timestamp}</span>
                      </div>
                      <div className="flex items-center mt-1 text-sm text-muted-foreground">
                        <UserCircle2 className="h-3.5 w-3.5 mr-1" />
                        <span>{activity.user}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
