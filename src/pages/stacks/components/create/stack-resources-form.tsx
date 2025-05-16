import { useState, useEffect, useRef, useCallback } from "react";
import { Accordion } from "@/components/ui/accordion";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import StackResourceItem from "./stack-resource-item";
import type { StackResourceData } from "@/pages/stacks/schemas/stack-create-schema";

// Props interface for StackResourcesForm
interface StackResourcesFormProps {
  resources: Partial<StackResourceData>[];
  onResourcesChange: (updatedResources: Partial<StackResourceData>[]) => void;
  errors: { [index: number]: { [field: string]: string | undefined } };
}

// Helper to create a default new resource, aligning with StackResourceData
function getDefaultResource(): Partial<StackResourceData> {
  return {
    name: "",
    sourceType: "image",
    labels: [],
    depends_on: [],
    ports: [],
    execution_config: { environment_variables: [] },
    build_spec: undefined,
    image_spec: { image: "" },
  };
}

export default function StackResourcesForm({
  resources,
  onResourcesChange,
  errors,
}: StackResourcesFormProps) {
  const [openAccordions, setOpenAccordions] = useState<string[]>([]);
  const [lastAddedIndex, setLastAddedIndex] = useState<number | null>(null);
  const itemRefs = useRef<Array<HTMLButtonElement | null>>([]);

  useEffect(() => {
    if (lastAddedIndex !== null && itemRefs.current[lastAddedIndex]) {
      itemRefs.current[lastAddedIndex]?.scrollIntoView({
        behavior: "smooth",
        block: "center",
      });
      setLastAddedIndex(null);
    }
  }, [resources, lastAddedIndex]);

  // This useEffect only manages newly added resources and cleanup
  // It doesn't interfere with user-closed accordions
  useEffect(() => {
    if (resources.length > 0) {
      const newOpenAccordionsSet = new Set(openAccordions);
      let updated = false;

      // Clean up accordion entries for resources that no longer exist
      const resourceIndices = new Set(resources.map((_, idx) => String(idx)));
      for (const openId of newOpenAccordionsSet) {
        if (!resourceIndices.has(openId)) {
          newOpenAccordionsSet.delete(openId);
          updated = true;
        }
      }
      
      // Apply updates if needed
      if (updated) {
        setOpenAccordions(Array.from(newOpenAccordionsSet));
      }
    }
  }, [resources, openAccordions]);
  
  // Separate useEffect that only runs when lastAddedIndex changes
  // This ensures newly added items are opened without affecting user interactions
  useEffect(() => {
    if (lastAddedIndex !== null) {
      const lastAddedIndexStr = String(lastAddedIndex);
      
      // Only add this item to open accordions if it's not already there
      if (!openAccordions.includes(lastAddedIndexStr)) {
        setOpenAccordions(prev => [...prev, lastAddedIndexStr]);
      }
    }
  }, [lastAddedIndex, openAccordions]);

  const handleResourceChange = useCallback(
    (index: number, updatedResourceData: Partial<StackResourceData>) => {
      const newResources = [...resources];
      newResources[index] = updatedResourceData;
      onResourcesChange(newResources);
    },
    [resources, onResourcesChange]
  );

  const handleRemoveResource = useCallback(
    (indexToRemove: number) => {
      const newResources = resources.filter((_, i) => i !== indexToRemove);
      onResourcesChange(newResources);
      
      // Remove the specific accordion from the open state
      setOpenAccordions((prev) => prev.filter((id) => id !== String(indexToRemove)));
      
      // Update lastAddedIndex if needed
      if (lastAddedIndex === indexToRemove) {
        setLastAddedIndex(null);
      }
      
      // Reset the refs array to match the new resource count
      if (itemRefs.current.length !== newResources.length) {
        const newRefs = Array(newResources.length).fill(null);
        
        // Copy existing refs, adjusting for the removed resource
        for (let i = 0, j = 0; i < itemRefs.current.length; i++) {
          if (i !== indexToRemove) {
            newRefs[j++] = itemRefs.current[i];
          }
        }
        
        itemRefs.current = newRefs;
      }
    },
    [resources, onResourcesChange, lastAddedIndex]
  );

  const handleAddResource = useCallback(() => {
    const newResource = getDefaultResource();
    const newResourceIndex = resources.length;
    
    // Add the new resource
    onResourcesChange([...resources, newResource]);
    
    // Set the lastAddedIndex which will trigger our useEffect to open the accordion
    setLastAddedIndex(newResourceIndex);
  }, [resources, onResourcesChange]);

  return (
    <div className="flex flex-col h-full">
      <div className="overflow-y-auto flex-1 scrollbar-hide px-6 pt-6 pb-6">
        <Accordion
          type="multiple"
          value={openAccordions}
          onValueChange={setOpenAccordions}
          className="w-full space-y-4"
        >
          {resources.length === 0 && (
            <div className="flex flex-col items-center justify-center py-16 border-2 border-dashed rounded-lg border-muted">
              <div className="mb-3 p-3 rounded-full bg-muted">
                <Plus className="h-6 w-6 text-muted-foreground" />
              </div>
              <p className="text-lg text-muted-foreground">
                No resources defined yet
              </p>
              <p className="text-sm text-muted-foreground mb-4">
                Add your first stack resource to get started
              </p>
              <Button onClick={handleAddResource}>
                <Plus className="w-4 h-4 mr-2" />
                Add First Resource
              </Button>
            </div>
          )}
          {resources.map((resource, idx) => (
            <StackResourceItem
              key={idx}
              resource={resource}
              index={idx}
              itemRef={(el: HTMLButtonElement | null) =>
                (itemRefs.current[idx] = el)
              }
              isOnlyResource={resources.length === 1}
              onChange={handleResourceChange}
              onRemove={handleRemoveResource}
              errors={errors[idx] || {}}
            />
          ))}
        </Accordion>
      </div>

      {resources.length > 0 && (
        <div className="sticky bottom-0 left-0 w-full bg-background/95 border-t py-6 px-6 backdrop-blur-sm shadow-md">
          <Button
            variant="outline"
            size="lg"
            className="border-dashed border-2 bg-background hover:bg-muted/20 text-foreground w-full max-w-md mx-auto flex items-center justify-center"
            onClick={handleAddResource}
          >
            <Plus className="w-4 h-4 mr-2" />
            Add Another Resource
          </Button>
        </div>
      )}
    </div>
  );
}
