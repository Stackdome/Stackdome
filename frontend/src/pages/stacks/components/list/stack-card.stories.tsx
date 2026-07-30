import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
import { makeStack } from '../../../../../.storybook/fixtures'
import { ReleaseState } from '@/pages/stacks/components/editor/tabs/deployments/release-states'
import type { Stack } from '@/pages/stacks/types'
import { DeployStackCard } from './stack-card'

const meta = {
  title: 'Features/StackCard',
  component: DeployStackCard,
  tags: ['ai-generated'],
  decorators: [
    (Story) => (
      <div className="max-w-[380px]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof DeployStackCard>

export default meta
type Story = StoryObj<typeof meta>

const released = {
  latest_release: { id: 'r1', state: ReleaseState.Released },
  converged_release: { id: 'r1', state: ReleaseState.Released, health: 'ok' },
} as Partial<Stack>

export const Running: Story = {
  args: { stack: makeStack(released), onDelete: fn() },
}

export const Deploying: Story = {
  args: {
    stack: makeStack({
      latest_release: { id: 'r2', state: ReleaseState.InProgress },
      converged_release: { id: 'r1', state: ReleaseState.Released, health: 'ok' },
    } as Partial<Stack>),
  },
  play: async ({ canvasElement }) => {
    await expect(canvasElement.querySelector('[data-rail]')).toBeTruthy()
  },
}

export const Failed: Story = {
  args: {
    stack: makeStack({
      latest_release: { id: 'r1', state: ReleaseState.Failed },
    } as Partial<Stack>),
  },
}

// Latest deploy failed while a previous release is still live: healthy word
// plus the amber alert triangle, never a doubled error state.
export const DeployFailedWhileLive: Story = {
  args: {
    stack: makeStack({
      latest_release: { id: 'r2', state: ReleaseState.Failed },
      converged_release: { id: 'r1', state: ReleaseState.Released, health: 'ok' },
    } as Partial<Stack>),
  },
  play: async ({ canvas }) => {
    await expect(canvas.getByLabelText('Latest deploy failed')).toBeVisible()
  },
}

export const NotDeployed: Story = {
  args: { stack: makeStack() },
}

export const Deleting: Story = {
  args: { stack: makeStack({ ...released, lifecycle: 'deleting' } as Partial<Stack>), onDelete: fn() },
}

export const LongName: Story = {
  args: {
    stack: makeStack({
      ...released,
      name: 'extremely-long-service-name-that-truncates-in-the-card-header',
    } as Partial<Stack>),
  },
}

export const GitSource: Story = {
  args: {
    stack: makeStack({
      ...released,
      spec: {
        stack_resources: [
          {
            name: 'web',
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
      },
    } as Partial<Stack>),
  },
}

// No onDelete wired → kebab menu absent (viewer without write access).
export const ReadOnly: Story = {
  args: { stack: makeStack(released) },
  play: async ({ canvas }) => {
    await expect(canvas.queryByLabelText(/Actions for/)).toBeNull()
  },
}
