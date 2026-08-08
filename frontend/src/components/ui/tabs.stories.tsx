import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { Tabs, TabsContent, TabsList, TabsTrigger } from './tabs'

const meta = {
  title: 'Primitives/Tabs',
  component: Tabs,
  tags: ['ai-generated'],
} satisfies Meta<typeof Tabs>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: () => (
    <Tabs defaultValue="overview" className="w-[420px]">
      <TabsList>
        <TabsTrigger value="overview">Overview</TabsTrigger>
        <TabsTrigger value="logs">Logs</TabsTrigger>
        <TabsTrigger value="settings">Settings</TabsTrigger>
      </TabsList>
      <TabsContent value="overview" className="text-sm text-fg-2">
        3 resources running, all healthy.
      </TabsContent>
      <TabsContent value="logs" className="text-sm text-fg-2">
        No log lines in the selected window.
      </TabsContent>
      <TabsContent value="settings" className="text-sm text-fg-2">
        Environment variables, secrets, and scaling.
      </TabsContent>
    </Tabs>
  ),
}

export const Disabled: Story = {
  render: () => (
    <Tabs defaultValue="overview" className="w-[420px]">
      <TabsList>
        <TabsTrigger value="overview">Overview</TabsTrigger>
        <TabsTrigger value="logs" disabled>Logs</TabsTrigger>
      </TabsList>
      <TabsContent value="overview" className="text-sm text-fg-2">
        Overview content.
      </TabsContent>
      <TabsContent value="logs">Unreachable.</TabsContent>
    </Tabs>
  ),
}

// D13/D14 + no filled brand pills: the active tab is an ink wash, not a
// brand/orange fill, and nothing scales or translates on selection.
export const ActiveStateIsInkNotBrand: Story = {
  render: () => (
    <Tabs defaultValue="overview" className="w-[420px]">
      <TabsList>
        <TabsTrigger value="overview">Overview</TabsTrigger>
        <TabsTrigger value="logs">Logs</TabsTrigger>
      </TabsList>
      <TabsContent value="overview">Overview content.</TabsContent>
      <TabsContent value="logs">Logs content.</TabsContent>
    </Tabs>
  ),
  play: async ({ canvas }) => {
    const active = canvas.getByRole('tab', { name: 'Overview' })
    await expect(active).toHaveAttribute('data-state', 'active')
    const style = getComputedStyle(active)
    const probe = document.createElement('div')
    probe.style.color = getComputedStyle(document.documentElement).getPropertyValue('--brand').trim()
    document.body.appendChild(probe)
    const brandColor = getComputedStyle(probe).color
    probe.remove()
    await expect(style.backgroundColor).not.toBe(brandColor)
  },
}
