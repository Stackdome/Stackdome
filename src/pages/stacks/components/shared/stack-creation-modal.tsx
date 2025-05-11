import { useState } from "react";
import type { StackCompose } from "@/types/stack";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { GitHubLogoIcon } from "@radix-ui/react-icons";
import { FileIcon, FileCode2 } from "lucide-react";
import { parseYaml, getSampleStackYaml } from "@/lib/yaml-parser";
import { StackConfigForm } from "./stack-config-form";
import { useStacks } from "@/pages/stacks/contexts/stack-context";

interface StackCreationModalProps {
  trigger: React.ReactNode;
}

export function StackCreationModal({ trigger }: StackCreationModalProps) {
  const [yamlContent, setYamlContent] = useState("");
  const [githubUrl, setGithubUrl] = useState("");
  const [stackConfig, setStackConfig] = useState<StackCompose | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoadingSample, setIsLoadingSample] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const { addStack } = useStacks();

  // Function to handle parsing YAML content
  const handleYamlParse = () => {
    try {
      const parsedYaml = parseYaml(yamlContent);
      setStackConfig(parsedYaml);
      setError(null);
    } catch (err) {
      setError("Invalid YAML format. Please check your input.");
      setStackConfig(null);
    }
  };

  // Function to handle GitHub repo connection
  const handleGithubConnect = () => {
    // This is a placeholder for the GitHub repo connection logic
    // In a real implementation, we would fetch the stack-compose.yaml from the repo
    setError("GitHub integration is not implemented yet");
  };

  // Load sample stack-compose.yaml
  const loadSampleStack = () => {
    try {
      setIsLoadingSample(true);
      const sampleYaml = getSampleStackYaml();
      setYamlContent(sampleYaml);
      setIsLoadingSample(false);
    } catch (err) {
      setError("Failed to load sample stack configuration");
      setIsLoadingSample(false);
    }
  };

  // Handle stack creation from the form
  const handleStackCreation = (name: string, description?: string, template?: string) => {
    addStack({
      name,
      description,
      template,
    });
    
    setIsOpen(false);
  };

  // Reset the form when dialog is closed
  const handleDialogChange = (open: boolean) => {
    setIsOpen(open);
    
    if (!open) {
      setYamlContent("");
      setGithubUrl("");
      setStackConfig(null);
      setError(null);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleDialogChange}>
      <DialogTrigger asChild onClick={() => setIsOpen(true)}>
        {trigger}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create New Stack</DialogTitle>
          <DialogDescription>
            Add a new stack by connecting to a GitHub repository or pasting your stack-compose.yaml content.
          </DialogDescription>
        </DialogHeader>

        {!stackConfig ? (
          <Tabs defaultValue="yaml" className="w-full">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="github">
                <GitHubLogoIcon className="mr-2" />
                GitHub Repository
              </TabsTrigger>
              <TabsTrigger value="yaml">
                <FileIcon className="mr-2 h-4 w-4" />
                YAML Content
              </TabsTrigger>
            </TabsList>
            
            <TabsContent value="github" className="p-4 border rounded-md mt-4">
              <div className="space-y-4">
                <div>
                  <label htmlFor="github-url" className="text-sm font-medium mb-2 block">
                    GitHub Repository URL
                  </label>
                  <Input
                    id="github-url"
                    placeholder="https://github.com/username/repo"
                    value={githubUrl}
                    onChange={(e) => setGithubUrl(e.target.value)}
                  />
                  <p className="text-xs text-muted-foreground mt-2">
                    The repository should contain a stack-compose.yaml file at the root.
                  </p>
                </div>
                <Button onClick={handleGithubConnect}>Connect to GitHub</Button>
              </div>
            </TabsContent>
            
            <TabsContent value="yaml" className="p-4 border rounded-md mt-4">
              <div className="space-y-4">
                <div>
                  <label htmlFor="yaml-content" className="text-sm font-medium mb-2 flex justify-between items-center">
                    <span>Stack Compose YAML</span>
                    <Button 
                      variant="outline" 
                      size="sm" 
                      onClick={loadSampleStack}
                      disabled={isLoadingSample}
                    >
                      <FileCode2 className="mr-2 h-4 w-4" />
                      Load Sample
                    </Button>
                  </label>
                  <Textarea
                    id="yaml-content"
                    placeholder="Paste your stack-compose.yaml content here..."
                    className="h-[250px] font-mono"
                    value={yamlContent}
                    onChange={(e) => setYamlContent(e.target.value)}
                  />
                </div>
                <Button onClick={handleYamlParse}>Parse YAML</Button>
              </div>
            </TabsContent>
          </Tabs>
        ) : (
          <StackConfigForm 
            stackConfig={stackConfig} 
            onBack={() => setStackConfig(null)} 
            onSubmit={handleStackCreation}
          />
        )}

        {error && (
          <div className="text-red-500 text-sm mt-2 p-2 bg-red-50 rounded-md">
            {error}
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
