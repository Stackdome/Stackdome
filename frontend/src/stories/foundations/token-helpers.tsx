import { useEffect, useState } from 'react'

// Resolved at render time so the light/dark toolbar shows each theme's value.
function useResolvedToken(token: string) {
  const [value, setValue] = useState('')
  useEffect(() => {
    const read = () =>
      setValue(getComputedStyle(document.documentElement).getPropertyValue(token).trim())
    read()
    const observer = new MutationObserver(read)
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
    return () => observer.disconnect()
  }, [token])
  return value
}

export function Swatch({ token }: { token: string }) {
  const value = useResolvedToken(token)
  return (
    <div className="flex items-center gap-3 py-1.5">
      <div
        className="h-9 w-9 shrink-0 rounded-sm border border-border"
        style={{ background: `var(${token})` }}
      />
      <div className="min-w-0">
        <div className="font-mono text-xs font-bold">{token}</div>
        <div className="font-mono text-[11px] text-muted-foreground">{value}</div>
      </div>
    </div>
  )
}

export function SwatchGroup({ title, tokens }: { title: string; tokens: string[] }) {
  return (
    <div className="mb-6">
      <div className="mb-2 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
        {title}
      </div>
      <div className="grid grid-cols-1 gap-x-8 sm:grid-cols-2 lg:grid-cols-3">
        {tokens.map((t) => (
          <Swatch key={t} token={t} />
        ))}
      </div>
    </div>
  )
}

export function TypeSample({
  label,
  className,
  children,
}: {
  label: string
  className: string
  children: React.ReactNode
}) {
  return (
    <div className="flex items-baseline gap-4 border-b border-border py-3">
      <div className="w-56 shrink-0 font-mono text-[11px] text-muted-foreground">{label}</div>
      <div className={className}>{children}</div>
    </div>
  )
}
