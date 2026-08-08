import { http, HttpResponse } from 'msw'

import {
  makeAddon,
  makeCluster,
  makeProject,
  makeStack,
  makeUser,
  ORG_ID,
} from '../../.storybook/fixtures'
import { ReleaseState } from '@/pages/stacks/components/editor/tabs/deployments/release-states'
import type { Stack } from '@/api/stack-types'
import type { ObjectStore } from '@/api/object-stores'

/**
 * Network for the browser preview (`pnpm dev:mock`).
 *
 * Reuses the Storybook fixtures rather than inventing a second set — the
 * factories are already typed off the generated OpenAPI schemas, so a response
 * that drifts from the contract fails to compile.
 *
 * The dataset is chosen for *design review*: a spread of states on the busiest
 * screen, and enough on every other destination that no page is judged on an
 * empty state it would rarely be in.
 */

const list = (items: unknown[]) => HttpResponse.json({ items, total: items.length })

/**
 * Which dataset the preview serves. Set by the npm script, not by a URL param —
 * a scenario has to be chosen before the service worker starts, and the state
 * being reviewed here is *first run*, which no amount of clicking can reach
 * once fixtures exist.
 *
 *   `pnpm dev:mock`        → the review dataset (busy, every status)
 *   `pnpm dev:mock:empty`  → a brand-new org with nothing in it
 *
 * They run on different ports deliberately, so the two can sit side by side in
 * a browser rather than being toggled and re-toggled.
 */
const SCENARIO = import.meta.env.VITE_PREVIEW_SCENARIO ?? 'default'
const isEmpty = SCENARIO === 'empty'

/**
 * A stack's spec, varied — every stack having the same two resources and the
 * same timestamp is half of why the list read flat in review. The columns were
 * fine; the data had nothing to say.
 *
 * `name` alone builds from git; `name@image:tag` runs a prebuilt image.
 *
 * The distinction matters for review: the component chips infer their icon from
 * the image, so a stack of pure git builds shows nothing but generic glyphs and
 * the brand logos never appear. At least one fixture has to run real software.
 */
const spec = (resources: string[], volumes: string[], branch: string, commit?: string) =>
  ({
    stack_resources: resources.map((entry) => {
      const [name, image] = entry.split('@')
      return {
        name,
        // Same literal every other fixture and `git-source-seed.ts` use.
        workload_type: 'Service',
        source: image
          ? { image: { ref: image } }
          : {
            git: {
              repo_url: 'https://github.com/acme/monorepo',
              branch,
              commit,
              dockerfile_path: 'Dockerfile',
              build_context: '.',
            },
          },
      }
    }),
    volumes: volumes.map((name) => ({ name })),
  }) as Stack['spec']

const stacks = [
  makeStack({
    id: 's3',
    name: 'docs-site',
    spec: spec(['web'], [], 'fix/og-tags', 'e91a02c4d5e6f7a8'),
    updated_at: '2026-08-05T15:40:00Z',
    latest_release: {
      id: 'r3',
      state: ReleaseState.Failed,
      message: 'web · image pull failed — ghcr.io/acme/docs:e91a02 not found',
      completed_at: '2026-08-05T15:40:00Z',
    },
  } as Partial<Stack>),
  makeStack({
    id: 's2',
    name: 'billing-worker',
    spec: spec(['api', 'worker', 'billing-db@postgres:16'], ['billing-data'], 'main', '7c14be91a02c4d5e'),
    updated_at: '2026-08-05T16:00:00Z',
    latest_release: {
      id: 'r2',
      state: ReleaseState.InProgress,
      message: '2 of 3 services updated',
      created_at: '2026-08-05T16:00:00Z',
    },
  } as Partial<Stack>),
  makeStack({
    id: 's5',
    name: 'auth-gateway',
    spec: spec(['edge', 'session', 'tokens', 'sessions@redis:7'], ['certs'], 'main', '11fd7c14be91a02c'),
    updated_at: '2026-07-30T09:00:00Z',
    latest_release: { id: 'r5', state: ReleaseState.Released },
    converged_release: {
      id: 'r5',
      state: ReleaseState.Released,
      health: 'degraded',
      message: 'session · 1 of 3 replicas available',
      completed_at: '2026-07-30T09:00:00Z',
    },
  } as Partial<Stack>),
  makeStack({
    spec: spec(['web', 'worker', 'orders-db@postgres:16', 'cache@redis:7'], ['uploads', 'assets'], 'main', 'a3f9d2e91a02c4d5'),
    updated_at: '2026-07-31T12:00:00Z',
    latest_release: { id: 'r1', state: ReleaseState.Released },
    converged_release: {
      id: 'r1',
      state: ReleaseState.Released,
      health: 'ok',
      completed_at: '2026-07-31T12:00:00Z',
    },
  } as Partial<Stack>),
  makeStack({ id: 's4', name: 'staging-sandbox', spec: spec(['web'], ['data'], 'main') }),
  makeStack({ id: 's6', name: 'search-indexer', spec: spec(['indexer', 'reaper', 'search@elasticsearch:8.13'], [], 'main') }),
  makeStack({ id: 's7', name: 'notifications', spec: spec(['dispatcher'], [], 'main') }),
  makeStack({
    id: 's8',
    name: 'admin-console-with-a-deliberately-long-name',
    spec: spec(['web'], [], 'chore/rename-everything-for-the-truncation-test'),
  }),
].map((s) => ({ ...s, project_id: makeProject().id })) as Stack[]

