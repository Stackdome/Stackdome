import { cn } from "@/lib/utils";

interface StackdomeMarkProps {
  size?: number;
  className?: string;
  tone?: "default" | "brand";
}

/**
 * Stackdome brand mark — hexagonal house outline in amber with three
 * stacked diamond slabs. Geometry mirrors design-system asset
 * `logo-mark.svg` (viewBox -60 -60 120 120).
 *
 * `tone="brand"` paints the whole mark in amber (slabs included) so it
 * pairs cleanly with brand-color text in a lockup.
 */
export function StackdomeMark({ size = 24, className, tone = "default" }: StackdomeMarkProps) {
  const monoBrand = tone === "brand";
  return (
    <svg
      width={size}
      height={size}
      viewBox="-60 -60 120 120"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={cn(monoBrand && "text-brand", className)}
    >
      <path
        d="M -25,50 L -50,25 L -50,-25 L 0,-50 L 50,-25 L 50,25 L 25,50"
        stroke={monoBrand ? "currentColor" : "var(--brand)"}
        strokeWidth="8"
        strokeLinejoin="miter"
        strokeLinecap="butt"
        strokeMiterlimit="10"
        fill="none"
      />
      {monoBrand ? (
        <>
          <path d="M -32.5,18 L 0,11 L 32.5,18 L 0,25 Z" fill="currentColor" opacity="0.55" />
          <path d="M -30,9 L 0,2 L 30,9 L 0,16 Z" fill="currentColor" opacity="0.75" />
          <path d="M -27.5,0 L 0,-7 L 27.5,0 L 0,7 Z" fill="currentColor" />
        </>
      ) : (
        <>
          <path d="M -32.5,18 L 0,11 L 32.5,18 L 0,25 Z" className="fill-slate-400" />
          <path d="M -30,9 L 0,2 L 30,9 L 0,16 Z" className="fill-slate-300" />
          <path d="M -27.5,0 L 0,-7 L 27.5,0 L 0,7 Z" className="fill-foreground" />
        </>
      )}
    </svg>
  );
}

/**
 * Horizontal lockup: mark + lowercase "stackdome" wordmark in Geist 600
 * with 0.04em tracking. Matches `logo-horizontal.svg`.
 */
export function StackdomeWordmark({
  size = 22,
  className,
}: {
  size?: number;
  className?: string;
}) {
  return (
    <span className={cn("inline-flex items-center gap-2.5", className)}>
      <StackdomeMark size={size} />
      <span
        className="font-semibold lowercase leading-none text-foreground"
        style={{ letterSpacing: "0.04em", fontSize: `${size}px` }}
      >
        stackdome
      </span>
    </span>
  );
}
