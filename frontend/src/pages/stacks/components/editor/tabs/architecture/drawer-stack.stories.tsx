import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'
import { HardDrive, Settings2 } from 'lucide-react'
import { DrawerStack, type DrawerPanelDescriptor } from './drawer-stack'

const meta = {
  title: 'Features/Canvas/DrawerStack',
  component: DrawerStack,
  tags: ['ai-generated'],
} satisfies Meta<typeof DrawerStack>

export default meta
type Story = StoryObj<typeof meta>

const resourcePanel: DrawerPanelDescriptor = {
  entry: { kind: 'resource', index: 0 },
  title: 'web',
  icon: <Settings2 className="size-[19px]" />,
}

const volumePanel: DrawerPanelDescriptor = {
  entry: { kind: 'volume', name: 'web-data' },
  title: 'web-data',
  icon: <HardDrive className="size-[19px]" />,
}

const front = (
  <>
    <div className="flex items-center gap-2.5 border-b border-border px-4 py-[15px]">
      <HardDrive className="size-[19px] shrink-0 text-brand" />
      <span className="text-base font-medium text-foreground">web-data</span>
    </div>
    <div className="p-4 text-body text-muted-foreground">Front panel body</div>
  </>
)

export const Default: Story = {
  args: {
    panels: [volumePanel],
    front,
    onTruncate: fn(),
    onPop: fn(),
    onCloseAll: fn(),
  },
}

/** Two panels: the back one is header-only, dimmed, and staggered behind the front. */
export const Stacked: Story = {
  args: {
    panels: [resourcePanel, volumePanel],
    front,
    onTruncate: fn(),
    onPop: fn(),
    onCloseAll: fn(),
  },
}