const addons = [
  makeAddon(),
  makeAddon({ id: 'pg-2', name: 'billing-db', status: { state: 'Ready' } }),
  makeAddon({ id: 'pg-3', name: 'analytics-db', status: { state: 'Creating' } }),
]

const clusters = [
  makeCluster(),
  makeCluster({ id: 'c2', name: 'eu-west-1' }),
]

const secrets = [
  { id: 'sec-1', name: 'STRIPE_API_KEY', created_at: '2026-07-20T10:00:00Z' },
  { id: 'sec-2', name: 'SENTRY_DSN', created_at: '2026-07-22T10:00:00Z' },
  { id: 'sec-3', name: 'SMTP_PASSWORD', created_at: '2026-07-29T10:00:00Z' },
]

/**
 * Typed, because the untyped version was wrong and crashed the page.
 *
 * This was a hand-written `{ provider: 's3', bucket: … }` literal — fields the
 * API does not have — so `/object-stores` died in the router's error boundary
 * on `store.spec.configuration`. Every other fixture here comes off a factory
 * that is typed against the generated schemas, which is exactly why none of
 * them could drift like this. Annotating it restores that guarantee.
 */
const objectStores: ObjectStore[] = [
  {
    id: 'os-1',
    name: 'backups',
    spec: {
      configuration: {
        s3_credentials: {
          access_key_id: { secret_id: 'sec-aws', key: 'access_key_id' },
          secret_access_key: { secret_id: 'sec-aws', key: 'secret_access_key' },
          region: 'eu-west-1',
        },
      },
      destination_path: 's3://acme-backups/postgres',
      retention_policy: '7d',
    },
    created_at: '2026-07-11T10:00:00Z',
  },
]

// Shaped to the real `GitIntegration` schema — `type`, a flat `status` enum and
// `credentials_configured` are what `usableIntegrations()` filters on, and
// without them the New stack page's repository tab shows "no provider
// connected" no matter what this list contains.
const gitIntegrations = [
  {
    id: 'gi-1',
    type: 'github_app',
    host: 'github.com',
    status: 'installed',
    credentials_configured: true,
    install_url: 'https://github.com/apps/stackdome/installations/new',
    created_at: '2026-07-02T10:00:00Z',
  },
]

const repositories = [
  { full_name: 'acme/checkout-api', clone_url: 'https://github.com/acme/checkout-api.git', default_branch: 'main', owner: 'acme' , pushed_at: '2026-08-07T18:00:00Z' },
  { full_name: 'acme/web-storefront', clone_url: 'https://github.com/acme/web-storefront.git', default_branch: 'main', owner: 'acme' , pushed_at: '2026-08-06T09:00:00Z' },
  { full_name: 'acme/billing-worker', clone_url: 'https://github.com/acme/billing-worker.git', default_branch: 'main', owner: 'acme' , pushed_at: '2026-08-04T11:00:00Z' },
  { full_name: 'acme/notifications', clone_url: 'https://github.com/acme/notifications.git', default_branch: 'develop', owner: 'acme' , pushed_at: '2026-08-02T15:00:00Z' },
  { full_name: 'acme/image-proxy', clone_url: 'https://github.com/acme/image-proxy.git', default_branch: 'main', owner: 'acme' , pushed_at: '2026-07-24T08:00:00Z' },
]

