import { Globe, Database, Wrench, ArrowRight } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";

const PARTS = [
  { icon: Globe, label: "web", desc: "The page people visit." },
  { icon: Database, label: "redis", desc: "Passes jobs along." },
  { icon: Wrench, label: "worker", desc: "Does the jobs." },
] as const;

interface WelcomeDialogProps {
  open: boolean;
  /** User accepted — caller seeds the demo draft and starts the canvas tour. */
  onTakeTour: () => void;
  /** Closed with X or escape — just hidden, offered again next visit. */
  onClose: () => void;
  /** Explicit opt-out — never shown again. */
  onOptOut: () => void;
}

export function WelcomeDialog({ open, onTakeTour, onClose, onOptOut }: WelcomeDialogProps) {
  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && onClose()}>
      <DialogContent className="block gap-0 overflow-hidden p-0 sm:max-w-[560px]">
        <DialogTitle className="sr-only">Deploy your first stack</DialogTitle>
        <DialogDescription className="sr-only">
          Take a short guided tour and deploy a demo app
        </DialogDescription>

        <div className="px-8 pb-8 pt-10">
          <div className="mb-7 text-center">
            <h2 className="mb-2 text-head font-medium">
              Deploy your first stack
            </h2>
            <p className="text-body text-muted-foreground">
              We put together a small demo app for you.
              <br />
              Take a short tour, press deploy, and watch it go live on your cluster.
            </p>
          </div>

          <div className="mb-8 grid grid-cols-3 gap-2.5">
            {PARTS.map((p) => (
              <div
                key={p.label}
                className="flex flex-col items-center gap-2 rounded-md border bg-card p-4 text-center"
              >
                <span className="flex h-9 w-9 items-center justify-center rounded border border-brand-border bg-brand-bg text-brand">
                  <p.icon className="h-[18px] w-[18px]" />
                </span>
                <span className="text-body font-medium text-foreground">{p.label}</span>
                <span className="text-meta text-muted-foreground">{p.desc}</span>
              </div>
            ))}
          </div>

          <div className="flex flex-col items-center gap-2">
            <Button onClick={onTakeTour}>
              Take the tour
              <ArrowRight className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="sm" onClick={onOptOut}>
              I&apos;ll explore on my own
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
