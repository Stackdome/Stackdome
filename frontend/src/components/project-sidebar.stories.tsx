import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn } from 'storybook/test'
import { Users, Settings } from 'lucide-react'
import { SidebarProvider } from '@/components/ui/sidebar'
import { ProjectSidebar, type SidebarSection } from './project-sidebar'

const sections: SidebarSection[] = [
  {
    label: 'Projects',
    icon: <Settings className="size-4" />,
    addHref: '/settings/projects/new',
    items: [
      { label: 'default', icon: <Users className="size-4" />, href: '/settings/projects/default', active: true },
      { label: 'platform', icon: <Users className="size-4" />, href: '/settings/projects/platform' },
      { label: 'growth', icon: <Users className="size-4" />, href: '/settings/projects/growth', onClick: fn() },
    ],
  },
]

const meta = {
  title: 'Features/ProjectSidebar',
  component: ProjectSidebar,
  tags: ['ai-generated'],
  decorators: [
    (Story) => (
      <SidebarProvider>
        <div className="flex h-[480px] w-[280px]">
          <Story />
        </div>
      </SidebarProvider>
    ),
  ],
} satisfies Meta<typeof ProjectSidebar>

export default meta
type Story = StoryObj<typeof meta>

// Active item reads via an ink tint (data-active on the shared sidebar
// primitive), never brand orange text (rubric #3, #4).
export const Default: Story = {
  args: { sections },
  play: async ({ canvasElement }) => {
    const activeLink = canvasElement.querySelector('a[data-active="true"]')
    await expect(activeLink).toBeTruthy()
    await expect(activeLink?.className).not.toContain('text-primary')
    await expect(activeLink?.className).not.toContain('bg-primary')

    // Brand mark tile is the one deliberate use of brand color (the mark),
    // not a generic per-instance icon.
    const mark = canvasElement.querySelector('.border-brand-border')
    await expect(mark).toBeTruthy()
  },
}

export const NoActiveItem: Story = {
  args: {
    sections: [
      {
        ...sections[0],
        items: sections[0].items.map((i) => ({ ...i, active: false })),
      },
    ],
  },
}

export const EmptySection: Story = {
  args: {
    sections: [{ label: 'Projects', icon: <Settings className="size-4" />, items: [] }],
  },
}

export const MultipleSections: Story = {
  args: {
    sections: [
      ...sections,
      {
        label: 'Members',
        icon: <Users className="size-4" />,
        items: [
          { label: 'Ada Lovelace', href: '/settings/users/ada' },
          { label: 'Grace Hopper', href: '/settings/users/grace' },
        ],
      },
    ],
  },
}
