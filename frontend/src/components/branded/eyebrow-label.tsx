import { cn } from "@/lib/utils";

interface EyebrowLabelProps extends React.HTMLAttributes<HTMLSpanElement> {
  tone?: "brand" | "muted";
}

export function EyebrowLabel({ tone = "brand", className, ...props }: EyebrowLabelProps) {
  return (
    <span
      className={cn(
        "font-mono text-[11px] font-medium uppercase tracking-[1.5px]",
        tone === "brand" ? "text-brand" : "text-muted-foreground",
        className,
      )}
      {...props}
    />
  );
}
