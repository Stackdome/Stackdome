# Design reference — Stack Detail v2 (env table + addon grouping)

**Date:** 2026-06-04
**Bundle:** Claude Design handoff `api.anthropic.com/v1/design/h/vUaEZk3pYpySPXUDa-sOyQ` → `stack-resource-page/`
**Scope of this ref:** the **Environment-tab table + inline addon group** from *Stack Detail v2* (the file the user flagged). The bundle also contains a broader "Approach A" redesign (session controller, soft-labels pass, addon detach state machine) — recorded as out-of-current-scope below.

## Component inventory (bundle → live target)
| Bundle source | Defines | Live target |
|---|---|---|
| `project/v2/v2-resources.jsx` `EnvBody` | env table rows + `from-pill` + `addon-group` fieldset | `components/detail/stack-resource-detail.tsx` (read), `components/shared/stack-resource-environment-tab.tsx` (edit) |
| `project/v2/v2.css` `.env-head/.env-row/.from-pill/.addon-group/.addon-group-head/.val.muted` | the visual rules | Tailwind classes via live tokens |
| `project/CLAUDE_CODE_HANDOFF.md` | full Approach-A scope + acceptance | (broader, deferred) |

## Token mapping (bundle `v2.css` ↔ live `frontend/src/index.css`)
| Bundle token | Bundle value (light) | Live token | Tailwind util | Match |
|---|---|---|---|---|
| `--bg` | oklch(0.96 0.015 85) | `--background` | `bg-background` | match (drift: live slightly lighter) |
| `--bg-elev` | oklch(0.985…) | `--secondary` | `bg-secondary` | map |
| `--fg1` | oklch(0.22…) | `--foreground` | `text-foreground` | match |
| `--fg2` | oklch(0.42…) | *(no mid-tone)* | `text-muted-foreground` | drift → use muted |
| `--fg-muted` | oklch(0.62…) | `--muted-foreground` | `text-muted-foreground` | match |
| `--border` | oklch(0.85…) | `--border` | `border-border` | match |
| `--border-strong` | oklch(0.78…) | `--border-strong` | `border-border-strong` | match |
| `--amber` / `--amber-press` | amber | `--brand` / `--brand-press` | `text-brand` / `text-brand-press` | match (brand IS the amber) |
| `--amber-soft` | amber/0.10 | `--brand-bg` | `bg-brand-bg` | match |
| `--r-2` / `--r-3` | 4px / 6px | `--radius` scale | `rounded` / `rounded-md` | match |
| PG mini badge | raw oklch blue | *(none — brand is amber)* | `AddonTypeIcon type="postgres"` | reuse existing component (no raw oklch) |

No new tokens. No raw hex/oklch in implementation — all via the live tokens above.

## Intent (from `chats/` + handoff, where the user landed)
The flat env list mixed addon-bound keys into the rows. v2 lands on: **each env row shows its source as a compact `from-pill`** (`Stack` / `Secret` / `Addon · host`), **reference values are masked** (`●●●●●● · master_key`, `●●● · host`), and **addon-bound keys are pulled into an inline dashed `addon-group` fieldset** whose header is a **notched legend** sitting on the top border (`PG  tooljet-db · Postgres`), with `+ Add variable from <addon>` / `Change addon` affordances in edit. The group reads as a scoped sub-region of the same table, not a separate card/tab.

## Acceptance (env-table aspect)
- [ ] Each env row's source renders as a `from-pill` (bordered, `bg-secondary`, sentence-case), addon rows as `Addon · <field>`.
- [ ] Secret value masked `●●●●●● · <key>`; addon value masked `●●● · <field>` (muted).
- [ ] Addon-bound keys live in a dashed `addon-group` with a **notched legend header** (postgres icon + `<addon> · Postgres`), on the table border — not a separate box.
- [ ] All via live tokens; dark mode resolves; no raw hex/oklch; mono only on keys/values, sans on pills/labels.

## Source-file map (for citation)
- `project/v2/v2-resources.jsx:203-241` — `env-head`, `env-row`, `from-pill`, `addon-group`/`addon-group-head`, `Addon · field` pill, `+ Add variable from`.
- `project/v2/v2.css:.env-head/.env-row/.from-pill/.addon-group/.addon-group-head` — exact dimensions (grid `1fr 1.4fr 110px 28px`; notched legend `position:absolute; top:-10px; left:12px; background:var(--bg)`; dashed `1px var(--border-strong)`).
- `project/CLAUDE_CODE_HANDOFF.md` — broader Approach-A (session bar, soft-labels, detach states) — **deferred**, not in this change.

## Out of current scope (deferred from the bundle)
Session-controller edit bar, full soft-labels typography pass, addon detach state machine (idle→editing→detaching→detached), per-row hover Edit. These are the larger Approach-A redesign; this change implements only the env-table + addon-group v2 visual the user flagged (image: read-mode grouping).
