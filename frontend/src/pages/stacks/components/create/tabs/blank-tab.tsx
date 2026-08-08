import { Boxes, Globe, HardDrive } from "lucide-react"

/**
 * The blank canvas **explains itself instead of jumping there.**
 *
 * Every other starting point shows you what you are about to get before you
 * commit to it. Sending this one straight to an empty canvas would be the only
 * option in the strip that answers a different question — so it lists what an
 * empty canvas actually gives you, and `Create stack` still does the leaving.
 */
const WHAT_YOU_GET = [
  {
    icon: <Globe />,
    title: "Services",
    description: "Your own image, or any public one. Give it a port and a URL.",
  },
  {
    icon: <Boxes />,
    title: "Data stores",
    description: "Postgres, MySQL, Redis and the rest, wired to your services.",
  },
  {
    icon: <HardDrive />,
    title: "Volumes",
    description: "Disk that survives a redeploy.",
  },
]

export function BlankTab() {
  return (
    <div className="flex flex-col">
      {WHAT_YOU_GET.map((thing) => (
        <div key={thing.title} className="flex min-h-11 items-center gap-[11px] py-1.5">
          <span className="border-border bg-control text-fg-2 flex size-[30px] flex-none items-center justify-center rounded-md border [&_svg]:size-4">
            {thing.icon}
          </span>
          <span className="flex flex-col">
            <span className="text-body text-foreground font-medium">{thing.title}</span>
            <span className="text-meta text-fg-muted">{thing.description}</span>
          </span>
        </div>
      ))}
    </div>
  )
}
