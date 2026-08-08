import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { Button } from './button'
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from './card'

const meta = {
  title: 'Primitives/Card',
  component: Card,
  tags: ['ai-generated'],
} satisfies Meta<typeof Card>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: () => (
    <Card className="w-[360px]">
      <CardHeader>
        <CardTitle>orders-api</CardTitle>
        <CardDescription>prod-cluster · eu-west-1</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-fg-2">Last deployed 4 minutes ago from main@a1b2c3d.</p>
      </CardContent>
      <CardFooter>
        <Button size="sm">Open</Button>
      </CardFooter>
    </Card>
  ),
}

export const WithAction: Story = {
  render: () => (
    <Card className="w-[360px]">
      <CardHeader>
        <CardTitle>web</CardTitle>
        <CardDescription>3 replicas · healthy</CardDescription>
        <CardAction>
          <Button variant="ghost" size="sm">Restart</Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-fg-2">No incidents in the last 24 hours.</p>
      </CardContent>
    </Card>
  ),
}

// §7 — a card is white fill, a 1px hairline, `rounded-lg`, and NO shadow.
// Not even the inset `--edge` highlight this used to assert: that is a bevel,
// and §6 rules bevels out at rest for the same reason. Content is flat; shadow
// is for overlays only.
export const FlatAtRest: Story = {
  render: () => (
    <Card className="w-[360px]">
      <CardHeader>
        <CardTitle>Flat by default</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-fg-2">No drop shadow on content surfaces.</p>
      </CardContent>
    </Card>
  ),
  play: async ({ canvas }) => {
    const card = canvas.getByText('Flat by default').closest('[data-slot="card"]') as HTMLElement
    const style = getComputedStyle(card)
    await expect(style.boxShadow).toBe('none')
    // The hairline is the only edge, and 12px is the card rung of §2's ladder.
    await expect(parseFloat(style.borderTopWidth)).toBe(1)
    await expect(parseFloat(style.borderRadius)).toBe(12)
    // White, same as the sheet — grey never means "a card" (§1).
    const probe = document.createElement('div')
    probe.className = 'bg-card'
    document.body.appendChild(probe)
    const sheet = getComputedStyle(probe).backgroundColor
    probe.remove()
    await expect(style.backgroundColor).toBe(sheet)
  },
}
