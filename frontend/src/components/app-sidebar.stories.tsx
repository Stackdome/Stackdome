import { useEffect, useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, within } from 'storybook/test'
import { useNavigate } from 'react-router-dom'
import { withCurrentUser } from '../../.storybook/decorators'
import { AppSidebar } from './app-sidebar'
import { SidebarProvider } from './ui/sidebar'

// The global preview decorator already supplies a MemoryRouter (nesting a
// second one throws); hop that router to /stacks so NavStacks' isActive
// check resolves before rendering.
function SidebarHarness({ collapsed = false }: { collapsed?: boolean }) {
  const navigate = useNavigate()
  const [ready, setReady] = useState(false)
  useEffect(() => {
    navigate('/stacks', { replace: true })
    setReady(true)
  }, [navigate])
  if (!ready) return null
  return (
    <SidebarProvider defaultOpen={!collapsed}>
      <div className="flex h-[560px]">
        <AppSidebar />
      </div>
    </SidebarProvider>
  )
}

const meta = {
  title: 'Features/AppSidebar',
  component: SidebarHarness,
  tags: ['ai-generated'],
  decorators: [withCurrentUser],
} satisfies Meta<typeof SidebarHarness>

export default meta
type Story = StoryObj<typeof meta>

// Active route item (Stacks, current path) reads as an ink tint — never
// brand orange text or icon fill.
export const Expanded: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const active = await canvas.findByRole('link', { name: /stacks/i })
    expect(active.className).not.toContain('brand')
    const style = getComputedStyle(active)
    expect(style.color).not.toContain('255, 96, 7')
  },
}

export const Collapsed: Story = {
  args: { collapsed: true },
}
