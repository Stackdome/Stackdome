import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'
import { SYNC_STATUS } from '@/pages/stacks/lib/draft-sync/constants'
import { withHeight } from '../../../../../.storybook/decorators'
import { EDITOR_TABS } from './editor-tabs'
import { CanvasEditorShell, type CanvasEditorShellProps } from './canvas-editor-shell'

const canvasPlaceholder = (
  <div className="flex h-full items-center justify-center border border-dashed border-border text-sm text-muted-foreground">
    canvas
  </div>
)

const defaultShellProps: CanvasEditorShellProps = {
  stackName: 'orders-api',
  stackId: 's1',
  headerHealth: 'ok',
  subtitle: '3 services · 2 volumes',
  hasResources: true,
  nameEditable: false,
  activeTab: EDITOR_TABS.architecture,
  onTabChange: fn(),
  isActive: false,
  dirtyResourceCount: 0,
  dirtyTotal: 0,
  isStaged: false,
  onViewChanges: fn(),
  syncStatus: SYNC_STATUS.idle,
  deployBusy: false,
  canWrite: true,
  onDeploy: fn(),
  canDiscardDraft: false,
  onDelete: fn(),
  canDeleteStack: true,
  publicEndpoints: [{ service: 'web', url: 'https://web.example.com', port: 443, variant: 'ready' }],
  architecture: canvasPlaceholder,
  deployments: <div />,
  logs: <div />,
  metrics: <div />,
}

const meta = {
  title: 'Features/EditorChrome/CanvasEditorShell',
  component: CanvasEditorShell,
  tags: ['ai-generated'],
  decorators: [withHeight(560)],
  args: defaultShellProps,
} satisfies Meta<typeof CanvasEditorShell>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

/** New stack: editable name, no autosave/actions menu, draft deploy pill. */
export const DirtyDraft: Story = {
  args: {
    stackName: 'new-stack',
    stackId: undefined,
    headerHealth: undefined,
    isNewStack: true,
    nameEditable: true,
    onNameChange: fn(),
    onDraftDeploy: fn(),
    isActive: true,
    dirtyResourceCount: 2,
    dirtyTotal: 3,
    publicEndpoints: [],
  },
}

/** Name validation failing on a draft. */
export const Validating: Story = {
  args: {
    stackName: 'New Stack!',
    stackId: undefined,
    headerHealth: undefined,
    isNewStack: true,
    nameEditable: true,
    onNameChange: fn(),
    nameError: 'Name must be lowercase letters, numbers, and dashes.',
    publicEndpoints: [],
  },
}

/** Viewer without write access: pending changes visible but deploy/delete gated. */
export const ReadOnly: Story = {
  args: {
    canWrite: false,
    canDeleteStack: false,
    isActive: true,
    dirtyResourceCount: 1,
    dirtyTotal: 2,
    syncStatus: SYNC_STATUS.saved,
  },
}
