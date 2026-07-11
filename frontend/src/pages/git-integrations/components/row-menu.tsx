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
        <DropdownMenuItem onSelect={() => onVerify()}>
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
        <DropdownMenuItem variant="destructive" onSelect={() => onRemove()}>
          <Trash2 className="h-4 w-4" />
          Remove integration
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
