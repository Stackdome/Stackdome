import { useCallback, useEffect, useRef, useState } from "react";
import { ExternalLink, Copy, Check } from "lucide-react";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";

const COPY_FLASH_MS = 1400;

export interface PublicEndpoint {
  service: string;
  url: string;
  port?: number;
}

function hostOf(url: string): string {
  try {
    return new URL(url).hostname;
  } catch {
    return url;
  }
}

async function copyText(text: string): Promise<void> {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text);
    return;
  }
  // Clipboard API unavailable (insecure context): textarea fallback.
  const ta = document.createElement("textarea");
  ta.value = text;
  document.body.appendChild(ta);
  ta.select();
  document.execCommand("copy");
  ta.remove();
}

/** Header row mapping each publicly exposed service to its best live URL. */
export function PublicEndpointRow({ endpoints }: { endpoints: PublicEndpoint[] }) {
  const [copiedUrl, setCopiedUrl] = useState<string | null>(null);
  const timer = useRef<ReturnType<typeof setTimeout>>(null);
  useEffect(() => () => { if (timer.current) clearTimeout(timer.current); }, []);

  const onCopy = useCallback((url: string) => {
    void copyText(url).then(() => {
      setCopiedUrl(url);
      if (timer.current) clearTimeout(timer.current);
      timer.current = setTimeout(() => setCopiedUrl(null), COPY_FLASH_MS);
    });
  }, []);

  if (endpoints.length === 0) return null;

  return (
    <div className="mt-3.5">
      <div className="mb-2 font-mono text-[9.5px] font-medium uppercase tracking-[0.16em] text-fg-muted">
        PUBLIC
      </div>
      <div className="flex flex-wrap items-center gap-2">
        {endpoints.map(({ service, url, port }) => (
          <span
            key={`${service}-${url}`}
            className="group inline-flex items-center gap-2 rounded-lg border border-border/60 bg-muted/25 py-1 pl-2.5 pr-1.5 font-mono text-[12px] transition-colors hover:border-border hover:bg-muted/40"
          >
            <Tooltip delayDuration={300}>
              <TooltipTrigger asChild>
                <span className="flex items-center gap-1.5 text-fg-muted">
                  <span aria-hidden className="size-[5px] rounded-full bg-success" />
                  {service}
                </span>
              </TooltipTrigger>
              <TooltipContent side="top">
              Mapped to {service}{port != null ? ` · :${port}` : ""}
              </TooltipContent>
            </Tooltip>
            <span aria-hidden className="text-border">|</span>
            <a
              href={url}
              target="_blank"
              rel="noreferrer"
              className="flex items-center gap-1.5 text-foreground/90 hover:text-foreground hover:underline"
            >
              {hostOf(url)}
              <ExternalLink className="size-3 text-fg-muted" />
            </a>
            <button
              type="button"
              onClick={() => onCopy(url)}
              aria-label={copiedUrl === url ? "Copied" : `Copy ${url}`}
              className="flex size-5 items-center justify-center rounded text-fg-muted opacity-0 transition-opacity hover:bg-muted hover:text-foreground focus-visible:opacity-100 group-hover:opacity-100"
            >
              {copiedUrl === url ? <Check className="size-3 text-success" /> : <Copy className="size-3" />}
            </button>
          </span>
        ))}
      </div>
    </div>
  );
}
