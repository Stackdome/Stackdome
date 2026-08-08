import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, userEvent, within } from 'storybook/test'
import { http, HttpResponse } from 'msw'

import { baselineHandlers } from '../../../../../.storybook/msw-handlers'
import { BreadcrumbProvider } from '@/contexts/breadcrumb-context'
import { SheetHeader } from '@/components/sheet-header'

import CreateStackPage from './index'

/** The page renders inside the sheet, under the header that gives it its back
 *  arrow — judging it without that is judging half a screen (§2). */
function Sheet() {
  return (
    <BreadcrumbProvider>
      <div className="bg-card flex h-[820px] flex-col overflow-hidden rounded-lg">
        <SheetHeader />
        <div className="min-h-0 flex-1">
          <CreateStackPage />
        </div>
      </div>
    </BreadcrumbProvider>
  )
}

const REPOS = {
  items: [
    { full_name: 'acme/checkout-api', clone_url: 'https://github.com/acme/checkout-api.git', default_branch: 'main' },
    { full_name: 'acme/web-storefront', clone_url: 'https://github.com/acme/web-storefront.git', default_branch: 'main' },
    { full_name: 'acme/billing-worker', clone_url: 'https://github.com/acme/billing-worker.git', default_branch: 'main' },
    { full_name: 'acme/notifications', clone_url: 'https://github.com/acme/notifications.git', default_branch: 'develop' },
    { full_name: 'acme/image-proxy', clone_url: 'https://github.com/acme/image-proxy.git', default_branch: 'main' },
  ],
}

const INTEGRATIONS_PATH = '/api/v1/organizations/:orgId/git-integrations'

const CONNECTED = [
  http.get(`${INTEGRATIONS_PATH}/:id/repositories`, () => HttpResponse.json(REPOS)),
  http.get(INTEGRATIONS_PATH, () =>
    HttpResponse.json({
      items: [
        { id: 'gi-1', type: 'github_app', status: 'installed', name: 'acme', install_url: 'https://x' },
      ],
    }),
  ),
]

const meta = {
  title: 'Pages/CreateStack',
  component: CreateStackPage,
  parameters: {
    layout: 'fullscreen',
    msw: [...CONNECTED, ...baselineHandlers],
  },
  render: () => <Sheet />,
} satisfies Meta<typeof CreateStackPage>

export default meta
type Story = StoryObj<typeof meta>

/** The landing state — a repository is live, and nothing is picked yet. */
export const Default: Story = {}

/** Picking a repository fills the "In this stack" panel, which is not on screen
 *  until there is something to put in it. */
export const RepositoryPicked: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    const row = await canvas.findByRole('option', { name: /checkout-api/ })
    await userEvent.click(row)

    await expect(canvas.getByText(/In this stack/)).toBeInTheDocument()
    await expect(canvas.getByRole('button', { name: 'Create stack' })).toBeEnabled()
  },
}

/** A brand-new organisation has connected nothing, so the repository tab's very
 *  first state is an empty state rather than an empty list. */
export const NoProviderConnected: Story = {
  parameters: {
    msw: [http.get(INTEGRATIONS_PATH, () => HttpResponse.json({ items: [] })), ...baselineHandlers],
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(await canvas.findByText('No git provider connected yet')).toBeInTheDocument()
  },
}

/** The same block goes in as many times as you click it — `postgres`,
 *  `postgres-2` — and each instance comes out on its own. */
export const BlocksAddRepeatedly: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.click(canvas.getByRole('tab', { name: /Assemble from blocks/ }))

    const postgres = await canvas.findByRole('button', { name: /Postgres/ })
    await userEvent.click(postgres)
    await userEvent.click(postgres)

    await expect(canvas.getByText('postgres')).toBeInTheDocument()
    await expect(canvas.getByText('postgres-2')).toBeInTheDocument()

    await userEvent.click(canvas.getByRole('button', { name: 'Remove postgres-2' }))
    await expect(canvas.queryByText('postgres-2')).not.toBeInTheDocument()
    await expect(canvas.getByText('postgres')).toBeInTheDocument()
  },
}

/** The blank canvas is the only tab that is ready before you touch anything —
 *  and it explains what an empty canvas gives you rather than jumping there. */
export const BlankCanvas: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.click(canvas.getByRole('tab', { name: /Start from nothing/ }))

    await expect(canvas.getByText('Volumes')).toBeInTheDocument()
    await expect(canvas.getByRole('button', { name: 'Create stack' })).toBeEnabled()
  },
}

/** Compose reads the file before you commit to it. */
export const ComposeParsePreview: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.click(canvas.getByRole('tab', { name: /Import a compose file/ }))
    await userEvent.click(canvas.getByRole('button', { name: 'Paste an example' }))

    await expect(canvas.getByText(/Found in your file/)).toBeInTheDocument()
    // Twice, and correctly so: once as a chip in the parse preview, once as a
    // row in "In this stack".
    await expect(canvas.getAllByText('pgdata')).toHaveLength(2)
  },
}

/** A ready-made app opens its detail panel — every field the record carries. */
export const TemplateDetailPanel: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.click(canvas.getByRole('tab', { name: /Run a ready-made app/ }))

    const first = (await canvas.findAllByRole('option'))[0]
    await userEvent.click(first)

    await expect(canvas.getByRole('link', { name: 'Docs' })).toBeInTheDocument()
  },
}

/** Switching tabs must not throw away what another tab already gathered. */
export const SwitchingTabsKeepsWhatYouPicked: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.click(await canvas.findByRole('option', { name: /checkout-api/ }))

    await userEvent.click(canvas.getByRole('tab', { name: /Start from nothing/ }))
    await userEvent.click(canvas.getByRole('tab', { name: /Deploy your own code/ }))

    const live = canvas.getAllByRole('option').filter((o) => o.getAttribute('aria-selected') === 'true')
    await expect(live).toHaveLength(1)
    await expect(live[0]).toHaveAccessibleName(/checkout-api/)
  },
}
