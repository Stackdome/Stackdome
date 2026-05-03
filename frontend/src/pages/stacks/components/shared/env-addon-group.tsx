import { ExternalLink, Pencil, Undo2, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { AddonTypeIcon } from "@/pages/addons/components/addon-type-icon";
import { SoftLabel } from "@/pages/stacks/components/shared/section-label";
import {
  CRED_FIELDS,
  CLUSTER_WIDE_FIELDS,
  type CredField,
} from "@/pages/stacks/lib/addon-presets";
import { cn } from "@/lib/utils";
import { useState } from "react";

export interface EnvAddonBinding {
  envName: string;
  credField?: CredField;
}

export type EnvAddonGroupState = "idle" | "editing-binding" | "detaching";

interface EnvAddonGroupProps {
  addonId: string;
  addonName: string;
  bindings: EnvAddonBinding[];
  database?: string;
  superuser?: boolean;
  state?: EnvAddonGroupState;
  onEditBinding?: () => void;
  onDetach?: () => void;
  onCancelDetach?: () => void;
  onChangeBinding?: (oldCredField: CredField | undefined, newCredField: CredField) => void;
  onRemoveBinding?: (credField: CredField | undefined, envName: string) => void;
}

/**
 * Inline group displayed at the bottom of a resource's Environment tab when
 * the resource binds env vars from an addon. State machine:
 *
 *   idle ─ Edit binding ──▶ editing-binding
 *   editing-binding ─ Detach addon ──▶ detaching
 *   detaching ─ Cancel detach ──▶ editing-binding
 *   detaching ─ Deploy + save ──▶ detached (group unmounts; provenance line
 *                                  shown by parent on the converted env rows)
 */
export default function EnvAddonGroup({
  addonId,
  addonName,
  bindings,
  database,
  superuser,
  state = "idle",
  onEditBinding,
  onDetach,
  onCancelDetach,
  onChangeBinding,
  onRemoveBinding,
}: EnvAddonGroupProps) {
  const subline = `Postgres · ${bindings.length} ${bindings.length === 1 ? "key" : "keys"} bound${
    superuser ? " · superuser" : database ? ` · db: ${database}` : ""
  }`;

  const [openRebindFor, setOpenRebindFor] = useState<string | null>(null);
  const isDetaching = state === "detaching";
  const isEditing = state === "editing-binding";

  return (
    <div
      className={cn(
        "border border-border rounded-md p-3 space-y-2",
        isDetaching ? "bg-danger-bg" : "bg-background",
      )}
    >
      <div className="flex items-center gap-2">
        <AddonTypeIcon type="postgres" size={20} />
        <span className="text-sm font-semibold text-foreground">{addonName}</span>
        <SoftLabel>{subline}</SoftLabel>
        <div className="grow" />
        {state === "idle" && (
          <>
            {onEditBinding && (
              <Button
                variant="ghost"
                size="sm"
                className="h-7 px-2 text-[12.5px] text-muted-foreground"
                onClick={onEditBinding}
              >
                <Pencil className="h-3 w-3" />
                Edit binding
              </Button>
            )}
            <a
              href={`/addons/postgres/${addonId}`}
              target="_blank"
              rel="noreferrer"
              className="text-[12.5px] text-brand hover:text-brand-press inline-flex items-center gap-1"
            >
              View binding <ExternalLink className="h-3 w-3" />
            </a>
          </>
        )}
        {isEditing && onDetach && (
          <Button
            variant="ghost"
            size="sm"
            className="h-7 px-2 text-[12.5px] text-danger hover:bg-danger-bg hover:text-danger"
            onClick={onDetach}
          >
            Detach addon
          </Button>
        )}
        {isDetaching && onCancelDetach && (
          <Button
            variant="ghost"
            size="sm"
            className="h-7 px-2 text-[12.5px] text-muted-foreground"
            onClick={onCancelDetach}
          >
            Cancel detach
          </Button>
        )}
      </div>
      <div className="space-y-1">
        {bindings.map((b, i) => {
          const rowKey = `${b.envName}-${b.credField ?? "?"}-${i}`;
          const showRebind = isEditing && openRebindFor === rowKey;
          return (
            <div
              key={rowKey}
              className={cn(
                "flex items-center gap-2 px-2 py-1 rounded-sm",
                isDetaching ? "" : "hover:bg-muted/30",
              )}
            >
              <span
                className={cn(
                  "font-mono text-xs min-w-[10rem]",
                  isDetaching ? "line-through text-danger" : "text-foreground",
                )}
              >
                {b.envName}
              </span>
              <span className="text-muted-foreground text-xs">→</span>
              {showRebind ? (
                <Select
                  defaultValue={b.credField}
                  onValueChange={(v) => {
                    onChangeBinding?.(b.credField, v as CredField);
                    setOpenRebindFor(null);
                  }}
                >
                  <SelectTrigger className="h-7 w-[160px] font-mono text-xs">
                    <SelectValue placeholder="Field" />
                  </SelectTrigger>
                  <SelectContent>
                    {CRED_FIELDS.filter((f) => f !== b.credField).map((f) => (
                      <SelectItem key={f} value={f}>
                        <span className="flex items-center gap-2">
                          <span>{f}</span>
                          {CLUSTER_WIDE_FIELDS.has(f) && (
                            <span className="text-[10px] text-muted-foreground">cluster</span>
                          )}
                        </span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              ) : (
                <span
                  className={cn(
                    "inline-flex items-center gap-1 px-2 py-0.5 rounded-sm border border-border bg-muted/30 font-mono text-xs",
                    isDetaching && "line-through text-danger",
                  )}
                >
                  {b.credField ?? "—"}
                  {b.credField && CLUSTER_WIDE_FIELDS.has(b.credField) && (
                    <span className="text-[10px] text-muted-foreground">(cluster)</span>
                  )}
                </span>
              )}
              <div className="grow" />
              {!isDetaching && isEditing && onChangeBinding && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 hover:bg-muted/40"
                  onClick={() => setOpenRebindFor(showRebind ? null : rowKey)}
                  aria-label={`Re-bind ${b.envName}`}
                >
                  <Undo2 className="h-3.5 w-3.5" />
                </Button>
              )}
              {!isDetaching && (isEditing || state === "idle") && onRemoveBinding && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 hover:bg-danger-bg hover:text-danger"
                  onClick={() => onRemoveBinding(b.credField, b.envName)}
                  aria-label={`Remove binding for ${b.envName}`}
                >
                  <X className="h-3.5 w-3.5" />
                </Button>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
