import { useState } from "react";
import type { StackCompose, StackResource, Port } from "@/types/stack";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { 
  ChevronLeft, Database, Globe, Server, Eye, EyeOff, Box, 
  Link as LinkIcon, Play, HardDrive, PlusCircle, Trash2, Plus 
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";

interface StackConfigFormProps {
  stackConfig: StackCompose;
  onBack: () => void;
}

export function StackConfigForm({ stackConfig, onBack }: StackConfigFormProps) {
  // Create a mutable copy of the stack configuration
  const [config, setConfig] = useState<StackCompose>(() => {
    // Deep clone the stackConfig object
    return JSON.parse(JSON.stringify(stackConfig));
  });
  
  // Stack name and namespace
  const [stackName, setStackName] = useState("my-awesome-stack");
  const [namespace, setNamespace] = useState("default");

  // Filter out the volumes key to get only the service resources
  const resources = Object.entries(config)
    .filter(([key]) => key !== "volumes")
    .map(([key, value]) => {
      // Ensure that each resource is of type StackResource
      return [key, value as StackResource] as const;
    });

  // Function to get an icon based on resource type/name
  const getResourceIcon = (resourceName: string) => {
    const name = resourceName.toLowerCase();
    
    if (name.includes("db") || name.includes("postgres") || name.includes("mysql") || name.includes("mongo")) {
      return <Database className="h-5 w-5" />;
    } else if (name.includes("api") || name.includes("backend")) {
      return <Server className="h-5 w-5" />;
    } else if (name.includes("frontend") || name.includes("ui") || name.includes("web")) {
      return <Globe className="h-5 w-5" />;
    } else if (name.includes("redis") || name.includes("cache")) {
      return <Database className="h-5 w-5 text-red-500" />;
    } else if (name.includes("nginx") || name.includes("proxy")) {
      return <Box className="h-5 w-5" />;
    } else {
      return <Server className="h-5 w-5" />;
    }
  };

  // Update resource value
  const updateResource = (resourceName: string, key: string, value: any) => {
    setConfig(prev => {
      const newConfig = { ...prev };
      const resource = { ...(newConfig[resourceName] as StackResource) };
      resource[key as keyof StackResource] = value;
      newConfig[resourceName] = resource;
      return newConfig;
    });
  };

  // Update a port configuration
  const updatePort = (resourceName: string, index: number, key: keyof Port, value: any) => {
    setConfig(prev => {
      const newConfig = { ...prev };
      const resource = { ...(newConfig[resourceName] as StackResource) };
      if (resource.ports) {
        const ports = [...resource.ports];
        ports[index] = { ...ports[index], [key]: value };
        resource.ports = ports;
        newConfig[resourceName] = resource;
      }
      return newConfig;
    });
  };

  // Add a new port
  const addPort = (resourceName: string) => {
    setConfig(prev => {
      const newConfig = { ...prev };
      const resource = { ...(newConfig[resourceName] as StackResource) };
      const newPort: Port = {
        number: 8080,
        exposeToPublic: false,
        isHttp: true
      };
      resource.ports = resource.ports ? [...resource.ports, newPort] : [newPort];
      newConfig[resourceName] = resource;
      return newConfig;
    });
  };

  // Remove a port
  const removePort = (resourceName: string, index: number) => {
    setConfig(prev => {
      const newConfig = { ...prev };
      const resource = { ...(newConfig[resourceName] as StackResource) };
      if (resource.ports) {
        const ports = [...resource.ports];
        ports.splice(index, 1);
        resource.ports = ports;
        newConfig[resourceName] = resource;
      }
      return newConfig;
    });
  };

  // Add a new dependency
  const addDependency = (resourceName: string, dependency: string) => {
    if (!dependency.trim()) return;
    
    setConfig(prev => {
      const newConfig = { ...prev };
      const resource = { ...(newConfig[resourceName] as StackResource) };
      resource.dependsOn = resource.dependsOn ? 
        [...resource.dependsOn, dependency] : 
        [dependency];
      newConfig[resourceName] = resource;
      return newConfig;
    });
  };

  // Remove a dependency
  const removeDependency = (resourceName: string, dependency: string) => {
    setConfig(prev => {
      const newConfig = { ...prev };
      const resource = { ...(newConfig[resourceName] as StackResource) };
      if (resource.dependsOn) {
        resource.dependsOn = resource.dependsOn.filter(dep => dep !== dependency);
        newConfig[resourceName] = resource;
      }
      return newConfig;
    });
  };

  // Add environment variable
  const addEnvironmentVariable = (resourceName: string) => {
    setConfig(prev => {
      const newConfig = { ...prev };
      const resource = { ...(newConfig[resourceName] as StackResource) };
      resource.environmentVariables = resource.environmentVariables ?
        { ...resource.environmentVariables, "NEW_ENV_VAR": "value" } :
        { "NEW_ENV_VAR": "value" };
      newConfig[resourceName] = resource;
      return newConfig;
    });
  };

  // Update environment variable
  const updateEnvironmentVariable = (resourceName: string, oldKey: string, newKey: string, value: string) => {
    setConfig(prev => {
      const newConfig = { ...prev };
      const resource = { ...(newConfig[resourceName] as StackResource) };
      if (resource.environmentVariables) {
        const envVars = { ...resource.environmentVariables };
        if (oldKey !== newKey) {
          delete envVars[oldKey];
        }
        envVars[newKey] = value;
        resource.environmentVariables = envVars;
        newConfig[resourceName] = resource;
      }
      return newConfig;
    });
  };

  // Remove environment variable
  const removeEnvironmentVariable = (resourceName: string, key: string) => {
    setConfig(prev => {
      const newConfig = { ...prev };
      const resource = { ...(newConfig[resourceName] as StackResource) };
      if (resource.environmentVariables) {
        const envVars = { ...resource.environmentVariables };
        delete envVars[key];
        resource.environmentVariables = envVars;
        newConfig[resourceName] = resource;
      }
      return newConfig;
    });
  };

  // Add env file
  const addEnvFile = (resourceName: string) => {
    setConfig(prev => {
      const newConfig = { ...prev };
      const resource = { ...(newConfig[resourceName] as StackResource) };
      resource.envFiles = resource.envFiles ? [...resource.envFiles, ".env.local"] : [".env.local"];
      newConfig[resourceName] = resource;
      return newConfig;
    });
  };

  // Update env file
  const updateEnvFile = (resourceName: string, index: number, value: string) => {
    setConfig(prev => {
      const newConfig = { ...prev };
      const resource = { ...(newConfig[resourceName] as StackResource) };
      if (resource.envFiles) {
        const envFiles = [...resource.envFiles];
        envFiles[index] = value;
        resource.envFiles = envFiles;
        newConfig[resourceName] = resource;
      }
      return newConfig;
    });
  };

  // Remove env file
  const removeEnvFile = (resourceName: string, index: number) => {
    setConfig(prev => {
      const newConfig = { ...prev };
      const resource = { ...(newConfig[resourceName] as StackResource) };
      if (resource.envFiles) {
        const envFiles = [...resource.envFiles];
        envFiles.splice(index, 1);
        resource.envFiles = envFiles;
        newConfig[resourceName] = resource;
      }
      return newConfig;
    });
  };

  // Add volume mount
  const addVolumeMount = (resourceName: string) => {
    setConfig(prev => {
      const newConfig = { ...prev };
      const resource = { ...(newConfig[resourceName] as StackResource) };
      resource.volumeMounts = resource.volumeMounts ? 
        { ...resource.volumeMounts, "source": "/target" } : 
        { "source": "/target" };
      newConfig[resourceName] = resource;
      return newConfig;
    });
  };

  // Update volume mount
  const updateVolumeMount = (resourceName: string, oldSource: string, newSource: string, target: string) => {
    setConfig(prev => {
      const newConfig = { ...prev };
      const resource = { ...(newConfig[resourceName] as StackResource) };
      if (resource.volumeMounts) {
        const mounts = { ...resource.volumeMounts };
        if (oldSource !== newSource) {
          delete mounts[oldSource];
        }
        mounts[newSource] = target;
        resource.volumeMounts = mounts;
        newConfig[resourceName] = resource;
      }
      return newConfig;
    });
  };

  // Remove volume mount
  const removeVolumeMount = (resourceName: string, source: string) => {
    setConfig(prev => {
      const newConfig = { ...prev };
      const resource = { ...(newConfig[resourceName] as StackResource) };
      if (resource.volumeMounts) {
        const mounts = { ...resource.volumeMounts };
        delete mounts[source];
        resource.volumeMounts = mounts;
        newConfig[resourceName] = resource;
      }
      return newConfig;
    });
  };

  // Handle form submission
  const handleSubmit = () => {
    // Here you would send the updated config to your backend
    console.log("Submitting updated stack configuration:", {
      name: stackName,
      namespace,
      config
    });
    // For demo purposes, show an alert
    alert("Stack configuration updated");
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center">
        <Button variant="ghost" onClick={onBack} className="p-0 mr-2">
          <ChevronLeft className="h-5 w-5" />
        </Button>
        <h3 className="text-lg font-medium">Configure Stack Resources</h3>
      </div>

      <div className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label htmlFor="stack-name" className="text-sm font-medium mb-2 block">
              Stack Name
            </label>
            <Input 
              id="stack-name" 
              placeholder="my-awesome-stack" 
              value={stackName}
              onChange={(e) => setStackName(e.target.value)}
            />
          </div>
          <div>
            <label htmlFor="stack-namespace" className="text-sm font-medium mb-2 block">
              Namespace
            </label>
            <Input 
              id="stack-namespace" 
              placeholder="default" 
              value={namespace}
              onChange={(e) => setNamespace(e.target.value)}
            />
          </div>
        </div>
      </div>

      <Accordion type="multiple" className="w-full" defaultValue={resources.map(([key]) => key)}>
        {resources.map(([resourceName, resource]) => (
          <AccordionItem key={resourceName} value={resourceName}>
            <AccordionTrigger className="hover:bg-muted/50 px-4 rounded-md">
              <div className="flex items-center">
                {getResourceIcon(resourceName)}
                <span className="ml-2 font-medium">{resourceName}</span>
                <span className="ml-2 text-xs text-muted-foreground bg-muted px-2 py-1 rounded">
                  {resource.image || resource.imageRegistry || "Custom Build"}
                </span>
              </div>
            </AccordionTrigger>
            <AccordionContent>
              <div className="space-y-4 px-4 pb-2">
                {/* Basic Configuration */}
                <div className="grid grid-cols-2 gap-4">
                  {resource.image !== undefined && (
                    <div>
                      <label className="text-sm font-medium mb-1 block">Image</label>
                      <Input 
                        value={resource.image}
                        onChange={(e) => updateResource(resourceName, 'image', e.target.value)}
                      />
                    </div>
                  )}
                  {resource.imageRegistry !== undefined && (
                    <div>
                      <label className="text-sm font-medium mb-1 block">Image Registry</label>
                      <Input 
                        value={resource.imageRegistry}
                        onChange={(e) => updateResource(resourceName, 'imageRegistry', e.target.value)}
                      />
                    </div>
                  )}
                </div>

                {/* Build Configuration */}
                {resource.build && (
                  <Card>
                    <CardContent className="p-4 space-y-3">
                      <h4 className="text-sm font-semibold flex items-center">
                        <Box className="h-4 w-4 mr-2" />
                        Build Configuration
                      </h4>
                      <div className="grid grid-cols-2 gap-3">
                        <div>
                          <label className="text-xs font-medium mb-1 block">Source Volume</label>
                          <Input 
                            className="text-sm" 
                            value={resource.build.sourceVolume}
                            onChange={(e) => {
                              const updatedBuild = { ...resource.build, sourceVolume: e.target.value };
                              updateResource(resourceName, 'build', updatedBuild);
                            }}
                          />
                        </div>
                        <div>
                          <label className="text-xs font-medium mb-1 block">Build Context</label>
                          <Input 
                            className="text-sm" 
                            value={resource.build.buildContext}
                            onChange={(e) => {
                              const updatedBuild = { ...resource.build, buildContext: e.target.value };
                              updateResource(resourceName, 'build', updatedBuild);
                            }}
                          />
                        </div>
                        <div className="col-span-2">
                          <label className="text-xs font-medium mb-1 block">Dockerfile Path</label>
                          <Input 
                            className="text-sm" 
                            value={resource.build.dockerFilePath}
                            onChange={(e) => {
                              const updatedBuild = { ...resource.build, dockerFilePath: e.target.value };
                              updateResource(resourceName, 'build', updatedBuild);
                            }}
                          />
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                )}

                {/* Ports Configuration */}
                <Card>
                  <CardContent className="p-4">
                    <div className="flex justify-between items-center mb-3">
                      <h4 className="text-sm font-semibold flex items-center">
                        <Globe className="h-4 w-4 mr-2" />
                        Ports Configuration
                      </h4>
                      <Button 
                        variant="outline" 
                        size="sm" 
                        onClick={() => addPort(resourceName)}
                        className="h-7 px-2"
                      >
                        <Plus className="h-3.5 w-3.5 mr-1" />
                        Add Port
                      </Button>
                    </div>
                    
                    {resource.ports && resource.ports.length > 0 ? (
                      <div className="space-y-3">
                        {resource.ports.map((port, index) => (
                          <div key={index} className="grid grid-cols-12 gap-2 items-center border p-2 rounded-md">
                            <div className="col-span-3">
                              <label className="text-xs font-medium mb-1 block">Port Number</label>
                              <Input 
                                className="text-sm" 
                                type="number" 
                                value={port.number}
                                onChange={(e) => updatePort(
                                  resourceName, 
                                  index, 
                                  'number', 
                                  parseInt(e.target.value)
                                )}
                              />
                            </div>
                            <div className="col-span-4 space-y-1">
                              <label className="text-xs font-medium block">Public Access</label>
                              <div className="flex items-center">
                                <Switch 
                                  id={`port-public-${resourceName}-${index}`}
                                  checked={port.exposeToPublic}
                                  onCheckedChange={(checked) => updatePort(
                                    resourceName, 
                                    index, 
                                    'exposeToPublic', 
                                    checked
                                  )}
                                  className="mr-2"
                                />
                                <Label htmlFor={`port-public-${resourceName}-${index}`} className="text-xs">
                                  {port.exposeToPublic ? "Public" : "Private"}
                                </Label>
                              </div>
                            </div>
                            <div className="col-span-4 space-y-1">
                              <label className="text-xs font-medium block">Protocol</label>
                              <div className="flex items-center">
                                <Switch 
                                  id={`port-http-${resourceName}-${index}`}
                                  checked={port.isHttp}
                                  onCheckedChange={(checked) => updatePort(
                                    resourceName, 
                                    index, 
                                    'isHttp', 
                                    checked
                                  )}
                                  className="mr-2"
                                />
                                <Label htmlFor={`port-http-${resourceName}-${index}`} className="text-xs">
                                  {port.isHttp ? "HTTP" : "TCP/UDP"}
                                </Label>
                              </div>
                            </div>
                            <div className="col-span-1 flex justify-end items-center">
                              <Button 
                                variant="ghost" 
                                size="sm" 
                                onClick={() => removePort(resourceName, index)}
                                className="h-7 w-7 p-0"
                              >
                                <Trash2 className="h-4 w-4 text-red-500" />
                              </Button>
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="border border-dashed rounded-md p-4 text-center">
                        <p className="text-sm text-muted-foreground">No ports configured</p>
                      </div>
                    )}
                  </CardContent>
                </Card>

                {/* Dependencies */}
                <Card>
                  <CardContent className="p-4">
                    <div className="flex justify-between items-center mb-3">
                      <h4 className="text-sm font-semibold flex items-center">
                        <LinkIcon className="h-4 w-4 mr-2" />
                        Dependencies
                      </h4>
                      <div className="flex">
                        <Input
                          id={`new-dep-${resourceName}`}
                          placeholder="Add dependency..."
                          className="h-7 text-xs mr-2 w-40"
                          onKeyDown={(e) => {
                            if (e.key === 'Enter') {
                              const input = e.currentTarget;
                              addDependency(resourceName, input.value);
                              input.value = '';
                            }
                          }}
                        />
                        <Button 
                          variant="outline" 
                          size="sm" 
                          onClick={() => {
                            const input = document.getElementById(`new-dep-${resourceName}`) as HTMLInputElement;
                            addDependency(resourceName, input.value);
                            input.value = '';
                          }}
                          className="h-7 px-2"
                        >
                          <Plus className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    </div>
                    
                    <div className="flex flex-wrap gap-2">
                      {resource.dependsOn && resource.dependsOn.length > 0 ? (
                        resource.dependsOn.map((dep) => (
                          <div 
                            key={dep} 
                            className="bg-muted px-2 py-1 rounded-md text-xs flex items-center"
                          >
                            <span className="mr-1">{dep}</span>
                            <Button 
                              variant="ghost" 
                              size="sm" 
                              onClick={() => removeDependency(resourceName, dep)}
                              className="h-4 w-4 p-0"
                            >
                              <Trash2 className="h-3 w-3 text-red-500" />
                            </Button>
                          </div>
                        ))
                      ) : (
                        <div className="border border-dashed rounded-md p-3 w-full text-center">
                          <p className="text-sm text-muted-foreground">No dependencies configured</p>
                        </div>
                      )}
                    </div>
                  </CardContent>
                </Card>

                {/* Volume Mounts */}
                <Card>
                  <CardContent className="p-4">
                    <div className="flex justify-between items-center mb-3">
                      <h4 className="text-sm font-semibold flex items-center">
                        <HardDrive className="h-4 w-4 mr-2" />
                        Volume Mounts
                      </h4>
                      <Button 
                        variant="outline" 
                        size="sm" 
                        onClick={() => addVolumeMount(resourceName)}
                        className="h-7 px-2"
                      >
                        <Plus className="h-3.5 w-3.5 mr-1" />
                        Add Volume
                      </Button>
                    </div>
                    
                    {resource.volumeMounts && Object.keys(resource.volumeMounts).length > 0 ? (
                      <div className="space-y-2">
                        {Object.entries(resource.volumeMounts).map(([source, target], idx) => (
                          <div key={idx} className="grid grid-cols-9 gap-2 items-center border p-2 rounded-md">
                            <div className="col-span-4">
                              <label className="text-xs font-medium mb-1 block">Source</label>
                              <Input 
                                className="text-sm" 
                                value={source}
                                onChange={(e) => updateVolumeMount(
                                  resourceName, 
                                  source, 
                                  e.target.value, 
                                  target
                                )}
                              />
                            </div>
                            <div className="col-span-4">
                              <label className="text-xs font-medium mb-1 block">Target</label>
                              <Input 
                                className="text-sm" 
                                value={target}
                                onChange={(e) => updateVolumeMount(
                                  resourceName, 
                                  source, 
                                  source, 
                                  e.target.value
                                )}
                              />
                            </div>
                            <div className="col-span-1 flex justify-end items-end pb-1">
                              <Button 
                                variant="ghost" 
                                size="sm" 
                                onClick={() => removeVolumeMount(resourceName, source)}
                                className="h-7 w-7 p-0"
                              >
                                <Trash2 className="h-4 w-4 text-red-500" />
                              </Button>
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="border border-dashed rounded-md p-4 text-center">
                        <p className="text-sm text-muted-foreground">No volume mounts configured</p>
                      </div>
                    )}
                  </CardContent>
                </Card>

                {/* Environment Variables */}
                <Card>
                  <CardContent className="p-4">
                    <div className="flex justify-between items-center mb-3">
                      <h4 className="text-sm font-semibold flex items-center">
                        <Play className="h-4 w-4 mr-2" />
                        Environment Variables
                      </h4>
                      <Button 
                        variant="outline" 
                        size="sm" 
                        onClick={() => addEnvironmentVariable(resourceName)}
                        className="h-7 px-2"
                      >
                        <Plus className="h-3.5 w-3.5 mr-1" />
                        Add Env Var
                      </Button>
                    </div>
                    
                    {resource.environmentVariables && Object.keys(resource.environmentVariables).length > 0 ? (
                      <div className="space-y-2">
                        {Object.entries(resource.environmentVariables).map(([key, value], idx) => (
                          <div key={idx} className="grid grid-cols-9 gap-2 items-center border p-2 rounded-md">
                            <div className="col-span-4">
                              <label className="text-xs font-medium mb-1 block">Key</label>
                              <Input 
                                className="text-sm" 
                                value={key}
                                onChange={(e) => updateEnvironmentVariable(
                                  resourceName, 
                                  key, 
                                  e.target.value, 
                                  value
                                )}
                              />
                            </div>
                            <div className="col-span-4">
                              <label className="text-xs font-medium mb-1 block">Value</label>
                              <Input 
                                className="text-sm" 
                                value={value}
                                onChange={(e) => updateEnvironmentVariable(
                                  resourceName, 
                                  key, 
                                  key, 
                                  e.target.value
                                )}
                              />
                            </div>
                            <div className="col-span-1 flex justify-end items-end pb-1">
                              <Button 
                                variant="ghost" 
                                size="sm" 
                                onClick={() => removeEnvironmentVariable(resourceName, key)}
                                className="h-7 w-7 p-0"
                              >
                                <Trash2 className="h-4 w-4 text-red-500" />
                              </Button>
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="border border-dashed rounded-md p-4 text-center">
                        <p className="text-sm text-muted-foreground">No environment variables configured</p>
                      </div>
                    )}
                  </CardContent>
                </Card>

                {/* Environment Files */}
                <Card>
                  <CardContent className="p-4">
                    <div className="flex justify-between items-center mb-3">
                      <h4 className="text-sm font-semibold flex items-center">
                        <Play className="h-4 w-4 mr-2" />
                        Environment Files
                      </h4>
                      <Button 
                        variant="outline" 
                        size="sm" 
                        onClick={() => addEnvFile(resourceName)}
                        className="h-7 px-2"
                      >
                        <Plus className="h-3.5 w-3.5 mr-1" />
                        Add Env File
                      </Button>
                    </div>
                    
                    {resource.envFiles && resource.envFiles.length > 0 ? (
                      <div className="space-y-2">
                        {resource.envFiles.map((file, idx) => (
                          <div key={idx} className="flex items-center border p-2 rounded-md">
                            <Input 
                              className="text-sm" 
                              value={file}
                              onChange={(e) => updateEnvFile(resourceName, idx, e.target.value)}
                            />
                            <Button 
                              variant="ghost" 
                              size="sm" 
                              onClick={() => removeEnvFile(resourceName, idx)}
                              className="h-7 w-7 p-0 ml-2"
                            >
                              <Trash2 className="h-4 w-4 text-red-500" />
                            </Button>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="border border-dashed rounded-md p-4 text-center">
                        <p className="text-sm text-muted-foreground">No environment files configured</p>
                      </div>
                    )}
                  </CardContent>
                </Card>

                {/* Command and Args */}
                {(resource.command || resource.args) && (
                  <Card>
                    <CardContent className="p-4">
                      <h4 className="text-sm font-semibold flex items-center mb-3">
                        <Play className="h-4 w-4 mr-2" />
                        Command Configuration
                      </h4>
                      {resource.command && (
                        <div className="mb-3">
                          <label className="text-xs font-medium mb-1 block">Command</label>
                          <Input 
                            className="text-sm" 
                            value={resource.command.join(" ")}
                            onChange={(e) => {
                              const newCommand = e.target.value.split(" ").filter(Boolean);
                              updateResource(resourceName, 'command', newCommand);
                            }}
                          />
                        </div>
                      )}
                      {resource.args && (
                        <div>
                          <label className="text-xs font-medium mb-1 block">Arguments</label>
                          <Input 
                            className="text-sm" 
                            value={resource.args.join(" ")}
                            onChange={(e) => {
                              const newArgs = e.target.value.split(" ").filter(Boolean);
                              updateResource(resourceName, 'args', newArgs);
                            }}
                          />
                        </div>
                      )}
                    </CardContent>
                  </Card>
                )}
              </div>
            </AccordionContent>
          </AccordionItem>
        ))}
      </Accordion>

      {/* Volumes Section */}
      {config.volumes && (
        <Accordion type="single" collapsible className="w-full">
          <AccordionItem value="volumes">
            <AccordionTrigger className="hover:bg-muted/50 px-4 rounded-md">
              <div className="flex items-center">
                <HardDrive className="h-5 w-5" />
                <span className="ml-2 font-medium">Volumes</span>
              </div>
            </AccordionTrigger>
            <AccordionContent>
              <div className="space-y-4 px-4 pb-2">
                {Object.entries(config.volumes).map(([volumeName, volume]) => (
                  <Card key={volumeName}>
                    <CardContent className="p-4">
                      <h4 className="text-sm font-semibold flex items-center mb-3">
                        {volumeName}
                      </h4>
                      <div className="grid grid-cols-2 gap-3">
                        <div>
                          <label className="text-xs font-medium mb-1 block">Size</label>
                          <Input 
                            className="text-sm" 
                            value={volume.size}
                            onChange={(e) => {
                              const newVolumes = { ...config.volumes };
                              newVolumes[volumeName] = { 
                                ...volume, 
                                size: e.target.value 
                              };
                              setConfig({ ...config, volumes: newVolumes });
                            }}
                          />
                        </div>
                        {volume.source?.localDir && (
                          <>
                            <div>
                              <label className="text-xs font-medium mb-1 block">Local Path</label>
                              <Input 
                                className="text-sm" 
                                value={volume.source.localDir.path}
                                onChange={(e) => {
                                  const newVolumes = { ...config.volumes };
                                  newVolumes[volumeName] = { 
                                    ...volume, 
                                    source: { 
                                      ...volume.source, 
                                      localDir: {
                                        ...volume.source.localDir,
                                        path: e.target.value
                                      }
                                    }
                                  };
                                  setConfig({ ...config, volumes: newVolumes });
                                }}
                              />
                            </div>
                            <div className="col-span-2">
                              <div className="flex items-center">
                                <Switch 
                                  id={`volume-sync-${volumeName}`}
                                  checked={volume.source.localDir.sync}
                                  onCheckedChange={(checked) => {
                                    const newVolumes = { ...config.volumes };
                                    newVolumes[volumeName] = { 
                                      ...volume, 
                                      source: { 
                                        ...volume.source, 
                                        localDir: {
                                          ...volume.source.localDir,
                                          sync: checked
                                        }
                                      }
                                    };
                                    setConfig({ ...config, volumes: newVolumes });
                                  }}
                                  className="mr-2"
                                />
                                <Label htmlFor={`volume-sync-${volumeName}`}>
                                  Sync Enabled
                                </Label>
                              </div>
                            </div>
                          </>
                        )}
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      )}

      <div className="pt-6 flex justify-end space-x-2">
        <Button variant="outline" onClick={onBack}>
          Back to Input
        </Button>
        <Button type="button" onClick={handleSubmit}>
          Create Stack
        </Button>
      </div>
    </div>
  );
}
