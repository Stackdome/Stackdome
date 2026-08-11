import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect } from 'storybook/test'
import { within } from 'storybook/test'
import { EndpointPills } from './entity-card'

const meta = {
  title: 'Branded/EntityCard/EndpointPills',
  component: EndpointPills,
  tags: ['ai-generated'],
} satisfies Meta<typeof EndpointPills>

export default meta
type Story = StoryObj<typeof meta>

// Focus is an outline, never a box-shadow ring (rubric #10); it never turns
// brand orange on interaction (rubric #3).
export const TwoPills: Story = {
  args: {
    urls: [
      { resource: 'web', url: 'https://web.example.com' },
      { resource: 'api', url: 'https://api.example.com' },
    ],
  },
  play: async ({ canvas }) => {
    const pill = canvas.getAllByRole('link')[0]
    await expect(pill.className).toContain('focus-visible:outline-2')
    await expect(pill.className).not.toContain('ring-')
  },
}

// Past two pills the rest collapse into a "+N" popover.
export const Overflow: Story = {
  args: {
    urls: [
      { resource: 'web', url: 'https://web.example.com' },
      { resource: 'api', url: 'https://api.example.com' },
      { resource: 'admin', url: 'https://admin.example.com' },
      { resource: 'docs', url: 'https://docs.example.com' },
    ],
  },
  play: async ({ canvas, canvasElement, userEvent }) => {
    await userEvent.click(canvas.getByRole('button', { name: /2 more endpoints/i }))
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByText('admin')).toBeInTheDocument()
    await expect(await body.findByText('docs')).toBeInTheDocument()
  },
}

export const UrllessEntriesFiltered: Story = {
  args: {
    urls: [{ resource: 'web', url: 'https://web.example.com' }, { resource: 'worker' }],
  },
}
