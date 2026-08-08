import { Moon, Sun } from "lucide-react";
import { Button, type buttonVariants } from "@/components/ui/button";
import { useTheme } from "@/hooks/use-theme";
import { cn } from "@/lib/utils";
import type { VariantProps } from "class-variance-authority";

interface ThemeToggleProps {
  /** `secondary` (default) is a raised control for standalone/floating
   *  placements (topbar, 404 page). `ghost` is flat, for a header row that
   *  already has its own chrome (the auth screen's nav-equivalent header) —
   *  hover swaps to `--muted` (an opaque wash) instead of `ghost`'s default
   *  translucent tint, since a flat header icon button sits on a gradient
   *  backdrop, not a solid one. */
  variant?: VariantProps<typeof buttonVariants>["variant"];
  /** `icon` (default) is the standalone circular control. `row` is the
   *  full-width sidebar form — icon + label on a 32px nav row — used where the
   *  toggle sits in the frame alongside the other global helpers. */
  presentation?: "icon" | "row";
}

export function ThemeToggle({ variant = "secondary", presentation = "icon" }: ThemeToggleProps) {
  const { setTheme } = useTheme();
  const toggle = () =>
    setTheme(document.documentElement.classList.contains("dark") ? "light" : "dark");

  if (presentation === "row") {
    return (
      <button
        type="button"
        onClick={toggle}
        // Geometry matches a nav row exactly — 32px, 8px inset, 10px gap, and
        // the label at weight 500 in ink. It sits in the same column, so any
        // difference reads as a mistake rather than a distinction.
        //
        // `relative` is load-bearing: the Moon is absolutely positioned, and
        // without a positioned ancestor here it resolves against whatever
        // happens to be positioned further up the tree.
        className="relative flex h-8 w-full items-center gap-2.5 rounded-md px-2 text-body font-medium text-sidebar-foreground transition-colors hover:bg-sidebar-accent focus-ring-edge"
      >
        <Sun className="size-4 shrink-0 rotate-0 scale-100 text-fg-2 transition-all dark:-rotate-90 dark:scale-0" />
        <Moon className="absolute size-4 shrink-0 rotate-90 scale-0 text-fg-2 transition-all dark:rotate-0 dark:scale-100" />
        {/* Collapsed there is no room for the word, and nothing clips it at
            this level — it would spill past the rail's edge. `rail-x` vacates
            the space and fades with the width rather than snapping. */}
        <span className="rail-x truncate">Appearance</span>
      </button>
    );
  }

  return (
    <Button
      variant={variant}
      size="icon"
      className={cn("rounded-full", variant === "ghost" && "hover:bg-muted")}
      onClick={() => setTheme(document.documentElement.classList.contains('dark') ? 'light' : 'dark')}
    >
      <Sun className="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
      <Moon className="absolute h-[1.2rem] w-[1.2rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
      <span className="sr-only">Toggle theme</span>
    </Button>
  );
}
