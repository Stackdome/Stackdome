import { BRAND_ICONS } from "./brand-icon-registry";

/** Themed brand logo for a registered slug; render only when `hasBrandIcon`. */
export function BrandIcon({ slug, size = 18 }: { slug: string; size?: number }) {
  const art = BRAND_ICONS[slug];
  if (!art) return null;
  const dims = { width: size, height: size };
  return (
    <>
      <img src={art.light} alt="" aria-hidden style={dims} className="object-contain dark:hidden" />
      <img src={art.dark} alt="" aria-hidden style={dims} className="hidden object-contain dark:block" />
    </>
  );
}
