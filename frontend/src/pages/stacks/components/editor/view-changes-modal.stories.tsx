import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, fn, within } from 'storybook/test'
import type { SnapshotDiff } from '@/pages/stacks/components/editor/tabs/deployments/release-snapshot-diff'
import { ViewChangesModal } from './view-changes-modal'

const diff: SnapshotDiff = {
  resources: [
    {
      name: 'web',
      change: 'modified',
      sections: [
        {
          kind: 'configuration',
          rows: [{ key: 'image', from: 'nginx:1.26', to: 'nginx:1.27', kind: 'changed' }],
        },
        {
          kind: 'environment',
          rows: [{ key: 'LOG_LEVEL', to: 'debug', kind: 'added' }],
        },
      ],
    },
    { name: 'worker', change: 'removed', sections: [], note: 'Resource removed from the stack' },
    { name: 'api-v2', fromName: 'api', change: 'renamed', sections: [] },
  ],
  volumes: [
    { name: 'web-data', change: 'added', rows: [{ key: 'size', to: '5Gi', kind: 'added' }] },
  ],
}

const meta = {
  title: 'Features/EditorChrome/ViewChangesModal',
  component: ViewChangesModal,
  tags: ['ai-generated'],
  args: {
    open: true,
    onOpenChange: fn(),
    count: 5,
    stackName: 'orders-api',
    onDiscardResource: fn(),
    onDiscardVolume: fn(),
    onDiscardAll: fn(),
    onDeploy: fn(),
    deployBusy: false,
    canWrite: true,
  },
} satisfies Meta<typeof ViewChangesModal>

export default meta
type Story = StoryObj<typeof meta>

export const Open: Story = {
  args: { diff },
  play: async ({ canvasElement }) => {
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByText('Undeployed changes')).toBeInTheDocument()
    await expect(body.getByText('Resources')).toBeInTheDocument()
    await expect(body.getByText('nginx:1.27')).toBeInTheDocument()
  },
}

/** Diff not resolved yet — autosave still in flight. */
export const Saving: Story = {
  args: { diff: undefined },
  play: async ({ canvasElement }) => {
    const body = within(canvasElement.ownerDocument.body)
    await expect(await body.findByText('Loading changes…')).toBeInTheDocument()
  },
}

/** Autosave in a terminal error state: empty diff means "not saved". */
export const SaveErrored: Story = {
  args: {
    diff: { resources: [], volumes: [] },
    errored: true,
  },
}
