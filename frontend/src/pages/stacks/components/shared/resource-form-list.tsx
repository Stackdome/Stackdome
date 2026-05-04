import { useState, useEffect, useRef, useCallback } from "react";
import type { ReactNode } from "react";
import { Accordion } from "@/components/ui/accordion";
import { Container } from "lucide-react";
import { Button } from "@/components/ui/button";

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
  emptyTitle?: string;
  emptyOptional?: boolean;
  emptyDescription?: string;
  emptyCtaLabel?: string;
  emptyOnAdd?: () => void;
  emptyIcon?: ReactNode;
  defaultAllCollapsed?: boolean; // If true, all accordions start closed
  defaultOpenIndex?: number | null; // If set (and not null), open this index by default instead of [0]
}

export default function ResourceFormList<T>({
  items,
  onItemsChange,
  errors,
  createDefaultItem,
  renderItem,
  autoAddFirstItem = false, // default to false for blank state
  emptyTitle = "Nothing added yet",
  emptyOptional = false,
  emptyDescription,
  emptyCtaLabel,
  emptyOnAdd,
  emptyIcon,
  defaultAllCollapsed = false,
  defaultOpenIndex = null,
}: ResourceFormListProps<T>) {
  // Set initial open accordions based on defaultAllCollapsed / defaultOpenIndex
  const [openAccordions, setOpenAccordions] = useState<string[]>(() => {
    if (defaultOpenIndex !== null && defaultOpenIndex !== undefined) {
      return [String(defaultOpenIndex)];
    }
    return defaultAllCollapsed ? [] : ["0"];
  });
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

  // Stable refs to the latest items + onItemsChange. Without this the
  // change/remove handlers below would be recreated on every keystroke
  // (because items changes), which defeats React.memo on each item child
  // and forces every other resource to re-render.
  const itemsRef = useRef(items);
  itemsRef.current = items;
  const onItemsChangeRef = useRef(onItemsChange);
  onItemsChangeRef.current = onItemsChange;

  const handleItemChange = useCallback((index: number, updatedItem: Partial<T>) => {
    onItemsChangeRef.current(
      itemsRef.current.map((item, i) => i === index ? updatedItem : item)
    );
  }, []);

  const handleRemoveItem = useCallback((index: number) => {
    onItemsChangeRef.current(itemsRef.current.filter((_, i) => i !== index));
  }, []);

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

  // Stable empty-errors reference — passing `{}` inline would break downstream memo.
  const EMPTY_ERRORS = useRef({}).current;

  return (
    <div>
      {items.length === 0 ? (
        <div className="flex flex-col items-center justify-center text-center py-14 px-6">
          {emptyIcon !== undefined ? (
            <div className="mb-4 text-muted-foreground/70">{emptyIcon}</div>
          ) : (
            <Container className="h-6 w-6 mb-4 text-muted-foreground/70" />
          )}
          <h3 className="text-sm font-semibold text-foreground">
            {emptyTitle}
            {emptyOptional && (
              <span className="ml-1.5 font-normal text-muted-foreground">(optional)</span>
            )}
          </h3>
          {emptyDescription && (
            <p className="text-[12.5px] text-muted-foreground mt-2 max-w-sm leading-relaxed">
              {emptyDescription}
            </p>
          )}
          {emptyOnAdd && emptyCtaLabel && (
            <Button
              type="button"
              variant="secondary"
              className="mt-5"
              onClick={emptyOnAdd}
            >
              + {emptyCtaLabel}
            </Button>
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
                  errors: errors[index] || EMPTY_ERRORS,
                })}
              </div>
            ))}
          </Accordion>
        </div>
      )}

    </div>
  );
}
