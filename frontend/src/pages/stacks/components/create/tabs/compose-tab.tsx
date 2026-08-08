import { useRef } from "react"
import { Upload } from "lucide-react"

import { AlertBanner } from "@/components/branded"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { parseAndValidateDockerCompose } from "@/pages/stacks/lib/docker-compose-parser"

import type { ComposePreview } from "../selection"

const EXAMPLE = `services:
  web:
    image: nginx:1.27
    ports:
      - "8080:80"
  api:
    image: ghcr.io/acme/api:latest
    depends_on:
      - db
  db:
    image: postgres:16
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
`

/**
 * Reads a compose file the moment there is one to read.
 *
 * **The preview is the point.** An import that only tells you what it did after
 * it has done it makes you undo work to find out; showing the services and
 * volumes it found, before you commit, is what turns this from a leap into a
 * choice.
 */
export function parseCompose(yaml: string): ComposePreview | null {
  if (!yaml.trim()) return null
  try {
    const parsed = parseAndValidateDockerCompose(yaml) as {
      services?: Record<string, unknown>
      volumes?: Record<string, unknown>
    }
    return {
      services: Object.keys(parsed.services ?? {}),
      volumes: Object.keys(parsed.volumes ?? {}),
      warnings: [],
      error: null,
    }
  } catch (error) {
    return {
      services: [],
      volumes: [],
      warnings: [],
      error: error instanceof Error ? error.message : "That file could not be read as compose YAML.",
    }
  }
}

export function ComposeTab({
  yaml,
  onChange,
}: {
  yaml: string
  onChange: (yaml: string) => void
}) {
  const fileInput = useRef<HTMLInputElement>(null)
  const preview = parseCompose(yaml)
  const hasContent = yaml.trim().length > 0

  async function readFile(file: File) {
    onChange(await file.text())
  }

  return (
    <div className="flex flex-col gap-2.5">
      <Textarea
        value={yaml}
        onChange={(event) => onChange(event.target.value)}
        placeholder="Paste your docker-compose.yml here…"
        aria-label="Compose file"
        spellCheck={false}
        // The field's own well, not `--code-bg`: that ground is the terminal,
        // and it is near-black in BOTH themes (§3). This is something you type
        // into, so it takes the input fill.
        className="h-[260px] resize-none font-mono text-meta leading-5"
        onDrop={(event) => {
          const file = event.dataTransfer.files[0]
          if (!file) return
          event.preventDefault()
          void readFile(file)
        }}
      />

      {/* Working controls, so `flat` and ghost — the one filled button on this
          screen is `Create stack` in the footer (§9/§11). */}
      <div className="flex items-center gap-1.5">
        <Button variant="outline" shape="flat" onClick={() => fileInput.current?.click()}>
          <Upload />
          Upload a file
        </Button>
        <Button variant="ghost" shape="flat" onClick={() => onChange(EXAMPLE)}>
          {/* Once there is content, pasting an example replaces it — and the
              label has to say so, or the click is a surprise. */}
          {hasContent ? "Replace with an example" : "Paste an example"}
        </Button>
        <Button
          variant="ghost"
          shape="flat"
          className="ml-auto"
          disabled={!hasContent}
          onClick={() => onChange("")}
        >
          Clear
        </Button>
        <input
          ref={fileInput}
          type="file"
          accept=".yml,.yaml,text/yaml"
          className="hidden"
          onChange={(event) => {
            const file = event.target.files?.[0]
            if (file) void readFile(file)
            event.target.value = ""
          }}
        />
      </div>

      {preview?.error && <AlertBanner tone="blocking">{preview.error}</AlertBanner>}

      {preview && !preview.error && (
        <div className="border-border-subtle mt-1.5 border-t pt-3">
          <div className="text-label text-fg-muted">
            Found in your file ·{" "}
            <span className="text-foreground">
              {preview.services.length} {preview.services.length === 1 ? "service" : "services"}
              {preview.volumes.length > 0 &&
                ` · ${preview.volumes.length} ${preview.volumes.length === 1 ? "volume" : "volumes"}`}
            </span>
          </div>
          <div className="mt-2 flex flex-wrap gap-1.5">
            {[...preview.services, ...preview.volumes].map((name) => (
              <span
                key={name}
                className="border-border bg-control text-foreground rounded-sm border px-1.5 py-0.5 font-mono text-[11px] leading-4"
              >
                {name}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
