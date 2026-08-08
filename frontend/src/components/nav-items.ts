import {
  Boxes,
  Cloud,
  GitBranch,
  GitPullRequest,
  Globe,
  KeyRound,
  Layers,
  Package,
  Puzzle,
  type LucideIcon,
} from "lucide-react";

export interface NavItem {
  label: string;
  path: string;
  icon: LucideIcon;
  /** Hidden from members. Mirrors the sidebar's previous `isOrgAdmin &&` gates. */
  adminOnly?: boolean;
  /**
   * Sub-paths that must NOT mark the item active. `/stacks/draft` is the
   * full-bleed editor on a draft — it is a different screen, not the list.
   */
  notActiveOn?: string[];
}

export interface NavGroup {
  /**
   * Omitted for the product's own destinations. §12: a label above the first
   * item is furniture — "Platform" over everything reported nothing.
   */
  label?: string;
  items: NavItem[];
}

/**
 * The sidebar, as a list rather than ten imports in a fixed order.
 *
 * Grouped by **who touches it** (§1), not by what kind of object it is: a
 * developer lives in the first group and never opens Infrastructure; a platform
 * admin lives in Infrastructure and rarely opens Stacks. That split is also
 * already how the permissions work.
 *
 * Order and labels come from the `app shell` Figma board.
 */
export const navGroups: NavGroup[] = [
  {
    items: [
      // `/stacks/new` deliberately stays active: it is the New stack journey,
      // and §12a drops that page's trail precisely BECAUSE the sidebar is still
      // saying which section you are in.
      { label: "Stacks", path: "/stacks", icon: Layers, notActiveOn: ["/stacks/draft"] },
      { label: "Previews", path: "/previews", icon: GitPullRequest },
    ],
  },
  {
    label: "Platform",
    items: [
      { label: "Addons", path: "/addons", icon: Puzzle },
      { label: "Secrets", path: "/secrets", icon: KeyRound },
      { label: "Object Stores", path: "/object-stores", icon: Cloud },
    ],
  },
  {
    label: "Infrastructure",
    items: [
      { label: "Clusters", path: "/clusters", icon: Boxes, adminOnly: true },
      { label: "Domains", path: "/domains", icon: Globe, adminOnly: true },
      { label: "Git Integrations", path: "/git-integrations", icon: GitBranch, adminOnly: true },
      { label: "Image Registries", path: "/image-registries", icon: Package, adminOnly: true },
    ],
  },
];

/** A destination is active on its own path and anything nested under it. */
export function isNavItemActive(pathname: string, item: NavItem): boolean {
  if (item.notActiveOn?.some((p) => pathname === p || pathname.startsWith(`${p}/`))) {
    return false;
  }
  return pathname === item.path || pathname.startsWith(`${item.path}/`);
}
