import { useState, useEffect, useRef, useCallback } from "react";
import type { ReactNode } from "react";
import { Accordion } from "@/components/ui/accordion";
import { Container } from "lucide-react";

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
  autoAddFirstItem?: boolean;
  emptyText?: string;
  emptyIcon?: ReactNode;
  readOnly?: boolean;
}

export default function ResourceFormList<T>({
  items,
  onItemsChange,
  errors,
  createDefaultItem,
  renderItem,
  autoAddFirstItem = false, // default to false for blank state
  emptyText = "No Resources added.",
  emptyIcon,
}: ResourceFormListProps<T>) {
  const [openAccordions, setOpenAccordions] = useState<string[]>(["0"]);
  const [lastAddedIndex, setLastAddedIndex] = useState<number | null>(null);
  const itemRefs = useRef<Array<HTMLButtonElement | null>>([]);

  // Initialize with a default item if none provided
  useEffect(() => {
    if (autoAddFirstItem && items.length === 0) {
      onItemsChange([createDefaultItem()]);
    }
  }, [items.length, onItemsChange, createDefaultItem, autoAddFirstItem]);

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
    <div>
      {items.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12">
          {emptyIcon !== undefined ? (
            <>{emptyIcon}</>
          ) : (
            <Container className="mx-auto h-8 w-8 mb-2 text-muted-foreground" />
          )}
          <div className="text-lg text-muted-foreground font-medium mb-1">{emptyText}</div>
          {/* For volumes, show (Optional) below the text, centered */}
          {emptyText.toLowerCase().includes("volume") && (
            <div className="text-sm text-muted-foreground mt-1">(Optional)</div>
          )}
        </div>
      ) : (
        <div className="rounded-lg overflow-hidden">
          <Accordion
            type="multiple"
            value={openAccordions}
            onValueChange={handleValueChange}
            className="rounded-none divide-y"
          >
            {items.map((item, index) => (
              <div key={index} className="relative">
                {renderItem({
                  item,
                  index,
                  itemRef: getItemRef(index),
                  isOnlyItem: items.length === 1,
                  onChange: handleItemChange,
                  onRemove: handleRemoveItem,
                  errors: errors[index] || {},
                })}
              </div>
            ))}
          </Accordion>
        </div>
      )}

    </div>
  );
}
