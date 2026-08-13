import type { Meta, StoryObj } from '@storybook/react-vite'
import { EventRow } from './event-row'

const meta = {
  title: 'Features/Deployments/EventRow',
  component: EventRow,
  tags: ['ai-generated'],
  decorators: [
    (Story) => (
      <div className="max-w-[640px] rounded-lg border border-border">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof EventRow>

export default meta
type Story = StoryObj<typeof meta>

export const Deploy: Story = {
  args: {
    ok: true,
    kind: 'deploy',
    title: 'Release #12 deployed',
    sub: 'git a1b2c3d · 2 resources updated',
    when: '11:59',
    duration: '1m 15s',
  },
}

export const Build: Story = {
  args: {
    ok: false,
    kind: 'build',
    title: 'Build failed for web',
    sub: 'Dockerfile step 7/9: npm run build exited 1',
    when: '11:42',
    duration: '3m 02s',
  },
}

export const Rollback: Story = {
  args: {
    ok: true,
    kind: 'rollback',
    title: 'Rolled back to release #11',
    when: 'yesterday 17:44',
  },
}
