import { useState } from "react";
import { useNavigate } from "react-router-dom";
import StackResourcesForm from "./stack-resources-form";
import { Button } from "@/components/ui/button";
import { Rocket, Tag as TagIcon, X } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

// Types from OpenAPI
import type { components } from "@/api/types/openapi";

type Label = components["schemas"]["Label"];

export default function StackCreatePage() {
  const [stackName, setStackName] = useState("");
  const [labels, setLabels] = useState<Label[]>([]);
  const [currentLabel, setCurrentLabel] = useState("");
  const navigate = useNavigate();
  
  const handleAddLabel = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && currentLabel.trim()) {
      // Simple key-value parser (format: key=value or just key)
      const labelParts = currentLabel.trim().split("=");
      const key = labelParts[0].trim();
      const value = labelParts.length > 1 ? labelParts[1].trim() : "true";
      
      if (key) {
        setLabels(prev => [...prev, { key, value }]);
        setCurrentLabel("");
      }
      e.preventDefault();
    }
  };
  
  const removeLabel = (indexToRemove: number) => {
    setLabels(prev => prev.filter((_, idx) => idx !== indexToRemove));
  };

  return (
    <div className="h-full overflow-hidden flex flex-col">
      <div className="flex-shrink-0 px-4">
        {/* Header with navigation */}
        <div className="flex items-center justify-between py-6">
          <div>
            <h2 className="text-3xl font-bold tracking-tight">Create New Stack</h2>
            <p className="text-muted-foreground mt-1">
              Define your stack resources to provision infrastructure
            </p>
          </div>
          <div className="flex items-center gap-3">
            <Button variant="outline" onClick={() => {
              if (window.history.length > 2) {
                navigate(-1);
              } else {
                navigate("/stacks");
              }
            }}>Cancel</Button>
            <Button variant="default">
              <Rocket className="mr-2 h-4 w-4" />
              Deploy
            </Button>
          </div>
        </div>
        <Separator className="mb-6" />
      </div>
        
      {/* Scrollable content area */}
      <div className="flex-grow overflow-y-auto scrollbar-hide px-4 pb-10">
        <div className="flex flex-col">
          {/* === Stack Name & Labels Section === */}
          <Card className="mb-6 rounded-lg overflow-hidden">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="text-xl">Stack Information</CardTitle>
                </div>
              </div>
            </CardHeader>
            <Separator />
            <CardContent className="pt-6">
              <div className="grid gap-6 max-w-5xl">
                <div>
                  <Label htmlFor="stack-name" className="text-sm font-medium flex items-center gap-1 mb-2">
                    Stack Name <span className="text-red-500">*</span>
                    <Tooltip delayDuration={300}>
                      <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                      <TooltipContent className="max-w-xs" side="right">
                        <p>A unique name to identify this stack.</p>
                      </TooltipContent>
                    </Tooltip>
                  </Label>
                  <Input
                    id="stack-name"
                    value={stackName}
                    onChange={(e) => setStackName(e.target.value)}
                    className="max-w-md"
                    placeholder="my-application-stack"
                    required
                  />
                </div>
                
                <div>
                  <Label htmlFor="stack-labels" className="text-sm font-medium flex items-center gap-1 mb-2">
                    Labels
                    <Tooltip delayDuration={300}>
                      <TooltipTrigger tabIndex={-1} className="cursor-help rounded-full bg-muted px-1 text-xs text-muted-foreground">?</TooltipTrigger>
                      <TooltipContent className="max-w-xs" side="right">
                        <p>Add metadata to your stack using key-value labels. Format: key=value or just a tag (press Enter to add)</p>
                      </TooltipContent>
                    </Tooltip>
                  </Label>
                  <div className="flex items-center">
                    <TagIcon className="h-4 w-4 mr-2 text-muted-foreground" />
                    <Input
                      id="stack-labels"
                      value={currentLabel}
                      onChange={(e) => setCurrentLabel(e.target.value)}
                      onKeyDown={handleAddLabel}
                      className="max-w-md"
                      placeholder="e.g., environment=dev or just tag (press Enter to add)"
                    />
                  </div>
                  {labels.length > 0 && (
                    <div className="flex flex-wrap gap-2 mt-3">
                      {labels.map((label, idx) => (
                        <Badge 
                          key={idx} 
                          variant="secondary"
                          className="flex items-center gap-1 px-2.5 py-1"
                        >
                          <span>{label.key}{label.value !== "true" ? `=${label.value}` : ""}</span>
                          <button 
                            onClick={() => removeLabel(idx)} 
                            className="ml-1 rounded-full hover:bg-secondary-foreground/20 h-4 w-4 flex items-center justify-center"
                            type="button"
                          >
                            <X className="h-3 w-3" />
                          </button>
                        </Badge>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>
          {/* Section Card: Stack Resources */}
          <Card className="mb-6 rounded-lg overflow-hidden">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="text-xl">Define Stack Resources</CardTitle>
                  <CardDescription className="mt-1">
                    Configure the containerized services that make up your stack
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <Separator />
            <CardContent className="p-0" style={{ height: "500px" }}>
              <StackResourcesForm />
            </CardContent>
          </Card>
          {/* === End Stack Resources Section === */}
        </div>
      </div>
    </div>
  );
}
