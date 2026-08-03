import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import { AddonTypeIcon } from "./addon-type-icon";

export type AddonType = "postgres";

interface AddonTypePickerDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelect: (type: AddonType) => void;
}

interface AddonTypeOption {
  id: AddonType | "redis" | "ollama";
  name: string;
  description: string;
  available: boolean;
}

const OPTIONS: AddonTypeOption[] = [
  {
    id: "postgres",
    name: "PostgreSQL",
    description: "Managed Postgres database with automated backups and high availability.",
    available: true,
  },
  {
    id: "redis",
    name: "Redis",
    description: "In-memory data store for caching and pub/sub.",
    available: false,
  },
  {
    id: "ollama",
    name: "Ollama",
    description: "Self-hosted LLM runtime for embeddings and chat completions.",
    available: false,
  },
];

export function AddonTypePickerDialog({
  open,
  onOpenChange,
  onSelect,
}: AddonTypePickerDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[560px]">
        <DialogHeader>
          <DialogTitle>Add Addon</DialogTitle>
          <DialogDescription>
            Pick a service to provision for your stacks.
          </DialogDescription>
        </DialogHeader>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2">
          {OPTIONS.map((option) => (
            <button
              key={option.id}
              type="button"
              disabled={!option.available}
              onClick={() => option.available && onSelect(option.id as AddonType)}
              className={cn(
                "text-left rounded-md border border-border bg-card p-4 transition-colors focus-visible:outline-2 focus-visible:outline-[var(--ring)] focus-visible:outline-offset-2",
                option.available
                  ? "hover:bg-muted/40 hover:border-border-strong cursor-pointer"
                  : "opacity-60 cursor-not-allowed",
              )}
            >
              <div className="flex items-center gap-2 mb-2">
                <AddonTypeIcon type={option.id} size={20} />
                <span className="font-medium text-foreground">{option.name}</span>
                {!option.available && (
                  <span className="ml-auto font-mono text-[10.5px] uppercase tracking-[1px] text-muted-foreground">
                    Soon
                  </span>
                )}
              </div>
              <p className="text-[12.5px] text-muted-foreground leading-relaxed">
                {option.description}
              </p>
            </button>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}
