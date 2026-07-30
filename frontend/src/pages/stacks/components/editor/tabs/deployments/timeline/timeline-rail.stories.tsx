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
} from '../../../../../../../../.storybook/fixtures'
import { baselineHandlers, releaseHandlers } from '../../../../../../../../.storybook/msw-handlers'
import { withReleaseDetail } from '../../../../../../../../.storybook/decorators'
import { ReleaseEventType } from '@/api/releases'
import type { Stack } from '@/api/stacks'
import { ReleaseState } from '../release-states'
import { TimelineRail } from './timeline-rail'

const releases = [
  makeRelease({ id: 'rel-13', sequence: 13 }),
  makeRelease({
    id: 'rel-12',
    sequence: 12,
    state: ReleaseState.Failed,
    message: 'timed out waiting for resource "web" to become ready',
  }),
  makeRelease({ id: 'rel-11', sequence: 11 }),
]

const stack = makeStack({
  converged_release: { id: 'rel-13', sequence: 13, state: ReleaseState.Released, health: 'ok' },
  latest_release: { id: 'rel-13', sequence: 13, state: ReleaseState.Released },
} as Partial<Stack>)

const details = {
  'rel-13': makeReleaseDetail({ id: 'rel-13', sequence: 13 }),
  'rel-12': makeReleaseDetail({
    id: 'rel-12',
    sequence: 12,
    state: ReleaseState.Failed,
    live_status: undefined,
    snapshot: {
      resources: [
        { name: 'web', workload_type: 'Service', source: { image: { ref: 'ghcr.io/acme/orders-api:1.4.1' } } },
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
    },
  }),
  'rel-11': makeReleaseDetail({ id: 'rel-11', sequence: 11 }),
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

const meta = {
  title: 'Features/Deployments/TimelineRail',
  component: TimelineRail,
  tags: ['ai-generated'],
  decorators: [
    withReleaseDetail,
    (Story) => (
      <div className="max-w-[960px]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof TimelineRail>

export default meta
type Story = StoryObj<typeof meta>

// Live release open at the head; its detail + events load through MSW via the
// real ReleaseDetailProvider, so the split console fills in asynchronously.
export const Default: Story = {
  args: {
    releases,
    activeRelease: releases[0],
    stack,
    logContext: { orgId: ORG_ID, projectName: DEFAULT_PROJECT, stackId: STACK_ID },
    onRollback: fn(),
    onCancel: fn(),
    onCopyId: fn(),
  },
  parameters: {
    msw: [...releaseHandlers(details, events), ...baselineHandlers],
  },
  play: async ({ canvas }) => {
    await expect(await canvas.findByText('all resources', undefined, { timeout: 4000 })).toBeInTheDocument()
    await expect(await canvas.findByText('Deploying — rolling out revision 4', undefined, { timeout: 4000 })).toBeInTheDocument()
  },
}
