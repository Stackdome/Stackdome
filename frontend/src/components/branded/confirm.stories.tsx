import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, waitFor, within } from 'storybook/test'
import { Button } from '@/components/ui/button'
import { ConfirmProvider, useConfirm, type ConfirmOptions } from './confirm'

/** Opens the dialog on mount-click so a play function has something to drive. */
function Trigger({ opts, label = 'Delete' }: { opts: ConfirmOptions; label?: string }) {
  const confirm = useConfirm()
  return (
    <Button variant="destructive" shape="flat" onClick={() => void confirm(opts)}>
      {label}
    </Button>
  )
}

const meta = {
  title: 'Branded/ConfirmDialog',
  component: Trigger,
  decorators: [
    (Story) => (
      <ConfirmProvider>
        <Story />
      </ConfirmProvider>
    ),
  ],
  args: { opts: { title: 'Delete?' } },
} satisfies Meta<typeof Trigger>

export default meta
type Story = StoryObj<typeof meta>

/** §6a level 1 — reversible or cheap. Red button, live immediately. */
export const Level1Confirm: Story = {
  args: {
    opts: {
      title: 'Remove this domain?',
      description: 'The domain stops routing to this stack. You can add it back at any time.',
      confirmLabel: 'Remove',
      variant: 'destructive',
    },
  },
  play: async ({ canvas, canvasElement, userEvent }) => {
    await userEvent.click(canvas.getByRole('button', { name: 'Delete' }))
    const body = within(canvasElement.ownerDocument.body)
    const commit = await body.findByRole('button', { name: 'Remove' })
    // No gate: live on arrival.
    await expect(commit).toBeEnabled()
  },
}

/** §6a level 2 — destroys something rebuildable. A checkbox must be ticked
 *  before the red button goes live. */
export const Level2Acknowledge: Story = {
  args: {
    opts: {
      title: 'Delete this API key?',
      description: 'All requests using this key will start failing as soon as it is deleted.',
      confirmLabel: 'Delete key',
      variant: 'destructive',
      gate: { kind: 'acknowledge', label: 'I understand that services using this key will break.' },
    },
  },
  play: async ({ canvas, canvasElement, userEvent }) => {
    await userEvent.click(canvas.getByRole('button', { name: 'Delete' }))
    const body = within(canvasElement.ownerDocument.body)
    const commit = await body.findByRole('button', { name: 'Delete key' })

    // Rendered DISABLED, not hidden — the cost must be visible before it is
    // payable, and the shape must survive being switched off. waitFor because
    // Radix fades the panel in; an immediate read catches opacity mid-flight.
    await waitFor(async () => {
      await expect(commit).toBeVisible()
    })
    await expect(commit).toBeDisabled()

    await userEvent.click(body.getByRole('checkbox'))
    await expect(commit).toBeEnabled()
  },
}

/** §6a level 3 — has dependents or data. The name has to be typed. */
export const Level3Retype: Story = {
  args: {
    opts: {
      title: 'Delete payments-gateway?',
      description:
        'Every service in this stack stops and its containers, volumes and routes are torn down. Any traffic still pointed at it starts failing immediately.',
      confirmLabel: 'Delete stack',
      variant: 'destructive',
      gate: { kind: 'retype', name: 'payments-gateway' },
    },
  },
  play: async ({ canvas, canvasElement, userEvent }) => {
    await userEvent.click(canvas.getByRole('button', { name: 'Delete' }))
    const body = within(canvasElement.ownerDocument.body)
    const commit = await body.findByRole('button', { name: 'Delete stack' })
    const field = body.getByLabelText(/type .* to confirm/i)

    await expect(commit).toBeDisabled()

    // A near miss is still a miss.
    await userEvent.type(field, 'payments-gatewa')
    await expect(commit).toBeDisabled()

    await userEvent.type(field, 'y')
    await expect(commit).toBeEnabled()
  },
}

/** The gate never survives into the next dialog — a ticked box from the last
 *  thing you deleted must not pre-arm the next one. */
export const GateResetsBetweenDialogs: Story = {
  args: {
    opts: {
      title: 'Delete this preview environment?',
      description: 'The environment and everything deployed into it are torn down.',
      confirmLabel: 'Delete env',
      variant: 'destructive',
      gate: { kind: 'acknowledge', label: 'I understand this environment will be destroyed.' },
    },
  },
  play: async ({ canvas, canvasElement, userEvent }) => {
    const body = within(canvasElement.ownerDocument.body)
    const open = canvas.getByRole('button', { name: 'Delete' })

    // The dialog fades in, so settle on the panel before driving it. The
    // checkbox input is deliberately transparent — the styled box is the paint
    // — so it is the commit button that reports the panel has arrived.
    const settled = async () => {
      await waitFor(async () => {
        await expect(await body.findByRole('button', { name: 'Delete env' })).toBeVisible()
      })
      return body.getByRole('button', { name: 'Delete env' })
    }

    await userEvent.click(open)
    let commit = await settled()
    await expect(commit).toBeDisabled()
    await userEvent.click(body.getByRole('checkbox'))
    await expect(commit).toBeEnabled()

    // Cancel, reopen: the box is clear and the button is dead again.
    await userEvent.click(body.getByRole('button', { name: 'Cancel' }))
    await waitFor(async () => {
      await expect(body.queryByRole('checkbox')).toBeNull()
    })
    await userEvent.click(open)
    commit = await settled()
    await expect(body.getByRole('checkbox')).not.toBeChecked()
    await expect(commit).toBeDisabled()
  },
}

/** §6a — the red button sits LAST, after Cancel, and it is a red FILL rather
 *  than a red outline or red text. Both buttons are `flat`: destroying is work,
 *  not a commitment. */
export const FooterRanking: Story = {
  args: {
    opts: {
      title: 'Delete this addon?',
      description: 'The database and every backup taken from it are removed.',
      confirmLabel: 'Delete addon',
      variant: 'destructive',
    },
  },
  play: async ({ canvas, canvasElement, userEvent }) => {
    await userEvent.click(canvas.getByRole('button', { name: 'Delete' }))
    const body = within(canvasElement.ownerDocument.body)
    const commit = await body.findByRole('button', { name: 'Delete addon' })
    const cancel = body.getByRole('button', { name: 'Cancel' })

    // Last position.
    await expect(cancel.compareDocumentPosition(commit) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()

    // Red fill, not a red outline: the background is the danger token and the
    // label is not.
    const probe = document.createElement('div')
    probe.className = 'bg-danger'
    document.body.appendChild(probe)
    const danger = getComputedStyle(probe).backgroundColor
    probe.remove()
    await expect(getComputedStyle(commit).backgroundColor).toBe(danger)

    // `flat`, so the radius comes off the height ladder — not a pill.
    for (const btn of [commit, cancel]) {
      const style = getComputedStyle(btn)
      await expect(parseFloat(style.borderRadius)).toBe(8)
      await expect(parseFloat(style.height)).toBe(32)
    }
  },
}
