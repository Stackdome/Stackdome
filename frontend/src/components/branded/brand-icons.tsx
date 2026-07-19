import { cn } from "@/lib/utils";
import { BRAND_ICONS } from "./brand-icon-registry";

/** Themed brand logo for a registered slug; render only when `hasBrandIcon`. */
export function BrandIcon({ slug, size = 18, className }: { slug: string; size?: number; className?: string }) {
  const art = BRAND_ICONS[slug];
  if (!art) return null;
  const dims = { width: size, height: size };
  return (
    <>
      <img src={art.light} alt="" aria-hidden style={dims} className={cn("object-contain dark:hidden", className)} />
      <img src={art.dark} alt="" aria-hidden style={dims} className={cn("hidden object-contain dark:block", className)} />
    </>
  );
}
