import { useCallback, useEffect, useRef, useState } from "react";
import { ExternalLink, Copy, Check } from "lucide-react";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
import type { StatusVariant } from "@/components/branded/status-variant";
import { cn } from "@/lib/utils";

const COPY_FLASH_MS = 1400;

/** Per-endpoint resource status variant → dot colour (semantic tokens only). */
const DOT_CLASS: Record<StatusVariant, string> = {
  ready: "bg-success",
  pending: "bg-warn",
  error: "bg-danger",
  info: "bg-info",
  neutral: "bg-fg-muted",
};

export interface PublicEndpoint {
  service: string;
  url: string;
  port?: number;
  /** The owning resource's live rollout status — each chip's dot reflects its
   *  own service, not the stack-level rollup. */
  variant?: StatusVariant;
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
export function PublicEndpointRow({
  endpoints,
}: {
  endpoints: PublicEndpoint[];
}) {
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
    <div className="mt-3.5 flex flex-wrap items-center gap-2">
      <span className="font-mono text-[9.5px] font-medium uppercase tracking-[0.16em] text-fg-muted">
        PUBLIC
      </span>
      {endpoints.map(({ service, url, port, variant }) => (
        <span
          key={`${service}-${url}`}
          className="group inline-flex items-center gap-1.5 rounded-lg border border-border/60 bg-muted/25 py-1 pl-2.5 pr-1.5 font-mono text-[12px] transition-colors hover:border-border hover:bg-muted/40"
        >
          <Tooltip delayDuration={300}>
            <TooltipTrigger asChild>
              <span className="flex items-center gap-1.5 text-fg-muted">
                <span aria-hidden className={cn("size-[5px] rounded-full", DOT_CLASS[variant ?? "neutral"])} />
                {service}
              </span>
            </TooltipTrigger>
            <TooltipContent side="top">
              Mapped to {service}{port != null ? ` · :${port}` : ""}
            </TooltipContent>
          </Tooltip>
          {/* Hostname stays collapsed until the chip is hovered or focused;
                the always-visible icons carry the affordance at rest. */}
          <a
            href={url}
            target="_blank"
            rel="noreferrer"
            aria-label={`Go to ${url}`}
            className="flex h-5 min-w-5 items-center justify-center rounded text-fg-muted transition-colors hover:bg-muted hover:text-foreground"
          >
            <span className="flex max-w-0 items-center overflow-hidden whitespace-nowrap text-foreground/90 opacity-0 transition-[max-width,opacity] duration-200 group-focus-within:max-w-[360px] group-focus-within:opacity-100 group-hover:max-w-[360px] group-hover:opacity-100">
              <span aria-hidden className="pr-1.5 text-border">|</span>
              <span className="pr-1.5 hover:underline">{hostOf(url)}</span>
            </span>
            <ExternalLink className="size-3 shrink-0" />
          </a>
          <button
            type="button"
            onClick={() => onCopy(url)}
            aria-label={copiedUrl === url ? "Copied" : `Copy ${url}`}
            className="flex size-5 items-center justify-center rounded text-fg-muted transition-colors hover:bg-muted hover:text-foreground"
          >
            {copiedUrl === url ? <Check className="size-3 text-success" /> : <Copy className="size-3" />}
          </button>
        </span>
      ))}
    </div>
  );
}
