# Frontend Glossary

The React SPA (`frontend/src/`) is the operator console for the backend hub. It
reuses backend resource nouns (Stack, Cluster, Secret, ObjectStore,
PostgresAddon, StackDomain, Project, Organisation — see `../CONTEXT.md`) and adds
the UI-only concepts below. Prefer these canonical terms in components, hooks,
and schemas.

## Resource API surface

| term | definition | source |
|---|---|---|
| API client | The shared Axios instance and typed wrappers (`api/*.ts`) that call the backend per resource. | `frontend/src/api/client.ts`, `frontend/src/api/*.ts` |
| openapi types | Backend-generated TypeScript types (`components["schemas"][...]`) that the SPA derives its resource types from. | `frontend/src/api/types/openapi.d.ts` |
| zod schema | A runtime validation/parsing schema for form input and API payloads. | `frontend/src/api/zod-schemas.ts`, `frontend/src/pages/**/schemas/` |
| AppError | The normalized client-side error union (Axios API error or generic Error) shown to the user. | `frontend/src/api/client.ts` |

## Stack edit session

| term | definition | source |
|---|---|---|
| EditSession | The in-progress, client-side editing state for a Stack before changes are persisted; tracks tabs, drafts, and dirtiness (the `EditSessionState` interface). | `frontend/src/pages/stacks/hooks/use-stack-edit-session.ts` |
| EditSessionTab | The active section of a stack edit: `configuration`, `deployment`, or `environment`. | `frontend/src/pages/stacks/hooks/use-stack-edit-session.ts` |
| EditSessionDraft | The pending, unsaved edits held in an EditSession. | `frontend/src/pages/stacks/hooks/use-stack-edit-session.ts` |
| PerResourceDirty / PerVolumeDirty | Field-level dirty-tracking for each StackResource / Volume in an EditSession, enabling per-field discard. | `frontend/src/pages/stacks/lib/stack-diff.ts` |
| discardResourceField | Field-level discard of a pending edit, reverting one resource field back to its persisted value; siblings: `discardResource`, `discardVolume`, `discardEnvRow`. | `frontend/src/pages/stacks/hooks/use-stack-edit-session.ts` |
| StickyActionBar | The persistent save/discard action bar (primary/secondary/segment parts) shown while editing a Stack. | `frontend/src/components/sticky-action-bar.tsx` |

## Environment & addon binding

| term | definition | source |
|---|---|---|
| EnvAddonBinding | A link in a stack's environment that injects a PostgresAddon's connection info into a resource's env vars. | `frontend/src/pages/stacks/components/editor/tabs/architecture/drawer-tabs/environment-tab.tsx` |
| EnvAddonGroupState | The UI state of an env-addon group: `idle`, `editing-binding`, or `detaching`. | `frontend/src/pages/stacks/components/editor/tabs/architecture/drawer-tabs/environment-tab.tsx` |
| AddonBindingPatch | A pending change to an EnvAddonBinding awaiting save. | `frontend/src/pages/stacks/components/editor/tabs/architecture/drawer-tabs/env-row.tsx` |
| CredField | A PostgreSQL connection credential field name; some are cluster-wide (`CLUSTER_WIDE_FIELDS`). | `frontend/src/pages/stacks/lib/addon-presets.ts` |

## Docker Compose import

| term | definition | source |
|---|---|---|
| DockerComposeFile | A parsed `docker-compose` specification the user imports to scaffold a Stack. | `frontend/src/types/docker-compose.ts` |
| ConversionResult | The outcome of converting a DockerComposeFile into stack resources, including parse/conversion errors. | `frontend/src/lib/docker-compose-converter.ts` |
| ServiceConversionResult | The per-service result of Docker Compose → StackResource conversion. | `frontend/src/lib/docker-compose-converter.ts` |
| ImportActions | The callbacks driving the compose-import flow. | `frontend/src/pages/stacks/hooks/use-docker-compose-import.ts` |

## Plans, credentials & backups

| term | definition | source |
|---|---|---|
| PlanId | A named PostgreSQL addon sizing tier: `basic`, `starter`, `launch`, `scale`, `performance`, or `custom`. | `frontend/src/pages/addons/lib/plan-presets.ts` |
| detectPlan | The function that infers a PlanId from a PostgresAddon's resource spec. | `frontend/src/pages/addons/lib/payload.ts` |
| DomainName | The validated custom-domain value entered when binding a StackDomain. | `frontend/src/pages/domains/schemas/api-schema.ts` |
| eligibleRestoreSources | The logic that determines which backups/object stores are valid sources for a Postgres restore. | `frontend/src/pages/addons/lib/restore-sources.ts` |
| TriggerBackupPayload | The request body for on-demand triggering of a PostgresBackup. | `frontend/src/api/postgres-backups.ts` |

## Navigation & layout

| term | definition | source |
|---|---|---|
| AppLayout | The authenticated shell (sidebar + content) wrapping all in-app routes. | `frontend/src/components/app-layout.tsx` |
| RequireAuth | The route guard that redirects unauthenticated users to sign-in. | `frontend/src/App.tsx` |
| Breadcrumb context | The provider that registers per-route labels and loading state for the breadcrumb trail. | `frontend/src/contexts/breadcrumb-context.tsx` |
| Page | A route-level screen under `pages/` mapping to a resource area (stacks, clusters, secrets, object-stores, addons, domains, auth). | `frontend/src/App.tsx` (router) |
| SheetHeader | The sheet's own two-row top band: the title row (toggle, page title, one fact, actions) and a conditional toolbar row that collapses itself when nothing portals in. | `frontend/src/components/sheet-header.tsx` |

## Empty & no-result states

**Two different moments, not one component with a prop.** First run is where the
product defines its core noun and gets the decorated glyph; a filter that matched
nothing gets a 34px lens and a way back. See `DESIGN-PRODUCT.md` §11.

| term | definition | source |
|---|---|---|
| EmptyState | The centred title/description/action block shown when a list or page has nothing in it. **Draws no box** — it sits on the sheet like the rows it replaces. | `frontend/src/components/branded/empty-state.tsx` |
| SearchGlyph | The no-results mark: a lens in a 34px hairline circle. Used when a filter excluded everything, never for first run. | `frontend/src/components/branded/empty-state.tsx` |
| StackArchitectureGlyph | The first-run illustration — one stack card on the dot-grid canvas with its wires fading off the edges. Authored in the board's 148×88 units and drawn at `SCALE`. | `frontend/src/components/branded/empty-state.tsx` |
| board units | The convention for illustrations: every coordinate is the Figma node's own number, passed through a `u()` helper, so code and board can be diffed line by line and resizing is one constant. | `frontend/src/components/branded/empty-state.tsx` |

## Browser preview (`dev:mock`)

| term | definition | source |
|---|---|---|
| preview | The real app — real router, real shell, real pages — against an MSW-mocked network. No Go server, Postgres or Kind cluster. | `frontend/src/preview/start.ts` |
| preview scenario | Which dataset the preview serves, chosen by `VITE_PREVIEW_SCENARIO` **before** the service worker boots. `default` is the busy review dataset; `empty` is a brand-new org. | `frontend/src/preview/handlers.ts` |
| `dev:mock` / `dev:mock:empty` | The two preview servers, on **:5273** and **:5274**. Separate ports on purpose, so a populated screen and its first-run state sit side by side rather than being toggled. | `frontend/package.json` |
