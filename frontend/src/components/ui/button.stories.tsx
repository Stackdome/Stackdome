import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { Plus } from 'lucide-react'
import { Button } from './button'

const meta = {
  component: Button,
  tags: ['ai-generated'],
  args: { children: 'Deploy stack' },
} satisfies Meta<typeof Button>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const Destructive: Story = { args: { variant: 'destructive', children: 'Delete cluster' } }
export const Outline: Story = { args: { variant: 'outline' } }
export const Secondary: Story = { args: { variant: 'secondary' } }
export const Ghost: Story = { args: { variant: 'ghost' } }
export const Link: Story = { args: { variant: 'link' } }
export const Inverse: Story = { args: { variant: 'inverse' } }
export const Mono: Story = { args: { variant: 'mono', children: 'Open console' } }
export const RailPrimary: Story = { args: { variant: 'railPrimary', size: 'rail', children: 'Save' } }
export const Disabled: Story = {
  args: { disabled: true },
  play: async ({ canvas }) => {
    await expect(canvas.getByRole('button', { name: /deploy stack/i })).toBeDisabled()
  },
}
export const WithIcon: Story = {
  args: {
    size: 'sm',
    children: (
      <>
        <Plus /> Add addon
      </>
    ),
  },
}

// Proves the app's Tailwind theme actually loaded in the preview: the default
// variant is bg-brand, which resolves to the --brand token from src/index.css.
export const CssCheck: Story = {
  play: async ({ canvas }) => {
    const button = canvas.getByRole('button', { name: /deploy stack/i })
    await expect(getComputedStyle(button).backgroundColor).toBe('oklch(0.72 0.2 40)')
  },
}
