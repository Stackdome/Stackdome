import { useState } from "react";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/use-toast";
import { useNavigate } from "react-router-dom";
import { useStacks } from "@/pages/stacks/contexts/stack-context";
import type { StackFormState } from "./types";
import BasicInfoSection from "./basic-info-section";
import SourceCodeSection from "./source-code-section";
import StackConfigSection from "./stack-config-section";
import EnvironmentSection from "./environment-section";
import ReviewDeploySection from "./review-deploy-section";

export default function StackCreatePage() {
  // Form state for all sections
  const [form, setForm] = useState<StackFormState>({
    name: "",
    description: "",
    region: "US East (N. Virginia)",
    template: "",
    repositoryUrl: "",
    yamlConfig: "",
    environment: {},
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [tab, setTab] = useState("basic-info");
  const { toast } = useToast();
  const navigate = useNavigate();
  const { addStack } = useStacks();

  // Handlers for each section
  const handleChange = (values: Partial<StackFormState>) => {
    setForm((prev) => ({ ...prev, ...values }));
  };

  const handleSubmit = async () => {
    setLoading(true);
    setError(null);
    try {
      addStack({
        name: form.name,
        description: form.description,
        template: form.template,
      });
      setLoading(false);
      toast({
        title: "Success",
        description: "Stack created successfully",
        variant: "success",
        duration: 3000,
      });
      navigate("/stacks");
    } catch (e: any) {
      setError(e?.message || "Failed to create stack");
      toast({
        title: "Failed to create stack",
        description: e?.message || "Failed to create stack. Please try again.",
        variant: "destructive",
        duration: 5000,
      });
      setLoading(false);
    }
  };

  return (
    <div className="p-4 pt-0 h-full">
      <div className="flex h-[calc(100vh-64px)]">
        <div className="flex-grow p-6 overflow-y-auto">
          <div className="max-w-2xl w-full mx-auto">
            <h2 className="text-2xl font-semibold mb-6">Create New Stack</h2>
            <Tabs value={tab} onValueChange={setTab} className="w-full">
              <TabsList className="mb-4">
                <TabsTrigger value="basic-info">Basic Info</TabsTrigger>
                <TabsTrigger value="source-code">Source Code</TabsTrigger>
                <TabsTrigger value="stack-config">Stack Configuration</TabsTrigger>
                <TabsTrigger value="environment">Environment</TabsTrigger>
                <TabsTrigger value="review">Review & Deploy</TabsTrigger>
              </TabsList>
              <TabsContent value="basic-info">
                <Card><CardContent><BasicInfoSection value={form} onChange={handleChange} error={error} loading={loading} /></CardContent></Card>
              </TabsContent>
              <TabsContent value="source-code">
                <Card><CardContent><SourceCodeSection value={form} onChange={handleChange} error={error} loading={loading} /></CardContent></Card>
              </TabsContent>
              <TabsContent value="stack-config">
                <Card><CardContent><StackConfigSection value={form} onChange={handleChange} error={error} loading={loading} /></CardContent></Card>
              </TabsContent>
              <TabsContent value="environment">
                <Card><CardContent><EnvironmentSection value={form} onChange={handleChange} error={error} loading={loading} /></CardContent></Card>
              </TabsContent>
              <TabsContent value="review">
                <Card><CardContent><ReviewDeploySection value={form} onSubmit={handleSubmit} loading={loading} error={error} /></CardContent></Card>
              </TabsContent>
            </Tabs>
          </div>
        </div>
      </div>
    </div>
  );
}
