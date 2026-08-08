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
}

export function ThemeToggle({ variant = "secondary" }: ThemeToggleProps) {
  const { setTheme } = useTheme();

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
