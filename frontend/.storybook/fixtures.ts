import type { components } from '../src/api/types/openapi'
import { ReleaseEventScope, ReleaseEventType } from '../src/api/releases'
import { BuildPhase, BUILD_JOB_CREATED_CONDITION } from '../src/api/image-builds'
import { ReleaseState } from '../src/pages/stacks/components/editor/tabs/deployments/release-states'

type Schemas = components['schemas']
export type User = Schemas['User']
export type Project = Schemas['Project']
export type Stack = Schemas['Stack']
export type StackRelease = Schemas['StackRelease']
export type StackReleaseDetail = Schemas['StackReleaseDetail']
export type ReleaseEvent = Schemas['ReleaseEvent']
export type PostgresAddon = Schemas['PostgresAddon']
export type PostgresBackup = Schemas['PostgresBackup']
export type Cluster = Schemas['Cluster']
export type ImageBuild = Schemas['ImageBuild']

export const ORG_ID = 'org-1'
export const DEFAULT_PROJECT = 'default'
export const STACK_ID = 's1'

export function makeUser(overrides: Partial<User> = {}): User {
  return {
    id: 'u1',
    name: 'Ada Lovelace',
    username: 'ada',
    email: 'ada@example.com',
    organisation: 'acme',
    organisation_id: ORG_ID,
    role: 'OrgAdmin',
    projects: [],
    ...overrides,
  }
}

export function makeProject(overrides: Partial<Project> = {}): Project {
  return {
    id: 'p1',
    name: DEFAULT_PROJECT,
    organisation_id: ORG_ID,
    default_project: true,
    ...overrides,
  }
}

// Same base shape the unit tests use (stack-card.test.tsx); the generated Stack
// schema marks everything optional-readonly, so a partial literal cast is the
// established pattern for fixtures.
export function makeStack(overrides: Partial<Stack> = {}): Stack {
  return {
    id: STACK_ID,
    name: 'orders-api',
    namespace: 'ns-orders-api',
    revision: 4,
    spec: {
      stack_resources: [{ name: 'web' }, { name: 'worker' }],
      volumes: [{}],
    },
    updated_at: '2026-07-30T12:00:00Z',
    ...overrides,
  } as Stack
}

export function makeAddon(overrides: Partial<PostgresAddon> = {}): PostgresAddon {
  return {
    id: 'pg-1',
    name: 'orders-db',
    project_id: 'p1',
    spec: {
      version: { major: 16 },
      instances: { count: 1 },
      storage: { size: '10Gi' },
    },
    status: { state: 'Ready' },
    created_at: '2026-07-28T09:00:00Z',
    ...overrides,
  } as PostgresAddon
}

export function makeBackup(overrides: Partial<PostgresBackup> = {}): PostgresBackup {
  return {
    id: 'bk-1',
    name: 'orders-db-backup-1',
    type: 'manual',
    phase: 'completed',
    started_at: '2026-07-29T02:00:00Z',
    completed_at: '2026-07-29T02:04:30Z',
    size_bytes: 734003200,
    ...overrides,
  }
}

export function makeRelease(overrides: Partial<StackRelease> = {}): StackRelease {
  return {
    id: 'rel-12',
    stack_id: STACK_ID,
    sequence: 12,
    state: ReleaseState.Released,
    cause: { kind: 'manual' },
    pins: { resources: { web: { git_sha: 'a1b2c3d4e5f6a7b8' } } },
    created_at: '2026-07-30T11:58:00Z',
    rendered_at: '2026-07-30T11:58:10Z',
    completed_at: '2026-07-30T11:59:25Z',
    ...overrides,
  } as StackRelease
}

// replicas/available_replicas are readonly in the generated schema — partial
// literal cast, same pattern as makeStack.
export function makeReleaseDetail(overrides: Partial<StackReleaseDetail> = {}): StackReleaseDetail {
  return {
    ...makeRelease(),
    snapshot: {
      resources: [
        { name: 'web', workload_type: 'Service', source: { image: { ref: 'ghcr.io/acme/orders-api:1.4.2' } } },
        {
          name: 'worker',
          workload_type: 'Service',
          source: {
            git: {
              repo_url: 'https://github.com/acme/orders',
              branch: 'main',
              dockerfile_path: 'Dockerfile',
              build_context: '.',
            },
          },
        },
      ],
      volumes: [],
      connections: [],
      captured_at: '2026-07-30T11:58:00Z',
    },
    live_status: {
      health: 'ok',
      resources: {
        web: { state: 'Ready', replicas: 2, available_replicas: 2 },
        worker: { state: 'Ready', replicas: 1, available_replicas: 1 },
      },
    },
    outcome: {
      resources: {
        web: { phase: 'Ready', replicas: 2, ready_replicas: 2 },
        worker: { phase: 'Ready', replicas: 1, ready_replicas: 1 },
      },
      duration: '75s',
    },
    ...overrides,
  } as StackReleaseDetail
}

export function makeReleaseEvent(overrides: Partial<ReleaseEvent> = {}): ReleaseEvent {
  return {
    id: 'ev-1',
    release_id: 'rel-12',
    sequence: 1,
    occurred_at: '2026-07-30T11:58:12Z',
    source: 'cluster',
    scope: ReleaseEventScope.Resource,
    resource_name: 'web',
    type: ReleaseEventType.ResourceDeploying,
    level: 'info',
    message: 'web: rolling out revision 4',
    ...overrides,
  }
}

// isBuildJobCreated gates the log stream on the BuildJobCreated condition, so
// `status` is merged rather than replaced: a story overriding only `state`
// would otherwise drop the condition and strand the modal on "waiting".
export function makeImageBuild(overrides: Partial<ImageBuild> = {}): ImageBuild {
  const { status, ...rest } = overrides
  return {
    id: 'build-1',
    stack_id: STACK_ID,
    resource_name: 'web',
    ...rest,
    status: {
      state: BuildPhase.Success,
      build_source_revision: 'a1b2c3d4e5f6a7b8',
      conditions: [{ type: BUILD_JOB_CREATED_CONDITION, status: 'True' }],
      ...status,
    },
  } as ImageBuild
}

export function makeCluster(overrides: Partial<Cluster> = {}): Cluster {
  return {
    id: 'c1',
    name: 'prod-us-east',
    cluster_url: 'https://10.0.0.1:6443',
    cluster_ca_data: '',
    cluster_sa_token: '',
    ...overrides,
  }
}
