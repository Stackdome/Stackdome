import { cn } from "@/lib/utils";

interface StackdomeMarkProps {
  size?: number;
  className?: string;
  variant?: "solid" | "tinted";
}

/**
 * Stackdome wordmark — three stacked slabs viewed in perspective with an amber dome.
 */
export function StackdomeMark({ size = 16, className, variant = "solid" }: StackdomeMarkProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={cn(variant === "tinted" ? "text-brand" : "text-foreground", className)}
    >
      {/* dome arc */}
      <path
        d="M5 9 A7 7 0 0 1 19 9"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        className="text-brand"
        fill="none"
      />
      {/* top slab */}
      <rect x="4" y="10" width="16" height="2.5" rx="0.5" fill="currentColor" opacity="0.95" />
      {/* mid slab */}
      <rect x="4" y="14" width="16" height="2.5" rx="0.5" fill="currentColor" opacity="0.7" />
      {/* bottom slab */}
      <rect x="4" y="18" width="16" height="2.5" rx="0.5" fill="currentColor" opacity="0.45" />
    </svg>
  );
}
