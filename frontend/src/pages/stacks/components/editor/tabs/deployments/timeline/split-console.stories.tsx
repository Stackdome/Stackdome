import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { makeReleaseEvent } from '../../../../../../../../.storybook/fixtures'
import { ReleaseEventType } from '@/api/releases'
import { ResourceFailureType } from '../derive'
import { SplitConsole } from './split-console'

const meta = {
  title: 'Features/Deployments/SplitConsole',
  component: SplitConsole,
  tags: ['ai-generated'],
  decorators: [
    (Story) => (
      <div className="max-w-[960px]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof SplitConsole>

export default meta
type Story = StoryObj<typeof meta>

const rows = [
  { name: 'web', phase: 'Ready', replicas: undefined, msg: undefined },
  { name: 'worker', phase: 'Progressing', replicas: '1/2' },
  {
    name: 'metrics',
    phase: 'Failed',
    msg: 'container keeps restarting',
    failure: {
      name: 'metrics',
      type: ResourceFailureType.Runtime,
      stage: 'runtime' as const,
      reason: 'CrashLoopBackOff',
      message: 'back-off 5m0s restarting failed container',
      exitCode: 1,
      restartCount: 4,
    },
  },
]

const events = [
  makeReleaseEvent({ sequence: 1, message: 'web: rolling out revision 4' }),
  makeReleaseEvent({
    sequence: 2,
    resource_name: 'worker',
    message: 'worker: rolling out revision 4',
    occurred_at: '2026-07-30T11:58:14Z',
  }),
  makeReleaseEvent({
    sequence: 3,
    type: ReleaseEventType.ResourceReady,
    level: 'success',
    message: 'web: Ready',
    occurred_at: '2026-07-30T11:58:40Z',
  }),
  makeReleaseEvent({
    sequence: 4,
    resource_name: 'metrics',
    type: ReleaseEventType.ResourceFailed,
    level: 'error',
    message: 'metrics: back-off restarting failed container',
    occurred_at: '2026-07-30T11:59:02Z',
  }),
]

// Pure mode (no logContext): selecting a resource pins its detail and filters
// the activity feed — the core interaction this console exists for.
export const Default: Story = {
  args: { rows, events, streaming: true },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole('button', { name: /metrics/ }))
    await expect(canvas.getByText('· metrics')).toBeInTheDocument()
    await expect(canvas.getByText('back-off 5m0s restarting failed container')).toBeInTheDocument()
    await expect(canvas.queryByText('Deploying — rolling out revision 4')).toBeNull()
  },
}

export const LongLines: Story = {
  args: {
    rows: [
      { name: 'extremely-long-resource-name-that-truncates', phase: 'Progressing' },
      { name: 'web', phase: 'Ready' },
    ],
    events: [
      makeReleaseEvent({
        sequence: 1,
        resource_name: 'extremely-long-resource-name-that-truncates',
        message:
          'extremely-long-resource-name-that-truncates: waiting for rollout to finish: 1 old replicas are pending termination, 2 of 3 updated replicas are available, readiness probe failed: HTTP probe failed with statuscode 503 on /healthz after exceeding the configured initialDelaySeconds and periodSeconds thresholds',
      }),
      makeReleaseEvent({
        sequence: 2,
        message:
          'web: pulling image ghcr.io/acme/orders-api@sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08 from registry with credential set acme-registry-readonly',
      }),
    ],
    streaming: false,
  },
}
