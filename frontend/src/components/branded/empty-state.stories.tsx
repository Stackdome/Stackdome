import type { Meta, StoryObj } from '@storybook/react-vite'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { EmptyState, SearchGlyph, StackArchitectureGlyph } from './empty-state'

const meta = {
  title: 'Branded/EmptyState',
  component: EmptyState,
  args: { title: 'No stacks yet' },
} satisfies Meta<typeof EmptyState>

export default meta
type Story = StoryObj<typeof meta>

export const Bare: Story = {}

/**
 * First run — the one screen that gets to define the product's core noun, and
 * the only one that earns the decorated glyph.
 */
export const FirstRun: Story = {
  args: {
    className: 'gap-6',
    icon: <StackArchitectureGlyph />,
    title: 'No stacks yet',
    description:
      'A stack is your app and everything it needs to run — services, databases and domains, deployed together from a Git branch.',
    action: (
      <Button>
        <Plus />
        New stack
      </Button>
    ),
  },
}

/** A filter that matched nothing: small mark, and a way back. */
export const NoResults: Story = {
  args: {
    icon: <SearchGlyph />,
    title: 'No stacks match',
    description: 'Try a different search, or clear the filters.',
    action: <Button variant="secondary">Clear filters</Button>,
  },
}

/** Read-only members get the explanation without an action they cannot take. */
export const NoAction: Story = {
  args: {
    icon: <SearchGlyph />,
    title: 'No stacks match',
    description: 'Try a different search, or clear the filters.',
  },
}

/** Long copy must stay centred and readable rather than running the full width. */
export const LongDescription: Story = {
  args: {
    className: 'gap-6',
    icon: <StackArchitectureGlyph />,
    title: 'No stacks yet',
    description:
      'A stack is your app and everything it needs to run — services, databases, object stores, domains and the secrets they read — deployed together from a single Git branch, to whichever cluster you point it at.',
    action: <Button>New stack</Button>,
  },
}
