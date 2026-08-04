import type { Meta, StoryObj } from '@storybook/react-vite'
import { FailureCard } from './failure-card'
import { LogSnapshot } from './log-snapshot'

const meta = {
  title: 'Features/Deployments/FailureCard',
  component: FailureCard,
  tags: ['ai-generated'],
  decorators: [
    (Story) => (
      <div className="max-w-[720px]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof FailureCard>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    resourceName: 'web',
    stage: 'runtime',
    reason: 'CrashLoopBackOff',
    message: 'back-off 5m0s restarting failed container=web pod=web-7d9f8b6c5-x2kqr',
    exitCode: 1,
    restartCount: 4,
    revision: 'a1b2c3d4e5f6',
    conditions: [
      { type: 'Available', reason: 'MinimumReplicasUnavailable', message: 'Deployment does not have minimum availability.' },
      { type: 'Progressing', reason: 'ProgressDeadlineExceeded', message: 'ReplicaSet "web-7d9f8b6c5" has timed out progressing.' },
    ],
    hint: 'The container keeps exiting shortly after start.',
  },
}

export const WithLogSnapshot: Story = {
  args: {
    resourceName: 'worker',
    stage: 'runtime',
    reason: 'OOMKilled',
    exitCode: 137,
    restartCount: 2,
    children: (
      <div className="mt-3">
        <LogSnapshot
          label="Last container output"
          lines={[
            '2026-07-30T11:59:01Z INFO  starting worker pool size=4',
            '2026-07-30T11:59:02Z INFO  connected to queue redis://queue:6379',
            '2026-07-30T11:59:14Z WARN  heap usage above 90%',
            '2026-07-30T11:59:15Z FATAL out of memory',
          ]}
        />
      </div>
    ),
  },
}
