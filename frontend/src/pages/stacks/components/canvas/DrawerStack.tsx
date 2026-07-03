import { useEffect, type ReactNode } from "react";
import { entryKey, type DrawerEntry } from "@/pages/stacks/lib/canvas/drawer-stack";

const BASE_INSET_PX = 12;
const STAGGER_Y_PX = 10;
const STAGGER_X_PX = 16;
const FRONT_Z = 200;

export interface DrawerPanelDescriptor {
  entry: DrawerEntry;
  title: string;
  icon: ReactNode;
}

interface DrawerStackProps {
  /** Bottom → top; last item is the front (interactive) panel. */
  panels: DrawerPanelDescriptor[];
  /** Body for the front panel only. */
  front: ReactNode;
  /** Truncate the stack back to the panel at this panels-array index (it becomes the front). */
  onTruncate: (index: number) => void;
  onPop: () => void;
  onCloseAll: () => void;
}

/**
 * Floating, stackable drawer panels over the canvas (no backdrop — the canvas
 * stays interactive). Behind panels are header-only and dimmed; clicking one
 * truncates the stack back to it. Esc pops the front panel, Shift+Esc closes all.
 */
export function DrawerStack({ panels, front, onTruncate, onPop, onCloseAll }: DrawerStackProps) {
  const open = panels.length > 0;

  useEffect(() => {
    if (!open) return;
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key !== "Escape" || e.defaultPrevented) return;
      if (e.shiftKey) onCloseAll();
      else onPop();
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open, onPop, onCloseAll]);

  if (!open) return null;
  const frontIdx = panels.length - 1;

  return (
    <>
      {panels.map(({ entry, title, icon }, i) => {
        const depth = frontIdx - i; // 0 = front
        const isFront = depth === 0;
        return (
          <aside
            key={entryKey(entry)}
            data-testid={`drawer-panel-${i}`}
            className="fixed flex w-[600px] max-w-[calc(100vw-24px)] flex-col overflow-hidden rounded-lg border border-border bg-background shadow-lg animate-in slide-in-from-right-8 fade-in duration-[260ms]"
            style={{
              top: BASE_INSET_PX + STAGGER_Y_PX * depth,
              bottom: BASE_INSET_PX + STAGGER_Y_PX * depth,
              right: BASE_INSET_PX + STAGGER_X_PX * depth,
              zIndex: FRONT_Z - depth,
            }}
          >
            {isFront ? (
              front
            ) : (
              <button
                type="button"
                aria-label={`Bring ${title} panel to front`}
                onClick={() => onTruncate(i)}
                className="relative flex h-full w-full flex-col items-stretch text-left"
              >
                <span className="flex items-center gap-2.5 border-b border-border px-4 py-[15px]">
                  <span className="size-[19px] shrink-0 text-brand">{icon}</span>
                  <span className="truncate text-base font-medium text-foreground">{title}</span>
                </span>
                {/* Scrim: dims the parked panel (maps design rgba(10,14,20,.55)). */}
                <span aria-hidden className="absolute inset-0 bg-background/55" />
              </button>
            )}
          </aside>
        );
      })}
    </>
  );
}
