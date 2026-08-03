import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { ThemeToggle } from './theme-toggle'

const meta = {
  title: 'Features/ThemeToggle',
  component: ThemeToggle,
  tags: ['ai-generated'],
} satisfies Meta<typeof ThemeToggle>

export default meta
type Story = StoryObj<typeof meta>

// A floating icon control (topbar, 404 page) reads as a physical control on
// `--control` fill rather than the transparent `ghost` material used for
// in-row/toolbar icon buttons — asserted against the parsed stylesheet per
// the plate rubric (never a synthetic-hover check).
export const Default: Story = {
  play: async ({ canvas }) => {
    const button = canvas.getByRole('button', { name: /toggle theme/i })
    const style = getComputedStyle(button)
    expect(style.backgroundColor).not.toBe('rgba(0, 0, 0, 0)')
    expect(style.backgroundColor).not.toBe('transparent')
  },
}

// The auth header is a nav-equivalent row that already has its own chrome
// (the room + plate), so the toggle sits flat at rest like the website's
// `.auth-top-right .icon-btn` — no fill, no border, no shadow.
export const OnAuthHeader: Story = {
  args: { variant: 'ghost' },
  play: async ({ canvas }) => {
    const button = canvas.getByRole('button', { name: /toggle theme/i })
    const style = getComputedStyle(button)
    expect(style.backgroundColor).toBe('rgba(0, 0, 0, 0)')
    expect(style.boxShadow).toBe('none')
  },
}
