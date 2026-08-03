import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
import {
  DEFAULT_PROJECT,
  ORG_ID,
  STACK_ID,
  makeRelease,
  makeReleaseDetail,
  makeReleaseEvent,
  makeStack,
} from '../../../../../../../.storybook/fixtures'
import { baselineHandlers, releaseHandlers } from '../../../../../../../.storybook/msw-handlers'
import { withReleaseDetail } from '../../../../../../../.storybook/decorators'
import { ReleaseEventType } from '@/api/releases'
import type { Stack } from '@/api/stacks'
import { ReleaseState } from './release-states'
import type { DeployLifecycle } from './use-deploy-lifecycle'
import type { SnapshotDiff } from './release-snapshot-diff'
import { DeploymentsTab } from './deployments-tab'

const releases = [
  makeRelease({ id: 'rel-13', sequence: 13 }),
  makeRelease({ id: 'rel-12', sequence: 12, state: ReleaseState.Superseded }),
  makeRelease({ id: 'rel-11', sequence: 11 }),
]

const stack = makeStack({
  converged_release: { id: 'rel-13', sequence: 13, state: ReleaseState.Released, health: 'ok' },
  latest_release: { id: 'rel-13', sequence: 13, state: ReleaseState.Released },
} as Partial<Stack>)

const details = {
  'rel-13': makeReleaseDetail({ id: 'rel-13', sequence: 13 }),
  'rel-12': makeReleaseDetail({ id: 'rel-12', sequence: 12, state: ReleaseState.Superseded, live_status: undefined }),
  'rel-11': makeReleaseDetail({ id: 'rel-11', sequence: 11, live_status: undefined }),
}

const events = [
  makeReleaseEvent({ release_id: 'rel-13', sequence: 1 }),
  makeReleaseEvent({
    release_id: 'rel-13',
    sequence: 2,
    type: ReleaseEventType.ResourceReady,
    level: 'success',
    message: 'web: Ready',
    occurred_at: '2026-07-30T11:59:20Z',
  }),
]

const clean: DeployLifecycle = { phase: 'clean', nextSeq: 14 }

const stagedDiff: SnapshotDiff = {
  resources: [
    {
      name: 'web',
      change: 'modified',
      sections: [
        {
          kind: 'configuration',
          rows: [
            { key: 'image', from: 'ghcr.io/acme/orders-api:1.4.2', to: 'ghcr.io/acme/orders-api:1.5.0', kind: 'changed' },
          ],
        },
      ],
    },
  ],
  volumes: [],
}

const meta = {
  title: 'Features/Deployments/DeploymentsTab',
  component: DeploymentsTab,
  tags: ['ai-generated'],
  decorators: [withReleaseDetail],
  args: {
    orgId: ORG_ID,
    projectName: DEFAULT_PROJECT,
    stackId: STACK_ID,
    stack,
    releases,
    activeRelease: releases[0],
    loading: false,
    error: null,
    lifecycle: clean,
    refetchReleases: fn(),
    onRollback: fn(),
    onCancel: fn(),
    onCopyId: fn(),
  },
  parameters: {
    msw: [...releaseHandlers(details, events), ...baselineHandlers],
  },
} satisfies Meta<typeof DeploymentsTab>

export default meta
type Story = StoryObj<typeof meta>

// Live release open at the top of the rail; detail + events resolve through MSW.
export const Default: Story = {
  play: async ({ canvas }) => {
    await expect(canvas.getByText('Deploy timeline')).toBeInTheDocument()
    await expect(await canvas.findByText('all resources', undefined, { timeout: 4000 })).toBeInTheDocument()
  },
}

export const Empty: Story = {
  args: {
    stack: makeStack(),
    releases: [],
    activeRelease: undefined,
    lifecycle: { phase: 'clean', nextSeq: 1 },
  },
}

// First deploy failed on a never-converged stack: post-mortem with the red banner.
export const DeployFailed: Story = {
  args: {
    stack: makeStack(),
    releases: [
      makeRelease({
        id: 'rel-1',
        sequence: 1,
        state: ReleaseState.Failed,
        message: 'timed out waiting for resource "web" to become ready',
      }),
    ],
    activeRelease: undefined,
    lifecycle: { phase: 'clean', nextSeq: 2 },
  },
  parameters: {
    msw: [
      ...releaseHandlers(
        {
          'rel-1': makeReleaseDetail({
            id: 'rel-1',
            sequence: 1,
            state: ReleaseState.Failed,
            live_status: undefined,
            outcome: {
              resources: { web: { phase: 'Failed', replicas: 2, ready_replicas: 0, message: 'CrashLoopBackOff' } },
            },
          }),
        },
        [
          makeReleaseEvent({
            release_id: 'rel-1',
            sequence: 1,
            type: ReleaseEventType.ResourceFailed,
            level: 'error',
            message: 'web: back-off restarting failed container',
          }),
        ],
      ),
      ...baselineHandlers,
    ],
  },
  // Banner renders only after the failed node's detail fetch resolves.
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByText('#1'))
    await expect(await canvas.findByText('Deploy failed', undefined, { timeout: 4000 })).toBeInTheDocument()
  },
}

// Saved-but-undeployed changes lead the rail as a dashed draft node.
export const WithDraft: Story = {
  args: {
    lifecycle: { phase: 'staged', stagedDiff, vsSeq: 13, nextSeq: 14 },
  },
}
