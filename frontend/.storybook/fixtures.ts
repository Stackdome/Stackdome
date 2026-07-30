import type { components } from '../src/api/types/openapi'

type Schemas = components['schemas']
export type User = Schemas['User']
export type Project = Schemas['Project']
export type Stack = Schemas['Stack']

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
