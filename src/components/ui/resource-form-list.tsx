import { useState, useEffect, useRef, useCallback } from "react";
import type { ReactNode } from "react";
import { Accordion } from "@/components/ui/accordion";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";

// Generic form list props that can be used for different types of forms
interface ResourceFormListProps<T> {
  items: Partial<T>[];
  onItemsChange: (updatedItems: Partial<T>[]) => void;
  errors: { [index: number]: { [field: string]: string | undefined } };
  createDefaultItem: () => Partial<T>;
  renderItem: (props: {
    item: Partial<T>;
    index: number;
    itemRef: (el: HTMLButtonElement | null) => void;
    isOnlyItem: boolean;
    onChange: (index: number, updatedItem: Partial<T>) => void;
    onRemove: (index: number) => void;
    errors: { [field: string]: string | undefined };
  }) => ReactNode;
  addButtonText?: string;
}

export default function ResourceFormList<T>({
  items,
  onItemsChange,
  errors,
  createDefaultItem,
  renderItem,
  addButtonText = "Add Item"
}: ResourceFormListProps<T>) {
  const [openAccordions, setOpenAccordions] = useState<string[]>(["0"]);
  const [lastAddedIndex, setLastAddedIndex] = useState<number | null>(null);
  const itemRefs = useRef<Array<HTMLButtonElement | null>>([]);
  
  // Initialize with a default item if none provided
  useEffect(() => {
    if (items.length === 0) {
      onItemsChange([createDefaultItem()]);
    }
  }, [items.length, onItemsChange, createDefaultItem]);

  useEffect(() => {
    if (lastAddedIndex !== null && itemRefs.current[lastAddedIndex]) {
      itemRefs.current[lastAddedIndex]?.scrollIntoView({
        behavior: "smooth",
        block: "center",
      });
      setLastAddedIndex(null);
    }
  }, [items, lastAddedIndex]);

  // This useEffect only manages newly added items and cleanup
  // It doesn't interfere with user-closed accordions
  useEffect(() => {
    if (items.length > 0) {
      const newOpenAccordionsSet = new Set(openAccordions);
      let updated = false;

      // Clean up accordion entries for items that no longer exist
      const itemIndices = new Set(items.map((_, idx) => String(idx)));
      for (const openId of newOpenAccordionsSet) {
        if (!itemIndices.has(openId)) {
          newOpenAccordionsSet.delete(openId);
          updated = true;
        }
      }
      
      // Apply updates if needed
      if (updated) {
        setOpenAccordions(Array.from(newOpenAccordionsSet));
      }
    }
  }, [items, openAccordions]);
  
  // Separate useEffect that only runs when lastAddedIndex changes
  // This ensures newly added items are opened without affecting user interactions
  useEffect(() => {
    if (lastAddedIndex !== null) {
      setOpenAccordions(prev => {
        const newOpen = [...prev];
        if (!newOpen.includes(String(lastAddedIndex))) {
          newOpen.push(String(lastAddedIndex));
        }
        return newOpen;
      });
    }
  }, [lastAddedIndex]);
  
  // Add a new item to the list
  const handleAddItem = useCallback(() => {
    const newItemIndex = items.length;
    onItemsChange([...items, createDefaultItem()]);
    setLastAddedIndex(newItemIndex);
    
    // We dynamically allocate refs as needed
    if (itemRefs.current.length <= newItemIndex) {
      itemRefs.current = [...itemRefs.current, null];
    }
  }, [items, onItemsChange, createDefaultItem]);
  
  // Update an existing item
  const handleItemChange = useCallback((index: number, updatedItem: Partial<T>) => {
    onItemsChange(
      items.map((item, i) => i === index ? updatedItem : item)
    );
  }, [items, onItemsChange]);
  
  // Remove an item
  const handleRemoveItem = useCallback((index: number) => {
    onItemsChange(items.filter((_, i) => i !== index));
  }, [items, onItemsChange]);
  
  // Track state of accordions
  const handleValueChange = useCallback((value: string[]) => {
    setOpenAccordions(value);
  }, []);

  // Set up refs for each item
  const getItemRef = useCallback((index: number) => {
    return (el: HTMLButtonElement | null) => {
      itemRefs.current[index] = el;
    };
  }, []);

  return (
    <div className="rounded-lg overflow-hidden">
      <Accordion
        type="multiple"
        value={openAccordions}
        onValueChange={handleValueChange}
        className="rounded-none divide-y"
      >
        {items.map((item, index) => (
          renderItem({
            item,
            index,
            itemRef: getItemRef(index),
            isOnlyItem: items.length === 1,
            onChange: handleItemChange,
            onRemove: handleRemoveItem,
            errors: errors[index] || {},
          })
        ))}
      </Accordion>
      <div className="flex justify-center py-4 border-t">
        <Button
          variant="ghost"
          className="flex items-center gap-1"
          onClick={handleAddItem}
          type="button"
        >
          <Plus className="h-4 w-4" />
          <span>{addButtonText}</span>
        </Button>
      </div>
    </div>
  );
}
