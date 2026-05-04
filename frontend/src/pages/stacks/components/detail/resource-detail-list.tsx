import { useState, useEffect, useCallback } from "react";
import type { ReactNode } from "react";
import { Accordion } from "@/components/ui/accordion";
import { Container } from "lucide-react";

// Generic detail list props that can be used for different types of details
interface ResourceDetailListProps<T> {
  items: T[];
  renderItem: (props: { item: T; index: number }) => ReactNode;
  emptyTitle?: string;
  emptyOptional?: boolean;
  emptyDescription?: string;
  emptyIcon?: ReactNode;
  defaultAllCollapsed?: boolean;
}

export default function ResourceDetailList<T>({
  items,
  renderItem,
  emptyTitle = "Nothing added yet",
  emptyOptional = false,
  emptyDescription,
  emptyIcon,
  defaultAllCollapsed = false,
}: ResourceDetailListProps<T>) {
  const [openAccordions, setOpenAccordions] = useState<string[]>(
    defaultAllCollapsed ? [] : items.map((_, i) => String(i))
  );

  // Update open accordions when items change
  useEffect(() => {
    if (!defaultAllCollapsed) {
      setOpenAccordions(items.map((_, i) => String(i)));
    }
  }, [items, items.length, defaultAllCollapsed]);

  const handleValueChange = useCallback((value: string[]) => {
    setOpenAccordions(value);
  }, []);

  // If no items, show empty state
  if (items.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center text-center py-14 px-6">
        {emptyIcon ? (
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
      </div>
    );
  }

  return (
    <div className="rounded-lg overflow-hidden">
      <Accordion
        type="multiple"
        value={openAccordions}
        onValueChange={handleValueChange}
        className="w-full"
      >
        {items.map((item, index) => renderItem({ item, index }))}
      </Accordion>
    </div>
  );
}
