import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
import { makeStack } from '../../../../../.storybook/fixtures'
import { ReleaseState } from '@/pages/stacks/components/editor/tabs/deployments/release-states'
import type { Stack } from '@/api/stack-types'
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

// The whole card navigates: hover is a flat ink wash only (D13 — nothing
// moves), focus is an outline (never a brand-colored box-shadow ring).
export const Running: Story = {
  args: { stack: makeStack(released), onDelete: fn() },
  play: async ({ canvas, canvasElement }) => {
    const card = canvas.getByRole('link')
    await expect(card.className).toContain('hover:bg-foreground/[0.04]')
    await expect(card.className).not.toContain('brand')
    await expect(card.className).toContain('focus-visible:outline-2')
    await expect(card.className).not.toContain('ring-')
    // Leading glyph is neutral ink — brand orange stays reserved for
    // eyebrows, wires, and the mark (rubric #3), never a per-card icon.
    await expect(canvasElement.querySelector('svg.text-brand')).toBeNull()
  },
}

// Building/progress reads amber, never brand orange (rubric #4 — hue
// reports state; the rail-sweep comment above literally says "amber").
export const Deploying: Story = {
  args: {
    stack: makeStack({
      latest_release: { id: 'r2', state: ReleaseState.InProgress },
      converged_release: { id: 'r1', state: ReleaseState.Released, health: 'ok' },
    } as Partial<Stack>),
  },
  play: async ({ canvas, canvasElement }) => {
    const rail = canvasElement.querySelector('[data-rail]')
    await expect(rail).toBeTruthy()
    await expect(rail?.querySelector('.bg-warn')).toBeTruthy()
    await expect(canvasElement.querySelector('.bg-brand')).toBeNull()
    await expect(canvas.getByText('progressing').className).toContain('text-warn')
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
