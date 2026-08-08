import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
import { ReactFlowProvider } from '@xyflow/react'
import { SidebarProvider } from '@/components/ui/sidebar'
import { CanvasControls } from './canvas-controls'

const meta = {
  title: 'Features/Canvas/CanvasControls',
  component: CanvasControls,
  tags: ['ai-generated'],
  // Zoom actions need a flow store; the zen toggle needs a sidebar to collapse.
  decorators: [
    (Story) => (
      <SidebarProvider>
        <ReactFlowProvider>
          <div className="relative h-40 w-64">
            <Story />
          </div>
        </ReactFlowProvider>
      </SidebarProvider>
    ),
  ],
  args: {
    showConnections: true,
    onToggleConnections: fn(),
    onAutoLayout: fn(),
  },
} satisfies Meta<typeof CanvasControls>

export default meta
type Story = StoryObj<typeof meta>

/** Zoom pill, layout pill, connections toggle — control-fill chrome, flat at rest. */
export const Default: Story = {}

export const ConnectionsHidden: Story = {
  args: { showConnections: false },
  play: async ({ canvas }) => {
    await expect(canvas.getByRole('button', { name: 'Show connections' })).toHaveAttribute(
      'aria-pressed',
      'false',
    )
  },
}
