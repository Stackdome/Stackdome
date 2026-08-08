import { useEffect, useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { expect, userEvent, within } from 'storybook/test'
import { useLocation, useNavigate } from 'react-router-dom'
import { PanelLeftClose } from 'lucide-react'
import { SheetHeader } from './sheet-header'
import { BreadcrumbProvider } from '@/contexts/breadcrumb-context'
import { useJourney } from '@/hooks/use-journey'
import { Button } from '@/components/ui/button'

/** Stands in for the shell's collapse toggle, which `AppLayout` passes in. */
const COLLAPSE = (
  <Button variant="ghost" size="icon" className="size-8 text-fg-2" aria-label="Collapse sidebar">
    <PanelLeftClose />
  </Button>
)

/** Declares the page a journey, the way the New stack page will. */
function DeclareJourney({ origin }: { origin: string }) {
  useJourney(origin)
  return null
}

/** Reports where we ended up, so a `play` can assert on navigation. */
function Where() {
  const { pathname } = useLocation()
  return <div data-testid="where">{pathname}</div>
}

/**
 * The global preview decorator already supplies a MemoryRouter (nesting a
 * second one throws), so this hops that router instead.
 *
 * `from` is the difference between the two back behaviours: with it, the
 * journey is PUSHed on top of a real previous entry; without it the journey is
 * REPLACEd in, which is what a deep link or a fresh tab looks like.
 */
function Harness({
  path,
  from,
  journey,
}: {
  path: string
  from?: string
  journey?: boolean
}) {
  const navigate = useNavigate()
  // Two navigations in one effect collapse into a single history entry, so the
  // push has to land on its own render for there to be anything behind it.
  const [step, setStep] = useState(0)
  useEffect(() => {
    if (step === 0) {
      navigate(from ?? path, { replace: true })
      setStep(from ? 1 : 2)
    } else if (step === 1) {
      navigate(path)
      setStep(2)
    }
  }, [step, navigate, path, from])
  if (step < 2) return null
  return (
    <BreadcrumbProvider>
      {journey && <DeclareJourney origin="/stacks" />}
      <div className="bg-card w-[1186px] max-w-full">
        <SheetHeader leading={COLLAPSE} />
      </div>
      <Where />
    </BreadcrumbProvider>
  )
}

const meta = {
  title: 'Branded/SheetHeader',
  component: SheetHeader,
  parameters: { layout: 'fullscreen' },
  render: () => <Harness path="/stacks" />,
} satisfies Meta<typeof SheetHeader>

export default meta
type Story = StoryObj<typeof meta>

/** A top-level page shows its title alone — there is no `Home` crumb. */
export const TopLevel: Story = {}

/** Nested: the trail is the way up, and there is no back arrow. */
export const Nested: Story = {
  render: () => <Harness path="/stacks/acme-web/environment" />,
}

/** A journey: back arrow, divider, and the title alone. No trail. */
export const Journey: Story = {
  render: () => <Harness path="/stacks/new" journey />,
}

/** The two treatments are mutually exclusive — never a trail and an arrow. */
export const JourneyDropsTheTrail: Story = {
  render: () => <Harness path="/stacks/new" journey />,
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(canvas.getByRole('button', { name: 'Back' })).toBeInTheDocument()
    // "New" is the title; "Stacks" would be the crumb we deliberately dropped.
    await expect(canvas.queryByRole('link', { name: 'Stacks' })).not.toBeInTheDocument()
  },
}

export const NestedHasNoBackArrow: Story = {
  render: () => <Harness path="/stacks/acme-web/environment" />,
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(canvas.queryByRole('button', { name: 'Back' })).not.toBeInTheDocument()
    await expect(canvas.getByRole('link', { name: 'Stacks' })).toBeInTheDocument()
  },
}

/** Arrived by navigating: back steps the way you came, not to the origin. */
export const BackStepsThroughHistory: Story = {
  render: () => <Harness path="/stacks/new" from="/previews" journey />,
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(canvas.getByTestId('where')).toHaveTextContent('/stacks/new')
    await userEvent.click(canvas.getByRole('button', { name: 'Back' }))
    await expect(canvas.getByTestId('where')).toHaveTextContent('/previews')
  },
}

/** Deep-linked with nothing behind it: back falls back to the origin, rather
 *  than stepping out of the product. */
export const BackFallsBackToOrigin: Story = {
  render: () => <Harness path="/stacks/new" journey />,
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.click(canvas.getByRole('button', { name: 'Back' }))
    await expect(canvas.getByTestId('where')).toHaveTextContent('/stacks')
  },
}
