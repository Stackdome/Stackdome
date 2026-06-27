import { useMemo, useState, type KeyboardEvent } from "react";
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Search, ExternalLink, ArrowRight } from "lucide-react";
import { cn } from "@/lib/utils";
import type { Template } from "@/data/templates/types";

interface TemplatesBrowserDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  templates: Template[];
  onUse: (template: Template) => void;
}

function TemplateBadge({
  template,
  className,
  active,
}: {
  template: Template;
  className?: string;
  active?: boolean;
}) {
  const [broken, setBroken] = useState(false);
  return (
    <div
      className={cn(
        "flex shrink-0 items-center justify-center overflow-hidden rounded-md border border-border bg-muted font-mono font-semibold",
        active ? "text-brand" : "text-foreground/80",
        className,
      )}
    >
      {template.icon && !broken ? (
        <img
          src={template.icon}
          alt=""
          className="h-full w-full object-contain p-1"
          onError={() => setBroken(true)}
        />
      ) : (
        <span>{template.initials}</span>
      )}
    </div>
  );
}

function ExternalLinkButton({ href, label }: { href: string; label: string }) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className="inline-flex items-center gap-2 rounded-md border border-border px-3 py-2 text-xs font-medium text-foreground/80 transition-colors hover:border-brand-border hover:text-brand"
    >
      <ExternalLink className="h-3.5 w-3.5" />
      {label}
    </a>
  );
}

export default function TemplatesBrowserDialog({
  open,
  onOpenChange,
  templates,
  onUse,
}: TemplatesBrowserDialogProps) {
  const [query, setQuery] = useState("");
  const [selectedId, setSelectedId] = useState<string | null>(null);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return templates;
    return templates.filter(
      (t) =>
        t.name.toLowerCase().includes(q) ||
        t.shortDescription.toLowerCase().includes(q) ||
        t.category.toLowerCase().includes(q),
    );
  }, [templates, query]);

  const activeId = filtered.some((t) => t.id === selectedId)
    ? selectedId
    : (filtered[0]?.id ?? null);
  const selected = filtered.find((t) => t.id === activeId) ?? null;

  const handleKeyDown = (e: KeyboardEvent) => {
    if (!filtered.length) return;
    if (e.key === "ArrowDown" || e.key === "ArrowUp") {
      e.preventDefault();
      let i = filtered.findIndex((t) => t.id === activeId);
      if (i < 0) i = 0;
      i =
        e.key === "ArrowDown"
          ? Math.min(filtered.length - 1, i + 1)
          : Math.max(0, i - 1);
      setSelectedId(filtered[i].id);
    } else if (e.key === "Enter") {
      e.preventDefault();
      if (selected) onUse(selected);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="block gap-0 overflow-hidden p-0 sm:max-w-[1040px]">
        <div onKeyDown={handleKeyDown} className="flex flex-col">
          {/* header */}
          <div className="flex items-start justify-between gap-4 border-b border-border px-6 py-5">
            <div>
              <div className="font-mono text-[11px] font-medium uppercase tracking-[1.5px] text-brand">
                Import / Templates
              </div>
              <DialogTitle className="mt-1.5 text-xl font-medium tracking-tight">
                Choose a template
              </DialogTitle>
              <DialogDescription className="sr-only">
                Pick a curated template to prefill the new-stack form.
              </DialogDescription>
            </div>
          </div>

          {/* body: split */}
          <div className="grid h-[520px] grid-cols-[340px_1fr]">
            {/* left: search + list */}
            <div className="flex min-h-0 flex-col border-r border-border">
              <div className="relative px-3.5 pb-2.5 pt-3.5">
                <Search className="pointer-events-none absolute left-[26px] top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
                <Input
                  autoFocus
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Search templates…"
                  className="pl-9"
                />
              </div>
              <ul
                role="listbox"
                aria-label="Templates"
                className="min-h-0 flex-1 space-y-0.5 overflow-y-auto px-2 pb-2"
              >
                {filtered.map((t) => {
                  const isActive = t.id === activeId;
                  return (
                    <li
                      key={t.id}
                      role="option"
                      aria-selected={isActive}
                      onClick={() => setSelectedId(t.id)}
                      className={cn(
                        "flex cursor-pointer items-center gap-2.5 rounded-md border px-2.5 py-2.5",
                        isActive
                          ? "border-border bg-muted"
                          : "border-transparent hover:bg-muted/50",
                      )}
                    >
                      <TemplateBadge
                        template={t}
                        active={isActive}
                        className="h-[30px] w-[30px] text-[11px]"
                      />
                      <div className="min-w-0 flex-1">
                        <div className="truncate text-sm font-medium">
                          {t.name}
                        </div>
                        <div className="truncate text-[11.5px] text-muted-foreground">
                          {t.shortDescription}
                        </div>
                      </div>
                    </li>
                  );
                })}
                {filtered.length === 0 && (
                  <li className="px-3.5 py-7 text-center text-sm text-muted-foreground">
                    No templates match{" "}
                    <span className="font-mono text-foreground">{query}</span>.
                  </li>
                )}
              </ul>
              <div className="border-t border-border px-4 py-3">
                <span className="font-mono text-[10.5px] tracking-[0.5px] text-muted-foreground">
                  {filtered.length} templates
                </span>
              </div>
            </div>

            {/* right: detail */}
            {selected ? (
              <div className="flex min-h-0 flex-col px-8 pb-7 pt-8">
                <div className="flex items-center gap-4">
                  <TemplateBadge
                    template={selected}
                    className="h-[60px] w-[60px] rounded-[10px] text-[22px]"
                  />
                  <div>
                    <div className="text-2xl font-medium tracking-tight">
                      {selected.name}
                    </div>
                    <div className="mt-1.5 font-mono text-[10.5px] uppercase tracking-[1.2px] text-muted-foreground">
                      Self-hosted · {selected.category} · {selected.version}
                    </div>
                  </div>
                </div>
                <p className="mt-5 max-w-[580px] text-[15px] leading-relaxed text-foreground/90">
                  {selected.longDescription}
                </p>
                <div className="mt-5 flex gap-2.5">
                  <ExternalLinkButton href={selected.website} label="Website" />
                  <ExternalLinkButton href={selected.docs} label="Docs" />
                </div>
                <div className="flex-1" />
                <div className="flex items-center justify-between gap-4 border-t border-border pt-5">
                  <p className="max-w-[300px] text-xs leading-snug text-muted-foreground">
                    Opens the new-stack form, prefilled with this template.
                  </p>
                  <Button onClick={() => onUse(selected)}>
                    Use template
                    <ArrowRight className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            ) : (
              <div className="flex items-center justify-center text-sm text-muted-foreground">
                Select a template
              </div>
            )}
          </div>

          {/* footer hints */}
          <div className="flex items-center gap-4 border-t border-border bg-background/50 px-6 py-3 font-mono text-[10.5px] tracking-[0.4px] text-muted-foreground">
            <span>↑↓ navigate</span>
            <span>↵ create stack</span>
            <span>esc close</span>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
