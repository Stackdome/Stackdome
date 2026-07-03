import { useCallback, useEffect, useRef, useState } from "react";
import { Globe, ExternalLink, Copy, Check } from "lucide-react";
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
    <div className="mt-2.5 flex flex-wrap items-center gap-2">
      <span className="font-mono text-[9.5px] uppercase tracking-[0.14em] text-fg-muted">PUBLIC</span>
      {endpoints.map(({ service, url, port }) => (
        <span
          key={`${service}-${url}`}
          className="inline-flex items-stretch overflow-hidden rounded-md border border-border bg-muted/40 text-[12px]"
        >
          <Tooltip delayDuration={300}>
            <TooltipTrigger asChild>
              <span className="flex items-center gap-1.5 border-r border-border px-2 py-1 font-mono text-fg-muted">
                <Globe className="size-3" />
                {service}
              </span>
            </TooltipTrigger>
            <TooltipContent side="top">
              Mapped to {service}{port != null ? ` · :${port}` : ""}
            </TooltipContent>
          </Tooltip>
          <a
            href={url}
            target="_blank"
            rel="noreferrer"
            className="flex items-center gap-1.5 px-2 py-1 font-mono text-foreground hover:bg-muted"
          >
            <span aria-hidden className="size-[5px] rounded-full bg-success" />
            {hostOf(url)}
            <ExternalLink className="size-3 text-fg-muted" />
          </a>
          <button
            type="button"
            onClick={() => onCopy(url)}
            aria-label={copiedUrl === url ? "Copied" : `Copy ${url}`}
            className="flex items-center border-l border-border px-1.5 text-fg-muted hover:bg-muted hover:text-foreground"
          >
            {copiedUrl === url ? <Check className="size-3 text-success" /> : <Copy className="size-3" />}
          </button>
        </span>
      ))}
    </div>
  );
}
