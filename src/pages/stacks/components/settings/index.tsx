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
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Settings as SettingsIcon, Trash2, AlertTriangle, Save } from "lucide-react";
import { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Switch } from "@/components/ui/switch";

export default function StackSettingsPage() {
  const { id } = useParams();
  const { stacks, removeStack } = useStacks();

  // Find the current stack
  const currentStack = stacks.find(stack => stack.id === id);

  const [stackName, setStackName] = useState(currentStack?.name || "");
  const [stackDescription, setStackDescription] = useState(currentStack?.description || "");
  const [autoScaling, setAutoScaling] = useState(true);
  const [isPublic, setIsPublic] = useState(false);
  const [loggingEnabled, setLoggingEnabled] = useState(true);

  if (!currentStack) {
    return (
      <div className="p-8 text-center">
        <h2 className="text-xl font-semibold mb-2">Stack not found</h2>
        <p className="text-muted-foreground mb-4">The stack you're looking for doesn't exist or has been deleted.</p>
        <Link to="/stacks" className="text-primary hover:underline">Return to Stacks</Link>
      </div>
    );
  }

  const handleDeleteStack = () => {
    if (window.confirm(`Are you sure you want to delete the stack "${currentStack.name}"?`)) {
      removeStack(currentStack.id);
      // Navigate back to stacks page
      window.location.href = "/stacks";
    }
  };

  const handleSaveSettings = () => {
    // In a real app, this would update the stack in the backend
    alert("Settings saved successfully!");
  };

  return (
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
              <BreadcrumbLink>Settings</BreadcrumbLink>
            </BreadcrumbItem>
          </BreadcrumbList>
        </Breadcrumb>

        <div className="flex items-center">
          <SettingsIcon className="mr-2 h-5 w-5" />
          <h1 className="text-2xl font-bold">Stack Settings</h1>
        </div>
        <p className="text-muted-foreground mt-1">
          Configure and manage settings for {currentStack.name}
        </p>
        <Separator className="mt-4" />
      </header>

      <Tabs defaultValue="general">
        <TabsList className="mb-4">
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="scaling">Scaling</TabsTrigger>
          <TabsTrigger value="network">Network</TabsTrigger>
          <TabsTrigger value="advanced">Advanced</TabsTrigger>
          <TabsTrigger value="danger">Danger Zone</TabsTrigger>
        </TabsList>

        <TabsContent value="general">
          <Card>
            <CardHeader>
              <CardTitle>General Settings</CardTitle>
              <CardDescription>
                Basic configuration for your stack
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="stack-name">Stack Name</Label>
                <Input
                  id="stack-name"
                  value={stackName}
                  onChange={(e) => setStackName(e.target.value)}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="stack-description">Description</Label>
                <Textarea
                  id="stack-description"
                  value={stackDescription}
                  onChange={(e) => setStackDescription(e.target.value)}
                  rows={3}
                />
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="public-stack">Public Stack</Label>
                  <Switch
                    id="public-stack"
                    checked={isPublic}
                    onCheckedChange={setIsPublic}
                  />
                </div>
                <p className="text-sm text-muted-foreground">
                  When enabled, this stack will be accessible without authentication.
                </p>
              </div>

              <Button className="mt-4" onClick={handleSaveSettings}>
                <Save className="h-4 w-4 mr-2" />
                Save Changes
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="scaling">
          <Card>
            <CardHeader>
              <CardTitle>Scaling Configuration</CardTitle>
              <CardDescription>
                Configure how your stack scales with traffic
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="auto-scaling">Auto-scaling</Label>
                  <Switch
                    id="auto-scaling"
                    checked={autoScaling}
                    onCheckedChange={setAutoScaling}
                  />
                </div>
                <p className="text-sm text-muted-foreground">
                  Automatically scale resources based on traffic and usage.
                </p>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="min-replicas">Minimum Replicas</Label>
                  <Input
                    id="min-replicas"
                    type="number"
                    defaultValue="1"
                    disabled={!autoScaling}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="max-replicas">Maximum Replicas</Label>
                  <Input
                    id="max-replicas"
                    type="number"
                    defaultValue="5"
                    disabled={!autoScaling}
                  />
                </div>
              </div>

              <Button className="mt-4" onClick={handleSaveSettings}>
                <Save className="h-4 w-4 mr-2" />
                Save Changes
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="network">
          <Card>
            <CardHeader>
              <CardTitle>Network Settings</CardTitle>
              <CardDescription>
                Configure network-related settings for your stack
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="custom-domain">Custom Domain</Label>
                <Input
                  id="custom-domain"
                  placeholder="example.com"
                />
                <p className="text-sm text-muted-foreground">
                  Enter a custom domain to use for your stack.
                </p>
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="ssl-enabled">SSL/TLS</Label>
                  <Switch
                    id="ssl-enabled"
                    defaultChecked={true}
                  />
                </div>
                <p className="text-sm text-muted-foreground">
                  Enable SSL/TLS encryption for your stack.
                </p>
              </div>

              <Button className="mt-4" onClick={handleSaveSettings}>
                <Save className="h-4 w-4 mr-2" />
                Save Changes
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="advanced">
          <Card>
            <CardHeader>
              <CardTitle>Advanced Settings</CardTitle>
              <CardDescription>
                Advanced configuration options for experienced users
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="logging-enabled">Logging</Label>
                  <Switch
                    id="logging-enabled"
                    checked={loggingEnabled}
                    onCheckedChange={setLoggingEnabled}
                  />
                </div>
                <p className="text-sm text-muted-foreground">
                  Enable detailed logging for this stack.
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="environment-variables">Environment Variables</Label>
                <Textarea
                  id="environment-variables"
                  placeholder="KEY=value
ANOTHER_KEY=another value"
                  rows={4}
                  className="font-mono text-sm"
                />
              </div>

              <Button className="mt-4" onClick={handleSaveSettings}>
                <Save className="h-4 w-4 mr-2" />
                Save Changes
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="danger">
          <Card className="border-red-200">
            <CardHeader className="text-red-600">
              <CardTitle className="flex items-center">
                <AlertTriangle className="h-5 w-5 mr-2" />
                Danger Zone
              </CardTitle>
              <CardDescription className="text-red-500">
                Destructive actions that cannot be undone
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="border border-red-200 rounded-md p-4 bg-red-50">
                <h3 className="font-medium">Delete Stack</h3>
                <p className="text-sm text-muted-foreground mb-4">
                  This will permanently delete the stack, all its services, and associated data.
                  This action cannot be undone.
                </p>
                <Button variant="destructive" onClick={handleDeleteStack}>
                  <Trash2 className="h-4 w-4 mr-2" />
                  Delete this Stack
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
