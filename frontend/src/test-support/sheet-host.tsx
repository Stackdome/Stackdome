import type { ReactNode } from "react";
import { BreadcrumbProvider } from "@/contexts/breadcrumb-context";
import { SheetHeader } from "@/components/sheet-header";

/**
 * Renders a page under test inside the real sheet header (§8).
 *
 * A page's title and its actions do not render where the page is — the title is
 * the breadcrumb's last segment and the actions portal into `#topnav-actions`.
 * A page mounted bare therefore has no title and no buttons, so an assertion
 * that looks for either finds nothing and reports a failure the product does
 * not have.
 *
 * Must be rendered inside a router — the header reads the location to build the
 * trail, and the caller already owns the route it wants to be on.
 */
export function SheetHost({ children }: { children: ReactNode }) {
  return (
    <BreadcrumbProvider>
      <SheetHeader />
      {children}
    </BreadcrumbProvider>
  );
}
