import { useEffect, useState } from "react";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { StatusPill, variantFromState } from "@/components/branded";
import { getRelease, type StackRelease } from "@/api/releases";
import { causeLabel, formatDuration } from "./derive";
import { diffSnapshots } from "./release-diff";

export interface ReleaseDetailDrawerProps {
  orgId: string;
  teamName: string;
  stackId: string;
  releaseId: string;
  previousRelease?: StackRelease;
  onClose: () => void;
}

export function ReleaseDetailDrawer({ orgId, teamName, stackId, releaseId, previousRelease, onClose }: ReleaseDetailDrawerProps) {
  const [release, setRelease] = useState<StackRelease | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let alive = true;
    getRelease(orgId, teamName, stackId, releaseId)
      .then((r) => { if (alive) setRelease(r); })
      .catch((e) => { if (alive) setError(e instanceof Error ? e.message : "Failed to load release"); });
    return () => { alive = false; };
  }, [orgId, teamName, stackId, releaseId]);

  const outcomes = Object.entries(
    (release as unknown as { outcome?: { resources?: Record<string, { phase?: string; ready_replicas?: number; replicas?: number; message?: string }> } })?.outcome?.resources ?? {}
  );

  // snapshot is an untyped JSONB blob on the full release; cast for diffing.
  const snap = (release as unknown as { snapshot?: unknown })?.snapshot;
  const prevSnap = (previousRelease as unknown as { snapshot?: unknown })?.snapshot;
  const changes = release && previousRelease ? diffSnapshots(prevSnap, snap) : [];

  return (
    <Sheet open onOpenChange={(o) => { if (!o) onClose(); }}>
      <SheetContent side="right" className="w-[480px] sm:max-w-[480px] overflow-y-auto">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2">
            Release #{release?.sequence ?? "…"}
            {release && <StatusPill variant={variantFromState(release.state ?? "")}>{release.state}</StatusPill>}
          </SheetTitle>
        </SheetHeader>

        {error && <p className="mt-4 text-[13px] text-danger">{error}</p>}

        {release && (
          <div className="mt-4 space-y-6 text-[13px]">
            <section className="space-y-1">
              <div className="text-muted-foreground">{causeLabel(release.cause)}</div>
              <div className="text-muted-foreground">
                duration {formatDuration(release.rendered_at, release.completed_at)}
              </div>
            </section>

            {release.state === "Failed" && release.message && (
              <section>
                <h3 className="mb-1 font-medium">Why it failed</h3>
                <p className="rounded-md border border-danger-border bg-danger-bg px-2 py-2 text-danger">{release.message}</p>
              </section>
            )}

            {outcomes.length > 0 && (
              <section>
                <h3 className="mb-2 font-medium">Resource outcomes</h3>
                <div className="divide-y divide-border rounded-md border border-border">
                  {outcomes.map(([name, o]) => (
                    <div key={name} className="flex items-center justify-between px-3 py-2">
                      <span className="font-medium">{name}</span>
                      <span className="text-muted-foreground">{o?.phase}</span>
                      <span className="font-mono text-muted-foreground">{o?.ready_replicas ?? 0}/{o?.replicas ?? 0}</span>
                    </div>
                  ))}
                </div>
              </section>
            )}

            {changes.length > 0 && (
              <section>
                <h3 className="mb-2 font-medium">Config changes vs #{previousRelease?.sequence}</h3>
                <div className="space-y-1 font-mono text-[12px]">
                  {changes.map((c) => (
                    <div key={c.path} className="rounded border-l-2 border-warn-border bg-warn-bg px-2 py-1">
                      <div className="text-muted-foreground">{c.path}</div>
                      {c.kind !== "added" && <div className="text-danger">- {JSON.stringify(c.before)}</div>}
                      {c.kind !== "removed" && <div className="text-success">+ {JSON.stringify(c.after)}</div>}
                    </div>
                  ))}
                </div>
              </section>
            )}
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}
