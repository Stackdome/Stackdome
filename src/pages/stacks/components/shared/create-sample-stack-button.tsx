import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { useStacks } from "@/pages/stacks/contexts/stack-context";

export function CreateSampleStackButton() {
  const { addStack } = useStacks();
  
  const createSampleStack = () => {
    // Create a sample stack with basic information
    addStack({
      name: "Sample MERN Stack",
      description: "MongoDB, Express, React and Node.js stack for web applications",
      template: "Express",
    });
  };
  
  return (
    <Button onClick={createSampleStack} variant="secondary" className="w-full mt-4">
      <Plus className="mr-2 h-4 w-4" />
      Create Sample Stack
    </Button>
  );
}
