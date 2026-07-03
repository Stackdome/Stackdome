# Environment tab redesign — plan

**Date:** 2026-07-04
**Design ref:** `docs/superpowers/design-refs/2026-07-04-env-tab-design-reference.md` (sources: `ResourceConfig.dc.html`, `EnvRow.dc.html`)
**User decisions:** keep paste/import/clear-all functionality but restyle to design ghost-button aesthetic; skip "raw editor" entirely; env tab only (no DEPENDS ON chips / drawer changes-chip in this slice).

## Scope

Visual restyle of the drawer Environment tab. All logic (5-arm `FormEnvVarSchema`, `envRowsDiff` dirty tracking, addon grouping per (addonId, database), orphan detection, paste/import parsing) stays untouched.

### `stack-resource-environment-tab.tsx`
1. Header → left: `Environment · N variables` (12.5px secondary + mono muted count); right: three ghost chips (mono 11px, bordered, hover→brand; clear-all hovers danger): **clear all**, **paste .env** (existing dialog), **import file** (existing dialog).
2. Delete outer table container + Key/Value/From header grid row.
3. Plain rows render in `flex flex-col gap-1.5` (6px), each an `EnvRow` card.
4. "+ Add variable" becomes full-width dashed centered ghost row (hover→brand), between plain rows and addon groups. Empty state = ghost row alone.
5. Addon group → dashed `brand-border` rounded container with `brand-bg` tint, floating legend chip on the top border (Plug icon "ADDON" marker + addon select + `·` + db select/static), rows inside as cards, right-aligned brand "Add binding".

### `env-row.tsx`
6. Row → relative card: `rounded-md border`, padding ~`9px 10px 9px 13px`, absolute 2px left accent bar (top/bottom 6px). States: dirty (added or modified) = brand accent + `bg-brand-bg` + `border-brand-border`; orphan addon = warn equivalents; unchanged = plain `border-border bg-background`.
7. Inner layout `flex items-start gap-2`: KEY input 130px mono 12px h-8 | value column `flex-1 min-w-0 flex-col gap-1.5` with per-type stacked h-8 selects | From select 106px h-8 | 24px action cell (reset brand when modified, × hover-danger otherwise).

### Token mapping
Design dark-theme amber → `--brand` family (`text-brand`, `bg-brand-bg`, `border-brand-border`); `--fg2/--fg3/--fg-muted` → `text-muted-foreground` (± /70); `--r-2` → `rounded-md`; `--err` → `danger`. No raw hex.

### Tests
Update to new DOM, keep green: `shared/tests/env-row.test.tsx`, `shared/tests/env-addon-group-render.test.tsx`, `detail/tests/stack-resource-detail-env.test.tsx`, form-schema tests. All `data-testid`s preserved (`env-row-*`, `env-addon-group`, `addon-picker-trigger`, `database-picker-trigger`, `field-picker-trigger`, `resource-picker-trigger`, `resource-output-trigger`, `self-output-trigger`).

Gate: vitest green + lint + `tsc -b` no new errors in touched files.
