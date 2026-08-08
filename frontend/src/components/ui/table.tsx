import * as React from "react"

import { cn } from "@/lib/utils"

/**
 * §7 — the list is not boxed. A border drawn around the whole table is the card
 * mistake at a larger scale: it says "this is a card", and §1 reserves that for
 * nothing. The rows and the sheet edge are the only boundaries there are.
 */
function Table({ className, ...props }: React.ComponentProps<"table">) {
  return (
    <div
      data-slot="table-container"
      className="relative w-full overflow-x-auto"
    >
      <table
        data-slot="table"
        className={cn("w-full caption-bottom text-body", className)}
        {...props}
      />
    </div>
  )
}

function TableHeader({ className, ...props }: React.ComponentProps<"thead">) {
  return (
    <thead
      data-slot="table-header"
      className={cn("[&_tr]:border-b [&_tr]:border-border", className)}
      {...props}
    />
  )
}

function TableBody({ className, ...props }: React.ComponentProps<"tbody">) {
  return (
    <tbody
      data-slot="table-body"
      className={cn("[&_tr:last-child]:border-0", className)}
      {...props}
    />
  )
}

function TableFooter({ className, ...props }: React.ComponentProps<"tfoot">) {
  return (
    <tfoot
      data-slot="table-footer"
      className={cn(
        "bg-muted/50 border-t font-medium [&>tr]:last:border-b-0",
        className
      )}
      {...props}
    />
  )
}

function TableRow({ className, ...props }: React.ComponentProps<"tr">) {
  return (
    <tr
      data-slot="table-row"
      className={cn(
        "group/row hover:bg-muted/50 data-[state=selected]:bg-muted border-b border-border transition-colors",
        className
      )}
      {...props}
    />
  )
}

function TableHead({ className, ...props }: React.ComponentProps<"th">) {
  return (
    <th
      data-slot="table-head"
      className={cn(
        // §7 — sentence case, `text-label`, `fg-muted`. Size and colour already
        // say "this is a header"; uppercase would be a third signal doing a job
        // that is already done, and it costs the word-shape the eye reads by.
        "text-fg-muted h-8 px-2 text-left align-middle text-label font-medium whitespace-nowrap normal-case [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px]",
        className
      )}
      {...props}
    />
  )
}

function TableCell({ className, ...props }: React.ComponentProps<"td">) {
  return (
    <td
      data-slot="table-cell"
      className={cn(
        "px-2 py-2.5 align-middle text-body whitespace-nowrap [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px]",
        className
      )}
      {...props}
    />
  )
}

/**
 * §7 — row actions appear on hover. A kebab on every row at rest is chrome
 * competing with content.
 *
 * Hidden by opacity rather than by mounting, so the control keeps its place in
 * the tab order and the row does not reflow when the pointer arrives. Keyboard
 * users get it via focus-within; an open menu holds it visible so the trigger
 * does not vanish from under its own popover.
 */
function TableRowActions({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="table-row-actions"
      className={cn(
        "flex items-center justify-end gap-1 opacity-0 transition-opacity",
        "group-hover/row:opacity-100 group-focus-within/row:opacity-100",
        "has-[[data-state=open]]:opacity-100",
        className
      )}
      {...props}
    />
  )
}

function TableCaption({
  className,
  ...props
}: React.ComponentProps<"caption">) {
  return (
    <caption
      data-slot="table-caption"
      className={cn("text-muted-foreground mt-4 text-body", className)}
      {...props}
    />
  )
}

export {
  Table,
  TableHeader,
  TableBody,
  TableFooter,
  TableHead,
  TableRow,
  TableRowActions,
  TableCell,
  TableCaption,
}
