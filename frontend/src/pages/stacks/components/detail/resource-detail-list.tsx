import { useState, useEffect, useCallback } from "react";
import type { ReactNode } from "react";
import { Accordion } from "@/components/ui/accordion";
import { Container } from "lucide-react";

// Generic detail list props that can be used for different types of details
interface ResourceDetailListProps<T> {
  items: T[];
  renderItem: (props: { item: T; index: number }) => ReactNode;
  emptyText?: string;
  emptyIcon?: ReactNode;
  defaultAllCollapsed?: boolean;
}

export default function ResourceDetailList<T>({
  items,
  renderItem,
  emptyText = "No Resources found.",
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
      <div className="flex flex-col items-center justify-center py-12 bg-muted/20 rounded-md">
        {emptyIcon ? (
          <>{emptyIcon}</>
        ) : (
          <Container className="mx-auto h-8 w-8 mb-2 text-muted-foreground" />
        )}
        <div className="text-lg text-muted-foreground font-medium mb-1">{emptyText}</div>
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
