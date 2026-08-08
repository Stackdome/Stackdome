import { useCallback, useState } from "react";
import { LayoutGrid, List } from "lucide-react";
import { SegmentedControl } from "@/components/ui/segmented-control";

export type ViewMode = "list" | "cards";

const OPTIONS = [
  { value: "list" as const, label: "List", icon: <List /> },
  { value: "cards" as const, label: "Cards", icon: <LayoutGrid /> },
];

const storageKey = (page: string) => `stackdome.view.${page}`;

/**
 * Which view a page is in, remembered per page per user (§7).
 *
 * **List is the default.** It is the denser view and the one that compares, so
 * a page with no stored preference opens in it.
 *
 * Persistence is `localStorage`, keyed by page: switching pages does not reset
 * the choice, and two pages do not share one. Reads are lazy — the initial
 * state runs once rather than on every render — and a storage failure (private
 * mode, quota) falls back to the default rather than taking the page down.
 */
export function useViewMode(page: string): [ViewMode, (v: ViewMode) => void] {
  const [mode, setMode] = useState<ViewMode>(() => {
    try {
      return localStorage.getItem(storageKey(page)) === "cards" ? "cards" : "list";
    } catch {
      return "list";
    }
  });

  const set = useCallback(
    (v: ViewMode) => {
      setMode(v);
      try {
        localStorage.setItem(storageKey(page), v);
      } catch {
        // A remembered preference is a convenience; losing it is not a failure
        // worth surfacing.
      }
    },
    [page],
  );

  return [mode, set];
}

/**
 * The list/cards toggle (§7) — a segmented control, icon-only, two options.
 * Not two buttons, not a dropdown.
 *
 * Lives in the content toolbar, **right side, last** — after the filters,
 * before nothing.
 */
export function ViewToggle({
  value,
  onValueChange,
}: {
  value: ViewMode;
  onValueChange: (v: ViewMode) => void;
}) {
  return (
    <SegmentedControl
      options={OPTIONS}
      value={value}
      onValueChange={onValueChange}
      size="sm"
      aria-label="View"
    />
  );
}
