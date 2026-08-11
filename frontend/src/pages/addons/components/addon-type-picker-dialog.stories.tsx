import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn, within } from 'storybook/test'
import { AddonTypePickerDialog } from './addon-type-picker-dialog'

// Zero story coverage existed before this pass. Added because the card
// hover state changed here (dropped `hover:border-brand-border` — rubric
// D3: orange is a voice/wire/mark, never a UI hover accent).
const meta = {
  title: 'Features/Addons/AddonTypePickerDialog',
  component: AddonTypePickerDialog,
  tags: ['ai-generated'],
  args: {
    open: true,
    onOpenChange: fn(),
    onSelect: fn(),
  },
} satisfies Meta<typeof AddonTypePickerDialog>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  play: async ({ canvasElement }) => {
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByRole('dialog')).toBeInTheDocument()
    await expect(body.getByText('PostgreSQL')).toBeInTheDocument()
    // Redis and Ollama are not yet available — rendered as disabled "Soon" cards.
    await expect(body.getAllByText('Soon')).toHaveLength(2)
  },
}

export const SelectAvailableType: Story = {
  play: async ({ canvasElement, userEvent, args }) => {
    const body = within(canvasElement.ownerDocument.body)
    const postgresCard = await body.findByText('PostgreSQL')
    await userEvent.click(postgresCard)
    await expect(args.onSelect).toHaveBeenCalledWith('postgres')
  },
}

// Card hover is a flat ink fill shift only — no brand-orange border accent,
// no shadow, no movement (rubric D3/D5/D6).
export const CardHoverIsNeutral: Story = {
  play: async ({ canvasElement, userEvent }) => {
    const body = within(canvasElement.ownerDocument.body)
    const postgresCard = await body.findByText('PostgreSQL')
    const card = postgresCard.closest('button') as HTMLElement
    const brand = getComputedStyle(document.documentElement).getPropertyValue('--brand').trim()
    const probe = document.createElement('div')
    probe.style.borderColor = brand
    document.body.appendChild(probe)
    const brandBorder = getComputedStyle(probe).borderColor
    probe.remove()

    const transformBefore = getComputedStyle(card).transform
    await userEvent.hover(card)
    await expect(getComputedStyle(card).borderColor).not.toBe(brandBorder)
    await expect(getComputedStyle(card).transform).toBe(transformBefore)
  },
}

// Unavailable options ("Soon") are inert — no click affordance, not focusable
// via the disabled attribute.
export const UnavailableTypeIsDisabled: Story = {
  play: async ({ canvasElement }) => {
    const body = within(canvasElement.ownerDocument.body)
    const redisCard = (await body.findByText('Redis')).closest('button') as HTMLButtonElement
    await expect(redisCard).toBeDisabled()
  },
}
