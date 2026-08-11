import { cn } from "@/lib/utils";

interface EyebrowLabelProps extends React.HTMLAttributes<HTMLSpanElement> {
  tone?: "brand" | "muted";
}

export function EyebrowLabel({ tone = "brand", className, ...props }: EyebrowLabelProps) {
  return (
    <span
      className={cn(tone === "brand" ? "eyebrow" : "eyebrow-muted", className)}
      {...props}
    />
  );
}
