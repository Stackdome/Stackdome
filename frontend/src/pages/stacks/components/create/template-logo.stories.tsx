import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, within } from 'storybook/test'

import { templates } from '@/pages/stacks/data/templates/registry'
import { TemplateLogo } from './template-logo'

const meta = {
  title: 'Features/CreateStack/TemplateLogo',
  component: TemplateLogo,
  parameters: { layout: 'centered' },
  args: { template: templates[0] },
} satisfies Meta<typeof TemplateLogo>

export default meta
type Story = StoryObj<typeof meta>

/**
 * **Every logo, on the chip it actually sits in.**
 *
 * This story exists to be looked at in both themes. A brand mark is authored
 * against one background, and several of these carry a near-black or near-white
 * element — which disappears on exactly one of our two chips. Judging them one
 * at a time, in one theme, is how that ships unnoticed.
 */
export const EveryTemplate: Story = {
  render: () => (
    <div className="bg-card flex flex-wrap gap-4 p-6">
      {templates.map((template) => (
        <div key={template.id} className="flex w-[92px] flex-col items-center gap-2">
          <span className="border-border bg-control flex size-8 items-center justify-center rounded-md border">
            <TemplateLogo template={template} />
          </span>
          <span className="text-label text-fg-muted text-center">{template.name}</span>
        </div>
      ))}
    </div>
  ),
}

/** The 26px rung, as the detail panel draws it. */
export const DetailBadge: Story = {
  render: () => (
    <div className="bg-card flex flex-wrap gap-4 p-6">
      {templates.map((template) => (
        <span
          key={template.id}
          className="border-border bg-control flex size-11 items-center justify-center rounded-lg border"
        >
          <TemplateLogo template={template} size={26} />
        </span>
      ))}
    </div>
  ),
}

/** No art, or art that fails to load — the record's initials carry it, because
 *  a template with a broken asset still has to be pickable. */
export const FallsBackToInitials: Story = {
  args: { template: { ...templates[0], icon: '' } },
  render: (args) => (
    <span className="border-border bg-control flex size-8 items-center justify-center rounded-md border">
      <TemplateLogo {...args} />
    </span>
  ),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(canvas.getByText(templates[0].initials)).toBeInTheDocument()
  },
}

/** The mark is never recoloured (§7) — no filter, no opacity, no currentColor. */
export const NeverTinted: Story = {
  play: async ({ canvasElement }) => {
    const img = canvasElement.querySelector('img')!
    const style = getComputedStyle(img)
    await expect(style.filter).toBe('none')
    await expect(style.opacity).toBe('1')
  },
}
