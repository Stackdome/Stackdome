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

function useCopyFlash() {
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

  return { copiedUrl, onCopy };
}

/** Hostname + go-to + copy on one static line — for drawer headers, where
 *  there's room to show the address and hover-reveal would be ceremony. */
export function EndpointInline({ url }: { url: string }) {
  const { copiedUrl, onCopy } = useCopyFlash();
  return (
    <span className="flex min-w-0 items-center gap-0.5">
      <Tooltip delayDuration={300}>
        <TooltipTrigger asChild>
          <a
            href={url}
            target="_blank"
            rel="noreferrer"
            className="flex min-w-0 items-center gap-1 font-mono text-[11px] text-fg-2 hover:text-foreground"
          >
            <span className="truncate hover:underline">{hostOf(url)}</span>
            <ExternalLink className="size-3 shrink-0 text-fg-muted" />
          </a>
        </TooltipTrigger>
        <TooltipContent side="bottom">{url}</TooltipContent>
      </Tooltip>
      <button
        type="button"
        onClick={() => onCopy(url)}
        aria-label={copiedUrl === url ? "Copied" : `Copy ${url}`}
        className="flex size-5 shrink-0 items-center justify-center rounded text-fg-muted transition-colors hover:bg-muted hover:text-foreground"
      >
        {copiedUrl === url ? <Check className="size-3 text-success" /> : <Copy className="size-3" />}
      </button>
    </span>
  );
}

interface ChipProps {
  endpoint: PublicEndpoint;
  /** "inline": hostname expands inside the chip on hover (expanded header).
   *  "tooltip": chip stays icon-compact; the URL lives in a tooltip (zen bar). */
  reveal: "inline" | "tooltip";
  copiedUrl: string | null;
  onCopy: (url: string) => void;
}

function EndpointChip({ endpoint: { service, url, port, variant }, reveal, copiedUrl, onCopy }: ChipProps) {
  const serviceLabel = (
    <span className="flex items-center gap-1.5 text-fg-muted">
      <span aria-hidden className={cn("size-[5px] rounded-full", DOT_CLASS[variant ?? "neutral"])} />
      {service}
    </span>
  );

  const chip = (
    <span
      className={cn(
        "group inline-flex items-center rounded-lg border border-border/60 bg-muted/25 font-mono transition-colors hover:border-border hover:bg-muted/40",
        reveal === "tooltip" ? "gap-1 py-0.5 pl-2 pr-1 text-[11px]" : "gap-1.5 py-1 pl-2.5 pr-1.5 text-[12px]",
      )}
    >
      {reveal === "inline" ? (
        <Tooltip delayDuration={300}>
          <TooltipTrigger asChild>{serviceLabel}</TooltipTrigger>
          <TooltipContent side="top">
            Mapped to {service}{port != null ? ` · :${port}` : ""}
          </TooltipContent>
        </Tooltip>
      ) : (
        serviceLabel
      )}
      <a
        href={url}
        target="_blank"
        rel="noreferrer"
        aria-label={`Go to ${url}`}
        className="flex h-5 min-w-5 items-center justify-center rounded text-fg-muted transition-colors hover:bg-muted hover:text-foreground"
      >
        <ExternalLink className="size-3 shrink-0" />
      </a>
      {/* Copy earns its keep in the full header; the zen bar keeps go-to only. */}
      {reveal === "inline" && (
        <button
          type="button"
          onClick={() => onCopy(url)}
          aria-label={copiedUrl === url ? "Copied" : `Copy ${url}`}
          className="flex size-5 items-center justify-center rounded text-fg-muted transition-colors hover:bg-muted hover:text-foreground"
        >
          {copiedUrl === url ? <Check className="size-3 text-success" /> : <Copy className="size-3" />}
        </button>
      )}
      {/* Hostname expands AFTER the icons so the click targets never move
          mid-hover. Redundant with the go-to icon, so it's out of the tab
          order and hidden from readers. */}
      {reveal === "inline" && (
        <a
          href={url}
          target="_blank"
          rel="noreferrer"
          tabIndex={-1}
          aria-hidden
          className="flex items-center text-foreground/90 hover:text-foreground"
        >
          <span className="flex max-w-0 items-center overflow-hidden whitespace-nowrap opacity-0 transition-[max-width,opacity] duration-200 group-focus-within:max-w-[360px] group-focus-within:opacity-100 group-hover:max-w-[360px] group-hover:opacity-100">
            <span className="px-1 text-border">|</span>
            <span className="pr-1 hover:underline">{hostOf(url)}</span>
          </span>
        </a>
      )}
    </span>
  );

  if (reveal === "inline") return chip;

  // Compact: one tooltip for the whole chip carries mapping + URL.
  return (
    <Tooltip delayDuration={300}>
      <TooltipTrigger asChild>{chip}</TooltipTrigger>
      <TooltipContent side="bottom">
        {service}{port != null ? ` · :${port}` : ""} — {url}
      </TooltipContent>
    </Tooltip>
  );
}

/** Public service → best live URL chips. Expanded header: labelled row with
 *  hover-expanding hostnames. `compact`: label-less tooltip chips sized for
 *  the zen/collapsed header bar. */
export function PublicEndpointRow({
  endpoints,
  compact = false,
}: {
  endpoints: PublicEndpoint[];
  compact?: boolean;
}) {
  const { copiedUrl, onCopy } = useCopyFlash();

  if (endpoints.length === 0) return null;

  if (compact) {
    return (
      <div className="flex flex-none items-center gap-1.5">
        {endpoints.map((e) => (
          <EndpointChip key={`${e.service}-${e.url}`} endpoint={e} reveal="tooltip" copiedUrl={copiedUrl} onCopy={onCopy} />
        ))}
      </div>
    );
  }

  return (
    <div className="mt-3.5 flex flex-wrap items-center gap-2">
      <span className="font-mono text-[9.5px] font-medium uppercase tracking-[0.16em] text-fg-muted">
        PUBLIC
      </span>
      {endpoints.map((e) => (
        <EndpointChip key={`${e.service}-${e.url}`} endpoint={e} reveal="inline" copiedUrl={copiedUrl} onCopy={onCopy} />
      ))}
    </div>
  );
}
