import { Link, useLocation } from "react-router-dom";

import { SidebarMenuButton, SidebarMenuItem } from "@/components/ui/sidebar";
import { isNavItemActive, type NavItem as NavItemData } from "@/components/nav-items";

/**
 * One sidebar destination. Replaces the ten near-identical `nav-*.tsx`
 * components that differed only in an icon, a label and a path — so active
 * state, the collapsed tooltip and the admin gate now have one implementation
 * instead of ten.
 *
 * The label is ink at rest; the grey is carried by the icon (§7).
 */
export function NavItem({ item }: { item: NavItemData }) {
  const { pathname } = useLocation();
  const Icon = item.icon;

  return (
    <SidebarMenuItem>
      <SidebarMenuButton asChild tooltip={item.label} isActive={isNavItemActive(pathname, item)}>
        <Link to={item.path}>
          <Icon />
          {/* `rail-x` vacates the width and fades with the rail rather than
              being guillotined by the button's overflow mid-collapse. */}
          <span className="rail-x">{item.label}</span>
        </Link>
      </SidebarMenuButton>
    </SidebarMenuItem>
  );
}
