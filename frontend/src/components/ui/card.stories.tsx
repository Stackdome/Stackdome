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

// D13/D14: content is flat at rest — a hairline border plus the inset --edge
// highlight, never a drop shadow. Reads the token live so this stays honest
// as the palette evolves.
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
    // The --edge var resolves to an inset shadow only — no offset/blur drop shadow component.
    await expect(style.boxShadow).not.toBe('none')
    await expect(style.boxShadow).toContain('inset')
  },
}
