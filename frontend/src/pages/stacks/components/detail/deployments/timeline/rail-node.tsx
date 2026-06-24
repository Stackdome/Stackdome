import { Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import type { Tone } from "../derive";
import { toneDotClass, toneTextClass } from "../derive";

export type RailDotShape = "solid" | "ring" | "spinner" | "dashed";

export interface RailNodeProps {
  tone: Tone;
  /** Dot rendering. Defaults to "ring" when `big`, else "solid". `dashed` is
   *  always brand-amber (the draft signal), independent of tone. */
  shape?: RailDotShape;
  big?: boolean;
  pulse?: boolean;
  isLast?: boolean;
  children: React.ReactNode;
}

export function RailNode({ tone, shape, big, pulse, isLast, children }: RailNodeProps) {
  const resolved: RailDotShape = shape ?? (big ? "ring" : "solid");

  const dot =
    resolved === "spinner" ? (
      <Loader2 data-testid="rail-dot" className={cn("mt-[13px] h-3.5 w-3.5 flex-none animate-spin", toneTextClass(tone))} />
    ) : resolved === "dashed" ? (
      <span data-testid="rail-dot" className="mt-[13px] h-3 w-3 flex-none rounded-full border-2 border-dashed border-brand bg-background" />
    ) : resolved === "ring" ? (
      <span data-testid="rail-dot" className={cn("mt-[14px] h-2.5 w-2.5 flex-none rounded-full border-2 border-current bg-background", toneTextClass(tone), pulse && "animate-pulse")} />
    ) : (
      <span data-testid="rail-dot" className={cn("mt-[15px] h-2.5 w-2.5 flex-none rounded-full", toneDotClass(tone), pulse && "animate-pulse")} />
    );

  return (
    <div className="flex items-stretch gap-3.5">
      <div className="flex w-[34px] flex-none flex-col items-center">
        {dot}
        <span
          data-testid="rail-connector"
          className={["mt-1 w-px flex-1 bg-border", isLast ? "invisible" : "visible"].join(" ")}
          style={{ minHeight: isLast ? 0 : 16 }}
        />
      </div>
      <div className={["min-w-0 flex-1", isLast ? "pb-1" : "pb-5"].join(" ")}>{children}</div>
    </div>
  );
}
