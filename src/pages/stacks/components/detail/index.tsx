import { useParams, Link, useSearchParams } from "react-router-dom";
import { useStacks } from "@/pages/stacks/contexts/stack-context";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import { Play, Pause, Maximize2, Minimize2, RotateCw, Clock, Terminal } from "lucide-react";
import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import { StackSidebar } from "@/pages/stacks/components/shared/stack-sidebar";

export default function StackDetailPage() {
  const { id } = useParams();
  const { stacks } = useStacks();
  const [searchParams] = useSearchParams();
  const selectedService = searchParams.get('service');
  const [isRunning, setIsRunning] = useState(true);
  const [isLogExpanded, setIsLogExpanded] = useState(false);
  
  // Find the current stack
  const currentStack = stacks.find(stack => stack.id === id);
  
  if (!currentStack) {
    return (
      <div className="p-8 text-center">
        <h2 className="text-xl font-semibold mb-2">Stack not found</h2>
        <p className="text-muted-foreground mb-4">The stack you're looking for doesn't exist or has been deleted.</p>
        <Button asChild>
          <Link to="/stacks">Return to Stacks</Link>
        </Button>
      </div>
    );
  }
  
  // Mock logs for the demo
  const logs = [
    { time: "10:15:32", service: "frontend", message: "Server started on port 3000" },
    { time: "10:15:33", service: "backend", message: "Connected to database" },
    { time: "10:15:35", service: "database", message: "Initializing database schema" },
    { time: "10:15:40", service: "frontend", message: "Proxying requests to backend" },
    { time: "10:15:45", service: "backend", message: "REST API ready on /api/v1" },
    { time: "10:16:00", service: "database", message: "Schema initialization complete" },
    { time: "10:16:10", service: "frontend", message: "Rendering application" },
    { time: "10:16:15", service: "backend", message: "Processing request: GET /api/v1/users" },
    { time: "10:16:16", service: "backend", message: "Request completed: 200 OK (10ms)" },
  ];
  
  // Mock metrics data
  const metrics = {
    cpu: "0.5%",
    memory: "256MB / 1GB",
    storage: "10GB / 20GB",
    network: "1.2 Mbps"
  };
  
  // Filter logs by selected service if applicable
  const filteredLogs = selectedService 
    ? logs.filter(log => log.service === selectedService)
    : logs;
  
  const toggleRunning = () => {
    setIsRunning(!isRunning);
  };
  
  const toggleLogExpanded = () => {
    setIsLogExpanded(!isLogExpanded);
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
                  <BreadcrumbLink>Overview</BreadcrumbLink>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
            
            <div className="flex justify-between items-center">
              <div>
                <h1 className="text-2xl font-bold">{currentStack.name}</h1>
                <p className="text-muted-foreground">{currentStack.description || "No description provided"}</p>
              </div>
              <div className="flex gap-3">
                <Button 
                  variant={isRunning ? "destructive" : "default"}
                  size="sm"
                  onClick={toggleRunning}
                >
                  {isRunning ? <Pause className="mr-2 h-4 w-4" /> : <Play className="mr-2 h-4 w-4" />}
                  {isRunning ? "Stop Stack" : "Start Stack"}
                </Button>
                <Button variant="outline" size="sm">
                  <RotateCw className="mr-2 h-4 w-4" />
                  Restart
                </Button>
              </div>
            </div>
            <Separator className="mt-4" />
          </header>
          
          <div className="grid grid-cols-3 gap-4 mb-6">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Stack Status</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex items-center">
                  <div className={`h-3 w-3 rounded-full mr-2 ${isRunning ? 'bg-green-500' : 'bg-gray-400'}`}></div>
                  <span className="font-semibold">{isRunning ? 'Running' : 'Stopped'}</span>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Uptime</CardTitle>
              </CardHeader>
              <CardContent className="flex items-center">
                <Clock className="h-4 w-4 mr-2 text-muted-foreground" />
                <span>{isRunning ? '12h 34m 56s' : 'N/A'}</span>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Last Deployed</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-sm">
                  <span>May 5, 2025 (1 day ago)</span>
                </div>
              </CardContent>
            </Card>
          </div>
          
          <div className="mb-6">
            <h2 className="text-lg font-semibold mb-3">Services</h2>
            <div className="grid grid-cols-3 gap-4">
              {['frontend', 'backend', 'database'].map(service => (
                <Card key={service} className="overflow-hidden">
                  <CardHeader className="bg-gray-50 pb-3">
                    <div className="flex justify-between">
                      <CardTitle className="text-sm font-medium">{service}</CardTitle>
                      <div className="flex items-center">
                        <div className="h-2 w-2 rounded-full mr-1 bg-green-500"></div>
                        <span className="text-xs">Running</span>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="pt-3">
                    <div className="space-y-1 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">CPU:</span>
                        <span>{metrics.cpu}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Memory:</span>
                        <span>{metrics.memory}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Storage:</span>
                        <span>{metrics.storage}</span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
          
          <div className={isLogExpanded ? "fixed inset-0 bg-white z-50 p-6" : ""}>
            <div className="flex justify-between items-center mb-3">
              <h2 className="text-lg font-semibold">
                <Terminal className="inline-block mr-2 h-5 w-5" />
                {selectedService ? `${selectedService} Logs` : "Stack Logs"}
              </h2>
              <Button 
                variant="ghost" 
                size="sm" 
                onClick={toggleLogExpanded}
              >
                {isLogExpanded ? <Minimize2 className="h-4 w-4" /> : <Maximize2 className="h-4 w-4" />}
              </Button>
            </div>
            
            <Tabs defaultValue="all">
              <TabsList>
                <TabsTrigger value="all">All Logs</TabsTrigger>
                <TabsTrigger value="error">Errors</TabsTrigger>
                <TabsTrigger value="info">Info</TabsTrigger>
              </TabsList>
              <TabsContent value="all" className="mt-2">
                <Card>
                  <CardContent className="p-0">
                    <div className={`font-mono text-xs ${isLogExpanded ? 'h-[calc(100vh-200px)]' : 'h-64'} overflow-auto bg-gray-900 text-gray-100 p-4`}>
                      {filteredLogs.map((log, index) => (
                        <div key={index} className="mb-1">
                          <span className="text-gray-400">[{log.time}]</span>{" "}
                          <span className="text-blue-400">[{log.service}]</span>{" "}
                          <span>{log.message}</span>
                        </div>
                      ))}
                      {/* Simulated cursor for the log terminal */}
                      <div className="h-4 w-2 bg-white animate-pulse inline-block"></div>
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>
              <TabsContent value="error" className="mt-2">
                <Card>
                  <CardContent className="p-4">
                    <div className="text-muted-foreground text-center py-8">
                      No errors found
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>
              <TabsContent value="info" className="mt-2">
                <Card>
                  <CardContent className="p-0">
                    <div className={`font-mono text-xs ${isLogExpanded ? 'h-[calc(100vh-200px)]' : 'h-64'} overflow-auto bg-gray-900 text-gray-100 p-4`}>
                      {filteredLogs.filter(log => !log.message.includes("error")).map((log, index) => (
                        <div key={index} className="mb-1">
                          <span className="text-gray-400">[{log.time}]</span>{" "}
                          <span className="text-blue-400">[{log.service}]</span>{" "}
                          <span>{log.message}</span>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>
            </Tabs>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
