import { Layers, PlusCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useStacks } from "@/pages/stacks/contexts/stack-context";

export default function StacksPage() {
  const { stacks, removeStack } = useStacks();

  return (
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
            <Button>
              <PlusCircle className="mr-2 h-4 w-4" />
                Create New Stack
            </Button>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">

          <div className="flex items-center justify-center border-2 border-dashed rounded-lg p-8 h-full">
            <Button variant="outline" className="w-full h-full py-8 flex flex-col items-center">
              <PlusCircle className="h-8 w-8 mb-2" />
              <span>Add New Stack</span>
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
