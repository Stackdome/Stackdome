import { EllipsisVertical, RefreshCw, ShieldCheck, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

export function RowMenu({
  onVerify,
  onSync,
  onRemove,
}: {
  onVerify: () => void;
  onSync: () => void;
  onRemove: () => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" aria-label="Open row menu" className="h-[30px] w-[30px]">
          <EllipsisVertical className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[236px]">
        {/* Verify and Remove both open a dialog. Radix's DropdownMenu→Dialog
            composition races the menu's close (which resets
            document.body.style.pointerEvents) against the dialog's mount, and
            can leave pointer-events "none" on body forever if the dialog is
            cancelled. Deferring the callback until after the menu has fully
            closed avoids the race. See https://github.com/radix-ui/primitives/issues/1836 */}
        <DropdownMenuItem onSelect={() => setTimeout(() => onVerify(), 0)}>
          <ShieldCheck className="h-4 w-4" />
          Verify repository access
        </DropdownMenuItem>
        <DropdownMenuItem onSelect={() => onSync()}>
          <RefreshCw className="h-4 w-4" />
          <div className="flex flex-col">
            <span>Sync from GitHub</span>
            <span className="text-xs text-muted-foreground">Re-check access now</span>
          </div>
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem variant="destructive" onSelect={() => setTimeout(() => onRemove(), 0)}>
          <Trash2 className="h-4 w-4" />
          Remove integration
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
