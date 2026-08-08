import { Search } from "lucide-react"

import { Input } from "@/components/ui/input"

/**
 * The one search box every catalogue tab uses.
 *
 * 32px, not 40 (§8/§11): a search is a **working control**, and 40 is reserved
 * for form fields and their primary button. There are no true form fields in
 * this flow — the configuration lives on the canvas.
 *
 * It runs the full width of the body column rather than the toolbar's 300px,
 * because here it is the only control above a long list rather than one of
 * several in a row.
 */
export function SearchField({
  value,
  onChange,
  placeholder,
  label,
}: {
  value: string
  onChange: (value: string) => void
  placeholder: string
  /** The accessible name. The placeholder disappears the moment you type. */
  label: string
}) {
  return (
    <div className="relative">
      <Search
        aria-hidden
        className="text-fg-muted absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2"
      />
      <Input
        type="search"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        aria-label={label}
        className="pl-[30px]"
      />
    </div>
  )
}
