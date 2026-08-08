import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn, within } from 'storybook/test'
import AddClusterDialog from './add-cluster-dialog'

const meta = {
  title: 'Features/Clusters/AddClusterDialog',
  component: AddClusterDialog,
  tags: ['ai-generated'],
  args: {
    open: true,
    onOpenChange: fn(),
    onAddCluster: fn(),
  },
} satisfies Meta<typeof AddClusterDialog>

export default meta
type Story = StoryObj<typeof meta>

/** Empty form, image registry enabled by default — the switch control here
 *  had zero story coverage before this pass (deferred from Task 7). */
export const Default: Story = {
  play: async ({ canvasElement }) => {
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByRole('dialog')).toBeInTheDocument()
    await expect(body.getByLabelText(/Cluster Name/i)).toBeInTheDocument()

    const registrySwitch = body.getByRole('switch', { name: /enable image registry/i })
    await expect(registrySwitch).toHaveAttribute('data-state', 'checked')
    await expect(body.getByLabelText(/Backend Storage Size/i)).toBeInTheDocument()

    // Submit stays disabled until every required field is filled.
    await expect(body.getByRole('button', { name: /add cluster/i })).toBeDisabled()
  },
}

/** Toggling the switch off hides the dependent registry-size field — the
 *  hue-free control-fill/state contract (rubric #6/#7), not a color state. */
export const RegistryDisabled: Story = {
  play: async ({ canvasElement, userEvent }) => {
    const body = within(canvasElement.ownerDocument.body)
    const registrySwitch = await body.findByRole('switch', { name: /enable image registry/i })
    await userEvent.click(registrySwitch)
    await expect(registrySwitch).toHaveAttribute('data-state', 'unchecked')
    await expect(body.queryByLabelText(/Backend Storage Size/i)).toBeNull()
  },
}

/** Top-level submission failure renders through the shared AlertBanner
 *  (danger tokens, hairline border) instead of an ad hoc page-patched div. */
export const SubmissionError: Story = {
  args: {
    error: 'No organization selected',
  },
  play: async ({ canvasElement }) => {
    const body = within(canvasElement.ownerDocument.body)
    const alert = await body.findByRole('alert')
    await expect(alert).toHaveTextContent('No organization selected')
    await expect(alert.className).toContain('border-danger-border')
  },
}