const imageRegistries = [
  { id: 'ir-1', name: 'ghcr', url: 'ghcr.io/acme', created_at: '2026-07-04T10:00:00Z' },
]

const ORG = `/api/v1/organizations/:orgId`
const PROJECT = `${ORG}/projects/:projectName`

export const previewHandlers = [
  // ── identity ──────────────────────────────────────────────────────────
  http.get('/api/v1/config', () => HttpResponse.json({})),
  http.get('/api/v1/users/current', () => HttpResponse.json(makeUser())),
  http.get('/api/v1/users/current/projects', () => list([makeProject()])),
  http.post('/api/v1/auth/refresh', () =>
    HttpResponse.json({ token: 'preview-token', refreshToken: 'preview-refresh' }),
  ),
  http.get(`/api/v1/organizations/${ORG_ID}`, () =>
    HttpResponse.json({ id: ORG_ID, name: 'acme' }),
  ),

  // ── the destinations in the sidebar ───────────────────────────────────
  // The `empty` scenario answers with nothing so the first-run state is
  // reachable. Everything else stays populated — a new org still has a
  // project and a cluster, and blanking those would test a different screen.
  http.get(`${ORG}/stacks`, () => list(isEmpty ? [] : stacks)),
  // Without this the catch-all below answered with an empty LIST, the editor
  // read `spec` off it, and every route past /stacks died on an error boundary
  // — so the journey could not be reviewed at all.
  http.get(`${PROJECT}/stacks/:id`, ({ params }) =>
    HttpResponse.json(stacks.find((s) => s.id === params.id) ?? stacks[0]),
  ),
  http.get(`${ORG}/stacks/:id`, ({ params }) =>
    HttpResponse.json(stacks.find((s) => s.id === params.id) ?? stacks[0]),
  ),
  http.get(`${ORG}/projects`, () => list([makeProject()])),
  http.get(`${ORG}/secrets`, () => list(secrets)),
  http.get(`${ORG}/object-stores`, () => list(objectStores)),
  http.get(`${ORG}/clusters`, () => list(clusters)),
  http.get(`${ORG}/git-integrations/:id/repositories`, () =>
    HttpResponse.json({ items: repositories, page: 1, total_count: repositories.length, has_next: false }),
  ),
  // A brand-new org has connected nothing. This is what makes the New stack
  // page's "no git provider" state reachable at all — with a provider in the
  // fixtures, no amount of clicking gets you to the first screen a new user
  // actually lands on.
  http.get(`${ORG}/git-integrations`, () => list(isEmpty ? [] : gitIntegrations)),
  http.get(`${ORG}/image_registries`, () => list(imageRegistries)),
  http.get(`${ORG}/users`, () => list([makeUser()])),
  http.get(`${ORG}/invites`, () => list([])),
  http.get(`${PROJECT}/addons/postgres`, () => list(addons)),
  http.get(`${PROJECT}/object-stores`, () => list(objectStores)),
  http.get(`${PROJECT}/stack-preview-configs`, () => list([])),
  http.get(`${PROJECT}/preview-stacks`, () => list([])),

  // ── everything else ───────────────────────────────────────────────────
  // A page that hits an endpoint nobody mocked should render its empty state,
  // not an error boundary — the preview exists to look at layouts, and an
  // unmocked GET is a gap in this file rather than something to design around.
  http.get('/api/v1/*', () => list([])),
  http.post('/api/v1/*', () => HttpResponse.json({}, { status: 200 })),
  http.put('/api/v1/*', () => HttpResponse.json({}, { status: 200 })),
  http.patch('/api/v1/*', () => HttpResponse.json({}, { status: 200 })),
  http.delete('/api/v1/*', () => HttpResponse.json({}, { status: 204 })),
]
