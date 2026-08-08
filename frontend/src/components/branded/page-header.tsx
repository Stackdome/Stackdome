import { useEffect, useState } from "react";
import { createPortal } from "react-dom";

interface PageHeaderProps {
  eyebrow?: React.ReactNode;
  title?: React.ReactNode;
  status?: React.ReactNode;
  subtitle?: React.ReactNode;
  actions?: React.ReactNode;
  /**
   * The section's tools — search, filters, sort, the view toggle. Portals into
   * the header's second row (§12a). Omit it and the row collapses itself, so
   * the band is 56px instead of 100px with no height maths anywhere.
   */
  toolbar?: React.ReactNode;
  /** Accepted for call-site compatibility; the bar centres its own row. */
  actionsAlign?: "start" | "center";
  className?: string;
}

/**
 * A page's contribution to the sheet header (§8).
 *
 * **It renders nothing where you put it.** The sheet has exactly one header —
 * the 52px bar at its top — so this portals `status` and `actions` into that
 * bar's right slot and puts no band on the page. Leaving it in place is what
 * made every screen say its own name twice: once in the breadcrumb, once in a
 * title 50px below it.
 *
 * Three props are accepted and **ignored**, so the fifteen existing call sites
 * still compile while they are cleaned up one at a time:
 *
 * - `title` — the bar takes the page title from the breadcrumb's last segment.
 *   A page that needs a different word registers it with `setCustomLabel`.
 * - `eyebrow` — an orange "Platform" over a title reports nothing.
 * - `subtitle` — explanatory copy belongs to the empty state, where you
 *   actually need it. Once the list has rows, the rows explain themselves.
 *
 * Order matters: `status` is the page's one fact, `actions` follow it.
 *
 * `toolbar` is the header's **second** row (§12a) and portals into its own slot.
 * A page's tools belong to the sheet the same way its actions do — rendering
 * them in the page body is what left the Stacks filters 20px further in than
 * the title above them.
 */
export function PageHeader({ status, actions, toolbar }: PageHeaderProps) {
  const [slots, setSlots] = useState<{ actions: HTMLElement | null; toolbar: HTMLElement | null }>({
    actions: null,
    toolbar: null,
  });

  useEffect(() => {
    setSlots({
      actions: document.getElementById("topnav-actions"),
      toolbar: document.getElementById("sheet-toolbar"),
    });
  }, []);

  return (
    <>
      {slots.actions &&
        (status || actions) &&
        createPortal(
          <>
            {/* 8px comes from the slot's own gap; this adds the 4px that takes
                the fact-to-action distance to the 12px §12a asks for. */}
            {status && <span className="mr-1 flex items-center">{status}</span>}
            {actions}
          </>,
          slots.actions,
        )}
      {slots.toolbar && toolbar && createPortal(toolbar, slots.toolbar)}
    </>
  );
}
